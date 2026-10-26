package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"lite-nas/shared/mountinfo"
)

const (
	clamscanBinary = "/usr/bin/clamscan"
	maxScanOutput  = 256 * 1024
)

// scanStatus distinguishes a complete clean scan, malware, an incomplete
// scan, and cancellation. Only a complete exit status 0 is called clean.
type scanStatus string

const (
	scanClean      scanStatus = "clean"
	scanInfected   scanStatus = "infected"
	scanIncomplete scanStatus = "incomplete"
	scanCancelled  scanStatus = "cancelled"
)

// scanResult carries bounded scanner evidence for operational reporting.
type scanResult struct {
	mount    mountinfo.Record
	eventID  string
	status   scanStatus
	evidence string
	err      error
}

// boundedOutput retains a limited prefix and records truncation without
// blocking the scanner process when its output exceeds the limit.
type boundedOutput struct {
	data      []byte
	truncated bool
}

func (out *boundedOutput) Write(chunk []byte) (int, error) {
	remaining := maxScanOutput - len(out.data)
	if remaining > 0 {
		out.data = append(out.data, chunk[:min(remaining, len(chunk))]...)
	}
	if len(chunk) > remaining {
		out.truncated = true
	}
	return len(chunk), nil
}

// liveMount verifies the mount ID and all relevant metadata immediately
// before a scan. The mount may still change while clamscan traverses it.
func liveMount(expected mountinfo.Record) error {
	current, err := mountinfo.ReadCurrent()
	if err != nil {
		return err
	}
	actual, exists := current[expected.ID]
	if !exists || actual != expected {
		return errors.New("mount identity changed or disappeared")
	}
	resolved, err := filepath.EvalSymlinks(expected.Point)
	if err != nil {
		return err
	}
	if filepath.Clean(resolved) != filepath.Clean(expected.Point) {
		return errors.New("mount point contains a symlink")
	}
	info, err := os.Stat(expected.Point)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New("mount point is not a directory")
	}
	return nil
}

// scanMount uses only fixed ClamAV options, an end-of-options delimiter, and
// a validated absolute mount path. It never invokes a shell or destructive
// ClamAV flags. The caller's context cancels on unmount or shutdown.
func scanMount(ctx context.Context, mount mountinfo.Record, timeout time.Duration) scanResult {
	result := scanResult{mount: mount}
	if err := liveMount(mount); err != nil {
		result.status = scanIncomplete
		result.err = err
		return result
	}
	if timeout <= 0 {
		result.status = scanIncomplete
		result.err = errors.New("scan timeout must be positive")
		return result
	}
	commandCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	command := exec.CommandContext(commandCtx, clamscanBinary,
		"--recursive=yes", "--cross-fs=no", "--follow-dir-symlinks=0", "--follow-file-symlinks=0",
		"--infected", "--no-summary", "--", mount.Point)
	command.Env = []string{"PATH=/usr/bin:/bin", "LANG=C"}
	var output boundedOutput
	command.Stdout = &output
	command.Stderr = &output
	err := command.Run()
	result.evidence = strings.TrimSpace(string(output.data))
	if ctx.Err() != nil {
		result.status = scanCancelled
		result.err = ctx.Err()
		return result
	}
	if commandCtx.Err() != nil {
		result.status = scanIncomplete
		result.err = commandCtx.Err()
		return result
	}
	if output.truncated {
		result.status = scanIncomplete
		result.err = errors.New("scanner output exceeded limit")
		return result
	}
	if identityErr := liveMount(mount); identityErr != nil {
		result.status = scanIncomplete
		result.err = identityErr
		return result
	}
	if err == nil {
		result.status = scanClean
		return result
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) && exit.ExitCode() == 1 {
		result.status = scanInfected
		return result
	}
	result.status = scanIncomplete
	result.err = fmt.Errorf("clamscan: %w", err)
	return result
}
