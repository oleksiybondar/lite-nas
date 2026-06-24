package workers

import (
	"context"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"lite-nas/shared/metrics"
	"lite-nas/shared/testutil/fstest"
)

// Requirements: process-metrics-svc/FR-001, process-metrics-svc/FR-002, process-metrics-svc/FR-003
func TestPollingWorkerPollBuildsLiveProcessSnapshot(t *testing.T) {
	t.Parallel()

	fixture := newPollingFixture(t)
	snapshot := mustPollFixtureSnapshot(t, fixture)

	if len(snapshot.Processes) != 6 {
		t.Fatalf("len(Processes) = %d, want 6", len(snapshot.Processes))
	}

	alpha := findProcessMetric(t, snapshot.Processes, 101)
	assertAlphaProcessMetric(t, alpha)
}

// Requirements: process-metrics-svc/FR-002
func TestParseProcessStatSupportsSpacesAndParenthesesInNames(t *testing.T) {
	t.Parallel()

	sample, err := parseProcessStat("202 (worker (alpha beta)) S 100 0 0 0 0 0 0 0 0 0 4 5 0 0 0 0 8 0 250 16384 7")
	if err != nil {
		t.Fatalf("parseProcessStat() error = %v", err)
	}

	if sample.Name != "worker (alpha beta)" {
		t.Fatalf("Name = %q, want %q", sample.Name, "worker (alpha beta)")
	}
	if sample.PPID != 100 {
		t.Fatalf("PPID = %d, want 100", sample.PPID)
	}
}

// Requirements: process-metrics-svc/FR-003
func TestPollingWorkerPollKeepsKernelThreadsWithEmptyCmdline(t *testing.T) {
	t.Parallel()

	fixture := newPollingFixture(t)
	snapshot := mustPollFixtureSnapshot(t, fixture)
	kthread := findProcessMetric(t, snapshot.Processes, 106)

	if kthread.Cmdline != "" {
		t.Fatalf("Cmdline = %q, want empty", kthread.Cmdline)
	}
	if kthread.Exe != "" {
		t.Fatalf("Exe = %q, want empty", kthread.Exe)
	}
	if kthread.Cwd != "" {
		t.Fatalf("Cwd = %q, want empty", kthread.Cwd)
	}
}

// Requirements: process-metrics-svc/FR-003
func TestPollingWorkerPollKeepsProcessesWhenExeAndCwdAreMissing(t *testing.T) {
	t.Parallel()

	fixture := newPollingFixture(t)
	snapshot := mustPollFixtureSnapshot(t, fixture)
	process := findProcessMetric(t, snapshot.Processes, 103)

	if process.Exe != "" {
		t.Fatalf("Exe = %q, want empty", process.Exe)
	}
	if process.Cwd != "" {
		t.Fatalf("Cwd = %q, want empty", process.Cwd)
	}
}

// Requirements: process-metrics-svc/FR-003
func TestPollingWorkerPollTreatsUnreadableFDDirectoriesAsZero(t *testing.T) {
	t.Parallel()

	fixture := newPollingFixture(t)
	snapshot := mustPollFixtureSnapshot(t, fixture)
	process := findProcessMetric(t, snapshot.Processes, 104)

	if process.OpenFDs != 0 {
		t.Fatalf("OpenFDs = %d, want 0", process.OpenFDs)
	}
}

// Requirements: process-metrics-svc/FR-003
func TestPollingWorkerPollSkipsDisappearingProcesses(t *testing.T) {
	t.Parallel()

	fixture := newPollingFixture(t)
	if err := os.Remove(filepath.Join(fixture.procRoot, "105", "status")); err != nil {
		t.Fatalf("Remove(status) error = %v", err)
	}

	snapshot := mustPollFixtureSnapshot(t, fixture)
	for _, process := range snapshot.Processes {
		if process.PID == 105 {
			t.Fatal("expected disappearing process to be skipped")
		}
	}
}

// Requirements: process-metrics-svc/FR-004
func TestDeriveStartTimeConvertsClockTicksUsingBootTime(t *testing.T) {
	t.Parallel()

	boot := time.Unix(100, 0).UTC()
	got := deriveStartTime(&boot, 250, 100)
	if got == nil {
		t.Fatal("deriveStartTime() = nil, want timestamp")
	}

	want := time.Unix(102, 500000000).UTC()
	if !got.Equal(want) {
		t.Fatalf("StartTime = %v, want %v", got, want)
	}
}

