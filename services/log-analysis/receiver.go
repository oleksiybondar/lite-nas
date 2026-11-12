package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/user"
	"strconv"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"

	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/workers"
)

// listenFDsStart is the first file descriptor passed by systemd socket
// activation (SD_LISTEN_FDS_START).
const listenFDsStart = 3

// readRetryDelay bounds the loop rate when the socket reports a transient
// read error, so a persistent fault cannot spin the receiver.
const readRetryDelay = 100 * time.Millisecond

// activatedDatagramSocket adopts the single Unix datagram socket passed by
// lite-nas-log-analysis.socket using the systemd LISTEN_FDS protocol. The
// socket stays owned by systemd, so records queued while the service
// restarts remain available to the next process.
func activatedDatagramSocket() (*net.UnixConn, error) {
	if os.Getenv("LISTEN_PID") != strconv.Itoa(os.Getpid()) || os.Getenv("LISTEN_FDS") != "1" {
		return nil, errors.New("log analysis requires exactly one socket from lite-nas-log-analysis.socket")
	}
	file := os.NewFile(listenFDsStart, "lite-nas-log-analysis.socket")
	defer file.Close()
	conn, err := net.FilePacketConn(file)
	if err != nil {
		return nil, err
	}
	unixConn, ok := conn.(*net.UnixConn)
	if !ok {
		_ = conn.Close()
		return nil, errors.New("activated log analysis socket is not a Unix datagram socket")
	}
	return unixConn, nil
}

// resolveSenderUIDs maps configured rsyslog account names to numeric
// identities. Missing accounts are skipped because rsyslog runs as syslog on
// Ubuntu and as root on Debian; at least one configured account must exist.
func resolveSenderUIDs(names []string) (map[uint32]bool, error) {
	senders := make(map[uint32]bool)
	for _, name := range names {
		account, err := user.Lookup(name)
		if err != nil {
			continue
		}
		uid, err := strconv.ParseUint(account.Uid, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("sender account %s has invalid uid: %w", name, err)
		}
		senders[uint32(uid)] = true
	}
	if len(senders) == 0 {
		return nil, errors.New("no configured rsyslog sender account exists")
	}
	return senders, nil
}

// rsyslogReceiver is the single producer of normalized log records. It owns
// the inherited datagram connection, accepts datagrams only from allowed
// rsyslog identities, validates the JSON contract and sends one record copy
// per selecting source into the dispatcher channel. It closes no shared
// channel; the runtime closes the output after Done.
type rsyslogReceiver struct {
	conn     *net.UnixConn
	sources  []sourceSpec
	senders  map[uint32]bool
	validate *validator.Validate
	output   chan<- logRecord
	worker   workers.BaseWorker
	done     chan struct{}
	log      sharedlogger.Logger
}

// newRsyslogReceiver enables kernel sender credentials on the socket. The
// socket unit also sets PassCredentials so datagrams queued before startup
// carry credentials; enabling it here keeps the receiver correct on its own.
func newRsyslogReceiver(conn *net.UnixConn, sources []sourceSpec, senders map[uint32]bool, output chan<- logRecord, log sharedlogger.Logger) (*rsyslogReceiver, error) {
	raw, err := conn.SyscallConn()
	if err != nil {
		return nil, err
	}
	var optErr error
	if err := raw.Control(func(fd uintptr) {
		optErr = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_PASSCRED, 1)
	}); err != nil {
		return nil, err
	}
	if optErr != nil {
		return nil, optErr
	}
	receiver := &rsyslogReceiver{
		conn: conn, sources: sources, senders: senders,
		validate: validator.New(validator.WithRequiredStructEnabled()),
		output:   output, done: make(chan struct{}), log: log,
	}
	receiver.worker = workers.NewBaseWorker(receiver.run)
	return receiver, nil
}

// Start launches the datagram reader under the service context.
func (r *rsyslogReceiver) Start(ctx context.Context) { r.worker.Start(ctx) }

// Done closes after the last record send; the runtime then closes dispatch input.
func (r *rsyslogReceiver) Done() <-chan struct{} { return r.done }

func (r *rsyslogReceiver) run(ctx context.Context) {
	defer close(r.done)
	defer r.conn.Close()
	// Closing this process's descriptor unblocks the read on shutdown;
	// systemd keeps its own copy of the socket open.
	stop := context.AfterFunc(ctx, func() { _ = r.conn.Close() })
	defer stop()
	data := make([]byte, maxDatagramBytes)
	oob := make([]byte, syscall.CmsgSpace(syscall.SizeofUcred))
	for {
		count, oobCount, flags, _, err := r.conn.ReadMsgUnix(data, oob)
		if err != nil {
			if ctx.Err() != nil || errors.Is(err, net.ErrClosed) {
				return
			}
			r.log.Error("rsyslog datagram read failed", "error", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(readRetryDelay):
			}
			continue
		}
		records, err := r.accept(data[:count], oob[:oobCount], flags)
		if err != nil {
			r.log.Warn("rejected rsyslog datagram", "error", err)
			continue
		}
		for _, record := range records {
			select {
			case r.output <- record:
			case <-ctx.Done():
				return
			}
		}
	}
}

// accept verifies the sender identity before decoding, then returns one
// normalized record per selecting source. Unselected records yield none.
func (r *rsyslogReceiver) accept(data, oob []byte, flags int) ([]logRecord, error) {
	if flags&(syscall.MSG_TRUNC|syscall.MSG_CTRUNC) != 0 {
		return nil, errors.New("truncated datagram")
	}
	uid, err := senderUID(oob)
	if err != nil {
		return nil, err
	}
	if !r.senders[uid] {
		return nil, fmt.Errorf("sender uid %d is not an allowed rsyslog account", uid)
	}
	record, at, err := decodeRsyslogRecord(data, r.validate)
	if err != nil {
		return nil, err
	}
	var records []logRecord
	for _, source := range r.sources {
		if !source.selects(record) {
			continue
		}
		normalized, err := newLogRecord(source, record, at)
		if err != nil {
			return nil, err
		}
		records = append(records, normalized)
	}
	return records, nil
}

// senderUID reads the kernel-attached SCM_CREDENTIALS of one datagram. The
// kernel fills these for SO_PASSCRED receivers, so a sender cannot forge them.
func senderUID(oob []byte) (uint32, error) {
	messages, err := syscall.ParseSocketControlMessage(oob)
	if err != nil {
		return 0, err
	}
	for i := range messages {
		if messages[i].Header.Level != syscall.SOL_SOCKET || messages[i].Header.Type != syscall.SCM_CREDENTIALS {
			continue
		}
		credentials, err := syscall.ParseUnixCredentials(&messages[i])
		if err != nil {
			return 0, err
		}
		return credentials.Uid, nil
	}
	return 0, errors.New("datagram has no sender credentials")
}
