package main

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"github.com/go-playground/validator/v10"
	"golang.org/x/sys/unix"

	"lite-nas/shared/fileaccess"
)

// readerInput is the validated CLI boundary for the Unix-user reader mode.
type readerInput struct {
	Path   string `validate:"required,max=4096"`
	Device uint64 `validate:"required"`
	Inode  uint64 `validate:"required"`
	Offset int64  `validate:"min=0"`
	Count  int64  `validate:"min=0"`
}

// parseReaderInput accepts only the fixed path and numeric operands sent by
// the parent service; count zero means stream through EOF.
func parseReaderInput(args []string) (readerInput, error) {
	if len(args) != 5 || fileaccess.ValidateHostPath(args[0]) != nil {
		return readerInput{}, fileaccess.ErrInvalidPath
	}
	device, deviceErr := strconv.ParseUint(args[1], 10, 64)
	inode, inodeErr := strconv.ParseUint(args[2], 10, 64)
	offset, offsetErr := strconv.ParseInt(args[3], 10, 64)
	count, countErr := strconv.ParseInt(args[4], 10, 64)
	input := readerInput{Path: args[0], Device: device, Inode: inode, Offset: offset, Count: count}
	if deviceErr != nil || inodeErr != nil || offsetErr != nil || countErr != nil || validator.New().Struct(input) != nil {
		return readerInput{}, fileaccess.ErrInvalidPath
	}
	return input, nil
}

// runUserReader opens the expected inode as the verified non-root Unix user.
// It writes only file bytes to stdout and never creates a temporary file.
func runUserReader(args []string) error {
	if os.Geteuid() == 0 {
		return errors.New("reader must not run as root")
	}
	input, err := parseReaderInput(args)
	if err != nil {
		return err
	}
	if err := requireLiveParent(); err != nil {
		return err
	}
	file, err := openExpectedFile(input.Path, input.Device, input.Inode)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.Seek(input.Offset, io.SeekStart); err != nil {
		return err
	}
	if input.Count == 0 {
		_, err = io.Copy(os.Stdout, file)
	} else {
		_, err = io.CopyN(os.Stdout, file, input.Count)
	}
	return err
}

// requireLiveParent kills a sudo-launched helper if its request process exits.
func requireLiveParent() error {
	parent := os.Getppid()
	if parent <= 1 {
		return errors.New("request parent is unavailable")
	}
	if err := unix.Prctl(unix.PR_SET_PDEATHSIG, uintptr(syscall.SIGKILL), 0, 0, 0); err != nil {
		return err
	}
	if os.Getppid() != parent {
		return errors.New("request parent exited before reader started")
	}
	return nil
}

// openExpectedFile uses the race-resistant opener and checks the parent-stated inode.
func openExpectedFile(path string, device, inode uint64) (*os.File, error) {
	before, err := os.Stat(path)
	if err != nil || !before.Mode().IsRegular() {
		return nil, errors.New("file is unavailable")
	}
	stat, ok := before.Sys().(*syscall.Stat_t)
	if !ok || uint64(stat.Dev) != device || stat.Ino != inode {
		return nil, errors.New("file identity changed")
	}
	return openAuthorizedFile(path, before)
}

// userReader starts a reader through the File Service style per-user sudo rule.
// The username is re-resolved from the host and matched to the RBAC UID.
func userReader(ctx context.Context, subject fileaccess.Subject, path string, info os.FileInfo, offset, count int64) (io.ReadCloser, error) {
	if fileaccess.ValidateHostPath(path) != nil || offset < 0 || count < 0 {
		return nil, fileaccess.ErrInvalidPath
	}
	username, err := fileaccess.VerifyUnixIdentity(subject)
	if err != nil {
		return nil, err
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || stat.Ino == 0 {
		return nil, errors.New("file identity unavailable")
	}
	command := exec.CommandContext(ctx, "/usr/bin/sudo", "-n", "-u", username, "--",
		"/usr/libexec/lite-nas/stream-service", "--read-file", path,
		strconv.FormatUint(uint64(stat.Dev), 10), strconv.FormatUint(stat.Ino, 10),
		strconv.FormatInt(offset, 10), strconv.FormatInt(count, 10))
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	output, err := command.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := command.Start(); err != nil {
		_ = output.Close()
		return nil, err
	}
	return &readerProcess{ReadCloser: output, command: command}, nil
}

// readerProcess owns and reaps one sudo reader and its pipe.
type readerProcess struct {
	io.ReadCloser
	command *exec.Cmd
	waited  bool
}

// Read surfaces a failed reader after it reaches EOF.
func (process *readerProcess) Read(data []byte) (int, error) {
	count, err := process.ReadCloser.Read(data)
	if err == io.EOF && !process.waited {
		process.waited = true
		if waitErr := process.command.Wait(); waitErr != nil {
			return count, waitErr
		}
	}
	return count, err
}

// Close cancels the reader if the client disconnected, then reaps it.
func (process *readerProcess) Close() error {
	err := process.ReadCloser.Close()
	if process.command.Process != nil && !process.waited {
		_ = syscall.Kill(-process.command.Process.Pid, syscall.SIGKILL)
		_ = process.command.Process.Kill()
		_ = process.command.Wait()
		process.waited = true
	}
	return err
}