// Requirements: process-metrics-svc/FR-004
func TestPollingWorkerPollConvertsResidentPagesToBytes(t *testing.T) {
	t.Parallel()

	fixture := newPollingFixture(t)
	snapshot := mustPollFixtureSnapshot(t, fixture)
	process := findProcessMetric(t, snapshot.Processes, 102)

	if process.Memory.RSSBytes != 2*4096 {
		t.Fatalf("RSSBytes = %d, want %d", process.Memory.RSSBytes, 2*4096)
	}
}

// Requirements: process-metrics-svc/FR-001
func TestPollingWorkerStartEmitsSnapshotAfterTick(t *testing.T) {
	t.Parallel()

	fixture := newPollingFixture(t)
	ticks := make(chan struct{}, 1)
	output := make(chan metrics.ProcessMetricsSnapshot, 1)
	errorsCh := make(chan error, 1)
	worker := NewPollingWorker(fixture.procRoot, ticks, output, errorsCh)
	worker.pageSize = 4096
	worker.clockTicksPerSecond = 100

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	worker.Start(ctx)
	ticks <- struct{}{}

	select {
	case snapshot := <-output:
		if len(snapshot.Processes) == 0 {
			t.Fatal("len(Processes) = 0, want non-empty snapshot")
		}
	case err := <-errorsCh:
		t.Fatalf("unexpected poll error = %v", err)
	case <-time.After(time.Second):
		t.Fatal("snapshot was not emitted after tick")
	}
}

// Requirements: process-metrics-svc/FR-001
func TestPollingWorkerPollAndSendReportsPollErrors(t *testing.T) {
	t.Parallel()

	output := make(chan metrics.ProcessMetricsSnapshot, 1)
	errorsCh := make(chan error, 1)
	worker := NewPollingWorker("/non-existent-proc", nil, output, errorsCh)
	worker.pollAndSend(context.Background())

	select {
	case err := <-errorsCh:
		if err == nil {
			t.Fatal("poll error = nil, want read failure")
		}
	case <-time.After(time.Second):
		t.Fatal("poll error was not reported")
	}
}

// Requirements: process-metrics-svc/FR-001
func TestWaitNextPollReturnsFalseWhenTicksChannelCloses(t *testing.T) {
	t.Parallel()

	ticks := make(chan struct{})
	close(ticks)

	worker := NewPollingWorker("", ticks, nil, nil)
	if worker.waitNextPoll(context.Background()) {
		t.Fatal("waitNextPoll() = true, want false after channel close")
	}
}

func mustPollFixtureSnapshot(t *testing.T, fixture pollingFixture) metrics.ProcessMetricsSnapshot {
	t.Helper()

	worker := NewPollingWorker(fixture.procRoot, nil, nil, nil)
	worker.pageSize = 4096
	worker.clockTicksPerSecond = 100
	worker.lookupUsernameByUID = func(uid string) (*user.User, error) {
		if uid == "1000" {
			return &user.User{Username: "alpha"}, nil
		}
		return nil, os.ErrNotExist
	}
	worker.now = func() time.Time { return time.Unix(1700000000, 0).UTC() }

	snapshot, err := worker.poll()
	if err != nil {
		t.Fatalf("poll() error = %v", err)
	}

	return snapshot
}

type pollingFixture struct {
	procRoot string
}

func newPollingFixture(t *testing.T) pollingFixture {
	t.Helper()

	root := t.TempDir()
	procRoot := filepath.Join(root, "proc")
	fstest.MustMkdirAll(t, procRoot, 0o755)
	fstest.MustWriteFile(t, filepath.Join(procRoot, "stat"), "cpu 1 2 3 4 5 6 7 8 9 10\nbtime 100\n", 0o755, 0o644)
	writePollingFixtureProcesses(t, procRoot, buildPollingProcessFixtures())
	fstest.MustWriteFile(t, filepath.Join(procRoot, "self"), "", 0o755, 0o644)

	return pollingFixture{procRoot: procRoot}
}

func assertAlphaProcessMetric(t *testing.T, got metrics.ProcessMetric) {
	t.Helper()

	assertAlphaIdentity(t, got)
	assertAlphaPathsAndFDs(t, got)
	assertAlphaResourceUsage(t, got)
}

