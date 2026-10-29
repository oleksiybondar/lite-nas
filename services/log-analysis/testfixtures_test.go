package main

import (
	"net"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"

	sharedlogger "lite-nas/shared/logger"
)

const (
	authDatagram   = `{"timestamp":"2026-10-25T08:26:06.830681+01:00", "facility":"authpriv", "input":"imuxsock", "program":"sshd", "message":" pam_unix(sshd:auth): authentication failure; logname= uid=0 rhost=192.0.2.4 user=alice"}` + "\n"
	kernelDatagram = `{"timestamp":"2026-10-25T06:28:55.132359+01:00", "facility":"kern", "input":"imklog", "program":"kernel", "message":"usb 3-5: New USB device found, idVendor=203a, idProduct=fff9, bcdDevice= 1.00"}` + "\n"
)

// testSources mirrors the packaged rules.json source selection.
var testSources = []sourceSpec{
	{ID: "auth", Format: formatAuth, Facilities: []string{"auth", "authpriv"}},
	{ID: "kernel", Format: formatKernel, Inputs: []string{"imklog"}},
}

// testValidator returns the receiver's record validator configuration.
func testValidator() *validator.Validate {
	return validator.New(validator.WithRequiredStructEnabled())
}

// mustDecodeRecord decodes a datagram that the test expects to be valid.
func mustDecodeRecord(t *testing.T, datagram string) (rsyslogRecord, time.Time) {
	t.Helper()
	record, at, err := decodeRsyslogRecord([]byte(datagram), testValidator())
	if err != nil {
		t.Fatalf("decode rsyslog record: %v", err)
	}
	return record, at
}

// mustNormalize decodes a valid datagram and normalizes it for one source.
func mustNormalize(t *testing.T, source sourceSpec, datagram string) logRecord {
	t.Helper()
	record, at := mustDecodeRecord(t, datagram)
	normalized, err := newLogRecord(source, record, at)
	if err != nil {
		t.Fatalf("normalize record: %v", err)
	}
	return normalized
}

// mustReceiver builds a receiver on a bound temporary datagram socket. The
// receiver accepts the given sender UIDs and writes to a buffered output.
func mustReceiver(t *testing.T, senders map[uint32]bool) (*rsyslogReceiver, *net.UnixConn, chan logRecord) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "rsyslog.sock")
	server, err := net.ListenUnixgram("unixgram", &net.UnixAddr{Name: path, Net: "unixgram"})
	if err != nil {
		t.Fatalf("listen unixgram: %v", err)
	}
	output := make(chan logRecord, 4)
	receiver, err := newRsyslogReceiver(server, testSources, senders, output, sharedlogger.NewNop())
	if err != nil {
		_ = server.Close()
		t.Fatalf("new receiver: %v", err)
	}
	client, err := net.DialUnix("unixgram", nil, &net.UnixAddr{Name: path, Net: "unixgram"})
	if err != nil {
		_ = server.Close()
		t.Fatalf("dial unixgram: %v", err)
	}
	t.Cleanup(func() { _ = client.Close(); _ = server.Close() })
	return receiver, client, output
}

// currentSender allows datagrams sent by the test process identity.
func currentSender() map[uint32]bool {
	return map[uint32]bool{uint32(os.Getuid()): true}
}

// credentialsFor encodes the SCM_CREDENTIALS control message the kernel
// attaches for a sender with the given UID.
func credentialsFor(uid uint32) []byte {
	return syscall.UnixCredentials(&syscall.Ucred{Pid: int32(os.Getpid()), Uid: uid, Gid: uint32(os.Getgid())})
}
