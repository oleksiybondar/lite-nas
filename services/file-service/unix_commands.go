package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"

	"lite-nas/services/file-service/filesystem"
	"lite-nas/shared/fileaccess"
)

const writerBinary = "/usr/libexec/lite-nas/lite-nas-file-writer"

// writerResult mirrors the installed file-writer's bounded JSON result.
type writerResult struct {
	Operation    string `json:"operation"`
	Device       uint64 `json:"device"`
	Inode        uint64 `json:"inode"`
	UID          uint32 `json:"uid"`
	Size         int64  `json:"size"`
	BytesWritten int64  `json:"bytes_written"`
	NextOffset   int64  `json:"next_offset"`
	SHA256       string `json:"sha256"`
}

// invokeWriter runs one fixed operation as the verified non-root Unix user.
func invokeWriter(ctx context.Context, subject fileaccess.Subject, operation string, record *uploadRecord, data []byte) (writerResult, error) {
	username, err := filesystem.VerifyUnixIdentity(subject)
	if err != nil {
		return writerResult{}, err
	}
	args := []string{
		"-n", "-u", username, "--", writerBinary, operation,
		"--target", record.Target + ".upload",
	}
	if operation != "create" {
		args = append(args,
			"--expected-device", strconv.FormatUint(record.StageDevice, 10),
			"--expected-inode", strconv.FormatUint(record.StageInode, 10),
			"--expected-uid", strconv.FormatUint(uint64(record.StageUID), 10))
	}
	switch operation {
	case "write":
		args = append(args, "--offset", strconv.FormatInt(record.Offset, 10),
			"--length", strconv.Itoa(len(data)))
	case "truncate":
		args = append(args, "--offset", strconv.FormatInt(record.Offset, 10))
	case "publish":
		if record.DestinationExists {
			args = append(args, "--destination-device", strconv.FormatUint(record.DestinationDevice, 10),
				"--destination-inode", strconv.FormatUint(record.DestinationInode, 10))
		} else {
			args = append(args, "--destination-absent")
		}
	}
	command := exec.CommandContext(ctx, "/usr/bin/sudo", args...)
	command.Stdin = bytes.NewReader(data)
	output, err := command.Output()
	if err != nil {
		return writerResult{}, fmt.Errorf("writer %s failed: %w", operation, err)
	}
	var result writerResult
	if len(output) > 4096 || json.Unmarshal(output, &result) != nil || result.Operation != operation {
		return writerResult{}, errors.New("writer returned invalid result")
	}
	return result, nil
}

// openCat streams one authorized host path through the user's Unix permissions.
// The accepted pathname race is recorded in READ_RACE_REVIEW.md.
func openCat(ctx context.Context, subject fileaccess.Subject, path string) (io.ReadCloser, error) {
	username, err := filesystem.VerifyUnixIdentity(subject)
	if err != nil {
		return nil, err
	}
	command := exec.CommandContext(ctx, "/usr/bin/sudo", "-n", "-u", username,
		"--", "/bin/cat", "--", path)
	command.Stderr = os.Stderr
	stdout, err := command.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := command.Start(); err != nil {
		stdout.Close()
		return nil, err
	}
	return &catProcess{ReadCloser: stdout, command: command}, nil
}

// catProcess stops a request-scoped cat process when the HTTP request ends.
type catProcess struct {
	io.ReadCloser
	command *exec.Cmd
	waited  bool
}

// Read reports a cat process failure when its output closes without data.
func (process *catProcess) Read(data []byte) (int, error) {
	count, err := process.ReadCloser.Read(data)
	if err == io.EOF && !process.waited {
		process.waited = true
		if waitErr := process.command.Wait(); waitErr != nil {
			return count, waitErr
		}
	}
	return count, err
}

// Close closes the pipe, stops cat, and reaps its process.
func (process *catProcess) Close() error {
	err := process.ReadCloser.Close()
	if process.command.Process != nil && !process.waited {
		_ = process.command.Process.Kill()
		_ = process.command.Wait()
		process.waited = true
	}
	return err
}
