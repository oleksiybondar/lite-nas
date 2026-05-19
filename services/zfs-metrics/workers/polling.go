package workers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"lite-nas/shared/metrics"
	iostatparser "lite-nas/shared/parsers/zfs/iostat"
	listparser "lite-nas/shared/parsers/zfs/list"
	snapshotcomposer "lite-nas/shared/parsers/zfs/snapshot"
	statusparser "lite-nas/shared/parsers/zfs/status"
)

var (
	// zpoolStatusArgs fetches topology and error breakdown for each pool.
	zpoolStatusArgs = []string{"status", "-P", "-L"}
	// zpoolListArgs fetches capacity/health values in stable numeric format.
	zpoolListArgs = []string{"list", "-H", "-p", "-o", strings.Join(listparser.DefaultHeaders, ",")}
	// zpoolIostatArgs fetches per-pool operations and bandwidth counters.
	zpoolIostatArgs = []string{"iostat", "-H", "-p"}

	// zpoolCommandTimeout bounds one zpool subprocess execution.
	zpoolCommandTimeout = 15 * time.Second
)

// PollingWorker periodically executes zpool commands and emits normalized ZFS snapshots.
type PollingWorker struct {
	zpoolPath string
	ticks     <-chan struct{}
	output    chan<- metrics.ZFSSnapshot
	errors    chan<- error
}

// pollInput groups raw command outputs for one polling cycle.
type pollInput struct {
	statusOut string
	listOut   string
	iostatOut string
}

// NewPollingWorker creates a PollingWorker.
func NewPollingWorker(
	zpoolPath string,
	ticks <-chan struct{},
	output chan<- metrics.ZFSSnapshot,
	errors chan<- error,
) PollingWorker {
	return PollingWorker{
		zpoolPath: zpoolPath,
		ticks:     ticks,
		output:    output,
		errors:    errors,
	}
}

// Start launches the worker loop in a dedicated goroutine.
func (w PollingWorker) Start(ctx context.Context) {
	go w.run(ctx)
}

// run drives the polling loop until context cancelation or channel closure.
func (w PollingWorker) run(ctx context.Context) {
	for {
		if !w.waitNextPoll(ctx) {
			return
		}
		w.pollAndSend(ctx)
	}
}

// waitNextPoll blocks until next tick or context cancellation.
func (w PollingWorker) waitNextPoll(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	case _, ok := <-w.ticks:
		return ok
	}
}

// pollAndSend executes one polling cycle and emits output/error non-blockingly.
func (w PollingWorker) pollAndSend(ctx context.Context) {
	snapshot, err := w.poll(ctx)
	if err != nil {
		select {
		case <-ctx.Done():
			return
		case w.errors <- err:
		default:
		}
		return
	}

	select {
	case <-ctx.Done():
		return
	case w.output <- snapshot:
	}
}

// poll collects raw command output and composes one normalized snapshot.
func (w PollingWorker) poll(ctx context.Context) (metrics.ZFSSnapshot, error) {
	input, err := w.collectPollInput(ctx)
	if err != nil {
		return metrics.ZFSSnapshot{}, err
	}

	return w.buildSnapshot(input)
}

// collectPollInput runs zpool commands required for one snapshot cycle.
func (w PollingWorker) collectPollInput(ctx context.Context) (pollInput, error) {
	statusOut, err := runZpoolCommand(ctx, w.zpoolPath, zpoolStatusArgs...)
	if err != nil {
		return pollInput{}, err
	}
	listOut, err := runZpoolCommand(ctx, w.zpoolPath, zpoolListArgs...)
	if err != nil {
		return pollInput{}, err
	}
	iostatOut, err := runZpoolCommand(ctx, w.zpoolPath, zpoolIostatArgs...)
	if err != nil {
		return pollInput{}, err
	}

	return pollInput{
		statusOut: statusOut,
		listOut:   listOut,
		iostatOut: iostatOut,
	}, nil
}

// buildSnapshot parses raw command outputs and composes the normalized snapshot.
func (w PollingWorker) buildSnapshot(input pollInput) (metrics.ZFSSnapshot, error) {
	statusDoc, _, err := statusparser.ParseZpoolStatus(input.statusOut, statusparser.ParseModeStrict)
	if err != nil {
		return metrics.ZFSSnapshot{}, err
	}
	usageByPool, err := listparser.Parse(input.listOut)
	if err != nil {
		return metrics.ZFSSnapshot{}, err
	}
	ioByPool, err := iostatparser.Parse(input.iostatOut)
	if err != nil {
		return metrics.ZFSSnapshot{}, err
	}

	return snapshotcomposer.Compose(time.Now(), statusDoc, usageByPool, ioByPool), nil
}

// runZpoolCommand executes one validated zpool command and returns stdout text.
func runZpoolCommand(ctx context.Context, zpoolPath string, args ...string) (string, error) {
	resolvedPath, err := validateZpoolPath(zpoolPath)
	if err != nil {
		return "", err
	}

	cmdCtx, cancel := context.WithTimeout(ctx, zpoolCommandTimeout)
	defer cancel()

	// #nosec G204 -- resolvedPath is strictly validated by validateZpoolPath.
	cmd := exec.CommandContext(cmdCtx, resolvedPath, args...)
	output, err := cmd.Output()
	if err != nil {
		if cmdCtx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("zpool command timed out after %s: %s %s", zpoolCommandTimeout, resolvedPath, strings.Join(args, " "))
		}
		return "", fmt.Errorf("zpool command failed: %s %s: %w", resolvedPath, strings.Join(args, " "), err)
	}
	return string(output), nil
}

// validateZpoolPath validates path shape and allowlist policy for zpool binary.
func validateZpoolPath(path string) (string, error) {
	cleanPath, err := validatePath(path)
	if err != nil {
		return "", err
	}

	if err := validateAllowedZpoolPath(cleanPath); err != nil {
		return "", err
	}

	return cleanPath, nil
}

// validatePath enforces generic executable path requirements for zpool.
func validatePath(path string) (string, error) {
	if path == "" {
		return "", errors.New("zpool path must not be empty")
	}
	if !filepath.IsAbs(path) {
		return "", fmt.Errorf("zpool path must be absolute: %q", path)
	}

	cleanPath := filepath.Clean(path)
	if filepath.Base(cleanPath) != "zpool" {
		return "", fmt.Errorf("zpool path must point to zpool binary: %q", cleanPath)
	}

	if err := validatePathAccessible(cleanPath); err != nil {
		return "", err
	}
	return cleanPath, nil
}

// validatePathAccessible checks that path exists and resolves to a file.
func validatePathAccessible(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("zpool path is not accessible: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("zpool path is a directory: %q", path)
	}
	return nil
}

// validateAllowedZpoolPath enforces service-specific zpool binary allowlist.
func validateAllowedZpoolPath(path string) error {
	allowedPaths := map[string]struct{}{
		"/sbin/zpool":     {},
		"/usr/sbin/zpool": {},
		"/usr/bin/zpool":  {},
	}
	if _, ok := allowedPaths[path]; !ok {
		return fmt.Errorf("zpool path is not allowed: %q", path)
	}

	return nil
}
