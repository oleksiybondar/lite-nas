package workers

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"lite-nas/shared/metrics"
)

// TestValidateZpoolPathRejectsUnpackagedPaths covers configured binary validation.
func TestValidateZpoolPathRejectsUnpackagedPaths(t *testing.T) {
	for _, path := range []string{"", "zpool", "/tmp/zpool", "/usr/bin/not-zpool"} {
		t.Run(path, func(t *testing.T) {
			if _, err := validateZpoolPath(path); err == nil {
				t.Fatalf("validateZpoolPath(%q) accepted invalid executable", path)
			}
		})
	}
}

// TestValidateZpoolPathAcceptsInstalledBinary covers the packaged binary when present.
func TestValidateZpoolPathAcceptsInstalledBinary(t *testing.T) {
	if _, err := os.Stat("/usr/sbin/zpool"); os.IsNotExist(err) {
		t.Skipf("installed zpool binary unavailable: %v", err)
	}
	if _, err := validateZpoolPath("/usr/sbin/zpool"); err != nil {
		t.Fatalf("validateZpoolPath() rejected installed binary: %v", err)
	}
}

// TestValidatePathRejectsDirectoriesAndMissingFiles covers non-file operands.
func TestValidatePathRejectsDirectoriesAndMissingFiles(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "zpool")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{directory, filepath.Join(directory, "zpool")} {
		t.Run(path, func(t *testing.T) {
			if _, err := validatePath(path); err == nil {
				t.Fatalf("validatePath(%q) accepted non-file", path)
			}
		})
	}
}

// TestValidatePathAcceptsRegularFile covers executable file shape.
func TestValidatePathAcceptsRegularFile(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "zpool")
	if err := os.WriteFile(binary, []byte("test executable"), 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := validatePath(binary); err != nil {
		t.Fatalf("validatePath() rejected regular file: %v", err)
	}
}

// TestValidateZpoolPathRejectsOutsideAllowlist covers binary location policy.
func TestValidateZpoolPathRejectsOutsideAllowlist(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "zpool")
	if err := os.WriteFile(binary, []byte("test executable"), 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := validateZpoolPath(binary); err == nil {
		t.Fatal("validateZpoolPath() accepted binary outside the packaged allowlist")
	}
}

// TestValidateAllowedZpoolPathAcceptsPackagedBinary covers the fixed path policy.
func TestValidateAllowedZpoolPathAcceptsPackagedBinary(t *testing.T) {
	if err := validateAllowedZpoolPath("/usr/sbin/zpool"); err != nil {
		t.Fatalf("validateAllowedZpoolPath() rejected packaged path: %v", err)
	}
}

// mustBuildSnapshot assembles the valid parser fixture for focused result checks.
func mustBuildSnapshot(t *testing.T) metrics.ZFSSnapshot {
	t.Helper()
	worker := PollingWorker{}
	snapshot, err := worker.buildSnapshot(validPollInput())
	if err != nil {
		t.Fatalf("buildSnapshot() error = %v", err)
	}
	return snapshot
}

// TestBuildSnapshotParsesOnePool covers the command output composition path.
func TestBuildSnapshotParsesOnePool(t *testing.T) {
	snapshot := mustBuildSnapshot(t)
	if len(snapshot.Pools) != 1 {
		t.Fatalf("buildSnapshot() pool count = %d", len(snapshot.Pools))
	}
	if snapshot.Pools[0].Name != "LiteNAS" {
		t.Fatalf("buildSnapshot() pools = %+v", snapshot.Pools)
	}
}

// TestBuildSnapshotRejectsMalformedOutputs covers failures from each external parser.
func TestBuildSnapshotRejectsMalformedOutputs(t *testing.T) {
	for _, mutate := range []func(*pollInput){
		func(input *pollInput) { input.statusOut = "invalid" },
		func(input *pollInput) { input.listOut = "invalid" },
		func(input *pollInput) { input.iostatOut = "invalid" },
	} {
		input := validPollInput()
		mutate(&input)
		if _, err := (PollingWorker{}).buildSnapshot(input); err == nil {
			t.Error("buildSnapshot() accepted malformed command output")
		}
	}
}

// TestPollingWorkerReportsInvalidBinary covers the worker's error channel and shutdown.
func TestPollingWorkerReportsInvalidBinary(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	ticks := make(chan struct{}, 1)
	snapshots := make(chan metrics.ZFSSnapshot, 1)
	errors := make(chan error, 1)
	worker := NewPollingWorker("/tmp/not-zpool", ticks, snapshots, errors)
	worker.Start(ctx)
	ticks <- struct{}{}
	select {
	case err := <-errors:
		if err == nil {
			t.Fatal("worker emitted nil error")
		}
	case <-time.After(time.Second):
		t.Fatal("worker did not report invalid executable")
	}
	cancel()
	close(ticks)
}

// TestWaitNextPollStopsOnClosedChannel covers worker shutdown without a tick.
func TestWaitNextPollStopsOnClosedChannel(t *testing.T) {
	ticks := make(chan struct{})
	close(ticks)
	if (PollingWorker{ticks: ticks}).waitNextPoll(t.Context()) {
		t.Fatal("closed tick channel should stop worker")
	}
}

// TestWaitNextPollStopsOnCancellation covers context-driven worker shutdown.
func TestWaitNextPollStopsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if (PollingWorker{ticks: make(chan struct{})}).waitNextPoll(ctx) {
		t.Fatal("canceled context should stop worker")
	}
}

// TestPollAndSendDiscardsErrorsAfterCancellation covers canceled poll delivery.
func TestPollAndSendDiscardsErrorsAfterCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	worker := PollingWorker{zpoolPath: "/tmp/not-zpool", errors: make(chan error)}
	worker.pollAndSend(ctx)
}

// TestPollAndSendDoesNotBlockWhenErrorChannelIsFull covers bounded error delivery.
func TestPollAndSendDoesNotBlockWhenErrorChannelIsFull(t *testing.T) {
	errorChannel := make(chan error, 1)
	errorChannel <- errors.New("earlier poll failed")
	worker := PollingWorker{zpoolPath: "/tmp/not-zpool", errors: errorChannel}
	worker.pollAndSend(t.Context())
	if len(errorChannel) != 1 {
		t.Fatalf("error channel length = %d", len(errorChannel))
	}
}

// validPollInput contains one consistent pool across the three command outputs.
func validPollInput() pollInput {
	return pollInput{
		statusOut: `pool: LiteNAS
 state: ONLINE
  scan: none requested
config:

	NAME      STATE     READ WRITE CKSUM
	LiteNAS   ONLINE       0     0     0

errors: No known data errors
`,
		listOut:   "LiteNAS\t1000\t300\t700\t30\tONLINE\n",
		iostatOut: "LiteNAS\t300\t700\t1\t2\t3\t4\n",
	}
}
