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

// validateWriterArguments restricts the writer invocation to supported operations and paths.
func validateWriterArguments(operation string, record *uploadRecord, data []byte) error {
	switch operation {
	case "create", "write", "stat", "checksum", "truncate", "publish", "cleanup":
	default:
		return errors.New("unsupported writer operation")
	}
	if record == nil || fileaccess.ValidateHostPath(record.Target) != nil ||
		fileaccess.ValidateHostPath(record.Target+".upload") != nil {
		return fileaccess.ErrInvalidPath
	}
	if record.PublicAvatar && (record.Private || sharedAvatarOwner(record.Target) != record.Owner) {
		return errors.New("public mode requires the owner's fixed avatar path")
	}
	if (operation == "write" || operation == "truncate") && record.Offset < 0 {
		return errors.New("invalid writer offset")
	}
	if len(data) > 16*1024*1024 {
		return errors.New("writer input exceeds maximum length")
	}
	return nil
}

// validateUserCommandArguments restricts user-scoped operations to fixed binaries and host paths.
func validateUserCommandArguments(binary string, paths []string) error {
	switch binary {
	case "/bin/mkdir", "/bin/rm", "/bin/rmdir":
	default:
		return errors.New("unsupported user command")
	}
	if binary == "/bin/mkdir" {
		if len(paths) != 2 || (paths[0] != "0770" && paths[0] != "0771" && paths[0] != "0775") {
			return errors.New("mkdir requires an allowed mode and one path")
		}
		return fileaccess.ValidateHostPath(paths[1])
	}
	if len(paths) != 1 {
		return errors.New("user command requires one path")
	}
	return fileaccess.ValidateHostPath(paths[0])
}

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
	if err := validateWriterArguments(operation, record, data); err != nil {
		return writerResult{}, err
	}
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
		mode := "0664"
		if record.Private {
			mode = "0660"
		}
		if record.PublicAvatar {
			mode = "0644"
		}
		args = append(args, "--mode", mode)
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
// The accepted pathname race is documented in requirements.md.
func openCat(ctx context.Context, subject fileaccess.Subject, path string) (io.ReadCloser, error) {
	if err := fileaccess.ValidateHostPath(path); err != nil {
		return nil, err
	}
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