func assertAlphaIdentity(t *testing.T, got metrics.ProcessMetric) {
	t.Helper()

	if got.Name != "alpha process" {
		t.Fatalf("Name = %q, want %q", got.Name, "alpha process")
	}
	if got.Cmdline != "/usr/bin/alpha --mode foreground" {
		t.Fatalf("Cmdline = %q", got.Cmdline)
	}
}

func assertAlphaPathsAndFDs(t *testing.T, got metrics.ProcessMetric) {
	t.Helper()

	if got.Exe != "/usr/bin/alpha" {
		t.Fatalf("Exe = %q, want /usr/bin/alpha", got.Exe)
	}
	if got.Cwd != "/srv/alpha" {
		t.Fatalf("Cwd = %q, want /srv/alpha", got.Cwd)
	}
	if got.OpenFDs != 3 {
		t.Fatalf("OpenFDs = %d, want 3", got.OpenFDs)
	}
}

func assertAlphaResourceUsage(t *testing.T, got metrics.ProcessMetric) {
	t.Helper()

	if got.CPU.TotalTicks != 30 {
		t.Fatalf("CPU.TotalTicks = %d, want 30", got.CPU.TotalTicks)
	}
	if got.Memory.RSSBytes != 5*4096 {
		t.Fatalf("Memory.RSSBytes = %d, want %d", got.Memory.RSSBytes, 5*4096)
	}
	if got.Memory.VMSBytes != 8192 {
		t.Fatalf("Memory.VMSBytes = %d, want 8192", got.Memory.VMSBytes)
	}
}

func writePollingFixtureProcesses(t *testing.T, procRoot string, fixtures []processFixture) {
	t.Helper()

	for _, fixture := range fixtures {
		writeProcessFixture(t, procRoot, fixture)
	}
}

func buildPollingProcessFixtures() []processFixture {
	return []processFixture{
		alphaProcessFixture(),
		workerBetaFixture(),
		goneLinksFixture(),
		fdDeniedFixture(),
		disappearingFixture(),
		kernelWorkerFixture(),
	}
}

func alphaProcessFixture() processFixture {
	return processFixture{
		pid:       101,
		ppid:      1,
		name:      "alpha process",
		state:     "S",
		userTicks: 11,
		sysTicks:  19,
		threads:   4,
		startTick: 250,
		vsize:     8192,
		rssPages:  5,
		uid:       1000,
		gid:       1000,
		cmdline:   []string{"/usr/bin/alpha", "--mode", "foreground"},
		exe:       "/usr/bin/alpha",
		cwd:       "/srv/alpha",
		fdCount:   3,
	}
}

func workerBetaFixture() processFixture {
	return processFixture{
		pid:       102,
		ppid:      101,
		name:      "worker (beta)",
		state:     "R",
		userTicks: 1,
		sysTicks:  2,
		threads:   2,
		startTick: 100,
		vsize:     4096,
		rssPages:  2,
		uid:       1001,
		gid:       1001,
		cmdline:   []string{"/usr/bin/worker"},
		fdCount:   1,
	}
}

func goneLinksFixture() processFixture {
	return processFixture{
		pid:       103,
		ppid:      1,
		name:      "gone-links",
		state:     "S",
		userTicks: 5,
		sysTicks:  5,
		threads:   1,
		startTick: 50,
		vsize:     2048,
		rssPages:  1,
		uid:       1002,
		gid:       1002,
		cmdline:   []string{"/usr/bin/missing-links"},
	}
}

func fdDeniedFixture() processFixture {
	return processFixture{
		pid:       104,
		ppid:      1,
		name:      "fd-denied",
		state:     "S",
		userTicks: 3,
		sysTicks:  4,
		threads:   1,
		startTick: 70,
		vsize:     1024,
		rssPages:  1,
		uid:       1003,
		gid:       1003,
		cmdline:   []string{"/usr/bin/fd-denied"},
		fdMode:    0,
	}
}

func disappearingFixture() processFixture {
	return processFixture{
		pid:       105,
		ppid:      1,
		name:      "disappears",
		state:     "S",
		userTicks: 1,
		sysTicks:  1,
		threads:   1,
		startTick: 80,
		vsize:     1024,
		rssPages:  1,
		uid:       1004,
		gid:       1004,
		cmdline:   []string{"/usr/bin/disappears"},
	}
}

func kernelWorkerFixture() processFixture {
	return processFixture{
		pid:       106,
		ppid:      2,
		name:      "kworker/0:1",
		state:     "I",
		userTicks: 0,
		sysTicks:  1,
		threads:   1,
		startTick: 10,
		vsize:     0,
		rssPages:  0,
		uid:       0,
		gid:       0,
		cmdline:   nil,
	}
}

type processFixture struct {
	pid       int
	ppid      int
	name      string
	state     string
	userTicks uint64
	sysTicks  uint64
	threads   int
	startTick uint64
	vsize     uint64
	rssPages  int64
	uid       uint32
	gid       uint32
	cmdline   []string
	exe       string
	cwd       string
	fdCount   int
	fdMode    fs.FileMode
}

func writeProcessFixture(t *testing.T, procRoot string, fixture processFixture) {
	t.Helper()

	processRoot := filepath.Join(procRoot, strconv.Itoa(fixture.pid))
	fstest.MustMkdirAll(t, processRoot, 0o755)
	fstest.MustWriteFile(t, filepath.Join(processRoot, "stat"), buildStatFixture(fixture), 0o755, 0o644)
	fstest.MustWriteFile(t, filepath.Join(processRoot, "status"), buildStatusFixture(fixture), 0o755, 0o644)
	fstest.MustWriteFile(t, filepath.Join(processRoot, "cmdline"), buildCmdlineFixture(fixture.cmdline), 0o755, 0o644)

	fdRoot := filepath.Join(processRoot, "fd")
	fstest.MustMkdirAll(t, fdRoot, 0o755)
	for idx := 0; idx < fixture.fdCount; idx++ {
		fdPath := filepath.Join(fdRoot, strconv.Itoa(idx))
		fstest.MustWriteFile(t, fdPath, "", 0o755, 0o644)
	}
	if fixture.fdMode != 0 {
		if err := os.Chmod(fdRoot, fixture.fdMode); err != nil {
			t.Fatalf("Chmod(%s) error = %v", fdRoot, err)
		}
	}

	if fixture.exe != "" {
		fstest.MustSymlink(t, fixture.exe, filepath.Join(processRoot, "exe"))
	}
	if fixture.cwd != "" {
		fstest.MustSymlink(t, fixture.cwd, filepath.Join(processRoot, "cwd"))
	}
}

func buildStatFixture(fixture processFixture) string {
	fields := []string{
		fixture.state,
		strconv.Itoa(fixture.ppid),
		"0", "0", "0", "0", "0", "0", "0", "0", "0",
		strconv.FormatUint(fixture.userTicks, 10),
		strconv.FormatUint(fixture.sysTicks, 10),
		"0", "0", "0", "0",
		strconv.Itoa(fixture.threads),
		"0",
		strconv.FormatUint(fixture.startTick, 10),
		strconv.FormatUint(fixture.vsize, 10),
		strconv.FormatInt(fixture.rssPages, 10),
	}

	return strconv.Itoa(fixture.pid) + " (" + fixture.name + ") " + strings.Join(fields, " ")
}

func buildStatusFixture(fixture processFixture) string {
	return "Name:\t" + fixture.name + "\n" +
		"State:\t" + fixture.state + " (state)\n" +
		"Uid:\t" + strconv.FormatUint(uint64(fixture.uid), 10) + "\t0\t0\t0\n" +
		"Gid:\t" + strconv.FormatUint(uint64(fixture.gid), 10) + "\t0\t0\t0\n" +
		"Threads:\t" + strconv.Itoa(fixture.threads) + "\n"
}

func buildCmdlineFixture(parts []string) string {
	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, "\x00") + "\x00"
}

func findProcessMetric(t *testing.T, processes []metrics.ProcessMetric, pid int) metrics.ProcessMetric {
	t.Helper()

	for _, process := range processes {
		if process.PID == pid {
			return process
		}
	}

	t.Fatalf("pid %d not found", pid)
	return metrics.ProcessMetric{}
}

func TestParseProcessStatusExtractsPrimaryIDsAndThreads(t *testing.T) {
	t.Parallel()

	sample, err := parseProcessStatus("Uid:\t1000\t1000\t1000\t1000\nGid:\t1001\t1001\t1001\t1001\nThreads:\t7\n")
	if err != nil {
		t.Fatalf("parseProcessStatus() error = %v", err)
	}

	want := processStatusSample{UID: 1000, GID: 1001, Threads: 7}
	if !reflect.DeepEqual(sample, want) {
		t.Fatalf("parseProcessStatus() = %#v, want %#v", sample, want)
	}
}
