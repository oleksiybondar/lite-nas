package workers

import (
	"bytes"
	"context"
	"errors"
	"math"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"lite-nas/shared/metrics"
)

// PollingWorker periodically reads procfs and emits process snapshots into an
// output channel.
type PollingWorker struct {
	procRoot            string
	ticks               <-chan struct{}
	output              chan<- metrics.ProcessMetricsSnapshot
	errors              chan<- error
	pageSize            uint64
	clockTicksPerSecond uint64
	lookupUsernameByUID func(string) (*user.User, error)
	now                 func() time.Time
}

// NewPollingWorker creates a PollingWorker with the dependencies required for
// periodic process snapshot collection.
func NewPollingWorker(
	procRoot string,
	ticks <-chan struct{},
	output chan<- metrics.ProcessMetricsSnapshot,
	errors chan<- error,
) PollingWorker {
	pageSize := os.Getpagesize()
	if pageSize < 0 {
		pageSize = 0
	}

	return PollingWorker{
		procRoot:            procRoot,
		ticks:               ticks,
		output:              output,
		errors:              errors,
		pageSize:            uint64(pageSize),
		clockTicksPerSecond: systemClockTicksPerSecond(),
		lookupUsernameByUID: user.LookupId,
		now:                 time.Now,
	}
}

// Start launches the polling worker in a separate goroutine.
func (w PollingWorker) Start(ctx context.Context) {
	go w.run(ctx)
}

// run executes the polling loop until the provided context is canceled.
func (w PollingWorker) run(ctx context.Context) {
	for {
		if !w.waitNextPoll(ctx) {
			return
		}

		w.pollAndSend(ctx)
	}
}

// waitNextPoll blocks until the next poll tick arrives or the context is
// canceled.
func (w PollingWorker) waitNextPoll(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	case _, ok := <-w.ticks:
		return ok
	}
}

// pollAndSend performs one polling cycle and forwards the resulting snapshot.
func (w PollingWorker) pollAndSend(ctx context.Context) {
	snapshot, err := w.poll()
	if err != nil {
		w.emitError(ctx, err)
		return
	}

	select {
	case <-ctx.Done():
		return
	case w.output <- snapshot:
	}
}

// emitError reports one polling error to the runtime when possible.
func (w PollingWorker) emitError(ctx context.Context, err error) {
	if err == nil {
		return
	}

	select {
	case <-ctx.Done():
		return
	case w.errors <- err:
	default:
	}
}

// poll reads procfs and composes one live process snapshot.
func (w PollingWorker) poll() (metrics.ProcessMetricsSnapshot, error) {
	bootTime, _ := w.readBootTime()

	entries, err := os.ReadDir(w.procRoot)
	if err != nil {
		return metrics.ProcessMetricsSnapshot{}, err
	}

	processes := make([]metrics.ProcessMetric, 0, len(entries))
	for _, entry := range entries {
		process, ok := w.collectProcess(entry.Name(), bootTime)
		if !ok {
			continue
		}
		processes = append(processes, process)
	}

	sort.Slice(processes, func(i int, j int) bool {
		return processes[i].PID < processes[j].PID
	})

	return metrics.ProcessMetricsSnapshot{
		Timestamp: w.now(),
		Processes: processes,
	}, nil
}

// collectProcess attempts to collect one process entry and reports whether the
// PID should be included in the snapshot.
func (w PollingWorker) collectProcess(pidDirName string, bootTime *time.Time) (metrics.ProcessMetric, bool) {
	pid, err := strconv.Atoi(pidDirName)
	if err != nil || pid <= 0 {
		return metrics.ProcessMetric{}, false
	}

	processPath := filepath.Join(w.procRoot, pidDirName)
	statSample, statusSample, ok := readProcessSamples(processPath)
	if !ok {
		return metrics.ProcessMetric{}, false
	}

	process := metrics.ProcessMetric{
		PID:     statSample.PID,
		PPID:    statSample.PPID,
		Name:    statSample.Name,
		State:   statSample.State,
		UID:     statusSample.UID,
		GID:     statusSample.GID,
		Threads: statusSample.Threads,
		CPU: metrics.ProcessCPU{
			UserTicks:   statSample.UserTicks,
			SystemTicks: statSample.SystemTicks,
			TotalTicks:  statSample.UserTicks + statSample.SystemTicks,
		},
		Memory: metrics.ProcessMemory{
			RSSBytes: rssPagesToBytes(statSample.RSSPages, w.pageSize),
			VMSBytes: statSample.VirtualMemoryBytes,
		},
	}

	process.Cmdline = readOptionalCmdline(filepath.Join(processPath, "cmdline"))
	process.Exe = readOptionalLink(filepath.Join(processPath, "exe"))
	process.Cwd = readOptionalLink(filepath.Join(processPath, "cwd"))
	process.OpenFDs = readOptionalFDCount(filepath.Join(processPath, "fd"))
	process.Username = w.lookupOptionalUsername(process.UID)
	process.StartTime = deriveStartTime(bootTime, statSample.StartTicks, w.clockTicksPerSecond)

	return process, true
}

// readProcessSamples loads the required procfs samples for one PID.
func readProcessSamples(processPath string) (processStatSample, processStatusSample, bool) {
	statSample, err := readProcessStat(filepath.Join(processPath, "stat"))
	if err != nil {
		return processStatSample{}, processStatusSample{}, false
	}

	statusSample, err := readProcessStatus(filepath.Join(processPath, "status"))
	if err != nil {
		return processStatSample{}, processStatusSample{}, false
	}

	return statSample, statusSample, true
}

// readBootTime reads the kernel boot time from /proc/stat when available.
func (w PollingWorker) readBootTime() (*time.Time, error) {
	data, err := os.ReadFile(filepath.Join(w.procRoot, "stat")) // #nosec G304 -- proc root is runtime-owned configuration.
	if err != nil {
		return nil, err
	}

	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "btime ") {
			continue
		}

		seconds, err := strconv.ParseInt(strings.TrimSpace(strings.TrimPrefix(line, "btime ")), 10, 64)
		if err != nil {
			return nil, err
		}

		boot := time.Unix(seconds, 0).UTC()
		return &boot, nil
	}

	return nil, errors.New("missing btime in /proc/stat")
}

// lookupOptionalUsername resolves a username for one UID when it can be
// determined.
func (w PollingWorker) lookupOptionalUsername(uid uint32) string {
	if w.lookupUsernameByUID == nil {
		return ""
	}

	account, err := w.lookupUsernameByUID(strconv.FormatUint(uint64(uid), 10))
	if err != nil || account == nil {
		return ""
	}

	return account.Username
}

type processStatSample struct {
	PID                int
	Name               string
	State              string
	PPID               int
	UserTicks          uint64
	SystemTicks        uint64
	StartTicks         uint64
	VirtualMemoryBytes uint64
	RSSPages           int64
}

func readProcessStat(path string) (processStatSample, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- procfs file path is derived from runtime-owned proc root.
	if err != nil {
		return processStatSample{}, err
	}

	return parseProcessStat(string(bytes.TrimSpace(data)))
}

func parseProcessStat(line string) (processStatSample, error) {
	pid, name, fields, err := splitProcessStatRecord(line)
	if err != nil {
		return processStatSample{}, err
	}

	sample, err := parseProcessStatFields(fields)
	if err != nil {
		return processStatSample{}, err
	}

	sample.PID = pid
	sample.Name = name
	return sample, nil
}

// splitProcessStatRecord separates the procfs stat prefix from the numeric fields.
func splitProcessStatRecord(line string) (int, string, []string, error) {
	openIdx := strings.IndexByte(line, '(')
	closeIdx := strings.LastIndexByte(line, ')')
	if openIdx <= 0 || closeIdx <= openIdx || closeIdx+2 > len(line) {
		return 0, "", nil, errors.New("invalid process stat format")
	}

	pid, err := strconv.Atoi(strings.TrimSpace(line[:openIdx]))
	if err != nil {
		return 0, "", nil, err
	}

	return pid, line[openIdx+1 : closeIdx], strings.Fields(line[closeIdx+2:]), nil
}

// parseProcessStatFields parses the numeric procfs stat fields used by LiteNAS.
func parseProcessStatFields(fields []string) (processStatSample, error) {
	if len(fields) < 22 {
		return processStatSample{}, errors.New("short process stat record")
	}

	cpuSample, err := parseProcessCPUFields(fields)
	if err != nil {
		return processStatSample{}, err
	}

	memorySample, err := parseProcessMemoryFields(fields)
	if err != nil {
		return processStatSample{}, err
	}

	ppid, err := parseIntField(fields, 1)
	if err != nil {
		return processStatSample{}, err
	}

	return processStatSample{
		State:              fields[0],
		PPID:               ppid,
		UserTicks:          cpuSample.userTicks,
		SystemTicks:        cpuSample.systemTicks,
		StartTicks:         cpuSample.startTicks,
		VirtualMemoryBytes: memorySample.virtualMemoryBytes,
		RSSPages:           memorySample.rssPages,
	}, nil
}

type processCPUSample struct {
	userTicks   uint64
	systemTicks uint64
	startTicks  uint64
}

// parseProcessCPUFields parses the CPU-related stat fields used by the snapshot.
func parseProcessCPUFields(fields []string) (processCPUSample, error) {
	userTicks, err := parseUint64Field(fields, 11)
	if err != nil {
		return processCPUSample{}, err
	}

	systemTicks, err := parseUint64Field(fields, 12)
	if err != nil {
		return processCPUSample{}, err
	}

	startTicks, err := parseUint64Field(fields, 19)
	if err != nil {
		return processCPUSample{}, err
	}

	return processCPUSample{
		userTicks:   userTicks,
		systemTicks: systemTicks,
		startTicks:  startTicks,
	}, nil
}

type processMemorySample struct {
	virtualMemoryBytes uint64
	rssPages           int64
}

// parseProcessMemoryFields parses the memory-related stat fields used by the snapshot.
func parseProcessMemoryFields(fields []string) (processMemorySample, error) {
	virtualMemoryBytes, err := parseUint64Field(fields, 20)
	if err != nil {
		return processMemorySample{}, err
	}

	rssPages, err := parseInt64Field(fields, 21)
	if err != nil {
		return processMemorySample{}, err
	}

	return processMemorySample{
		virtualMemoryBytes: virtualMemoryBytes,
		rssPages:           rssPages,
	}, nil
}

type processStatusSample struct {
	UID     uint32
	GID     uint32
	Threads int
}

func readProcessStatus(path string) (processStatusSample, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- procfs file path is derived from runtime-owned proc root.
	if err != nil {
		return processStatusSample{}, err
	}

	return parseProcessStatus(string(data))
}

func parseProcessStatus(data string) (processStatusSample, error) {
	var sample processStatusSample
	var fields processStatusFields

	for _, line := range strings.Split(data, "\n") {
		if err := parseProcessStatusLine(line, &sample, &fields); err != nil {
			return processStatusSample{}, err
		}
	}

	if !fields.complete() {
		return processStatusSample{}, errors.New("incomplete process status record")
	}

	return sample, nil
}

type processStatusFields struct {
	haveUID     bool
	haveGID     bool
	haveThreads bool
}

// complete reports whether all required status fields were parsed.
func (f processStatusFields) complete() bool {
	return f.haveUID && f.haveGID && f.haveThreads
}

// parseProcessStatusLine extracts one status field from a procfs status line.
func parseProcessStatusLine(line string, sample *processStatusSample, fields *processStatusFields) error {
	if strings.HasPrefix(line, "Uid:") {
		return parseProcessStatusUID(line, sample, fields)
	}

	if strings.HasPrefix(line, "Gid:") {
		return parseProcessStatusGID(line, sample, fields)
	}

	if strings.HasPrefix(line, "Threads:") {
		return parseProcessStatusThreads(line, sample, fields)
	}

	return nil
}

// parseProcessStatusUID parses the primary UID status field.
func parseProcessStatusUID(line string, sample *processStatusSample, fields *processStatusFields) error {
	value, err := parseFirstUint32Field(strings.TrimPrefix(line, "Uid:"))
	if err != nil {
		return err
	}

	sample.UID = value
	fields.haveUID = true
	return nil
}

// parseProcessStatusGID parses the primary GID status field.
func parseProcessStatusGID(line string, sample *processStatusSample, fields *processStatusFields) error {
	value, err := parseFirstUint32Field(strings.TrimPrefix(line, "Gid:"))
	if err != nil {
		return err
	}

	sample.GID = value
	fields.haveGID = true
	return nil
}

// parseProcessStatusThreads parses the thread-count status field.
func parseProcessStatusThreads(line string, sample *processStatusSample, fields *processStatusFields) error {
	value, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "Threads:")))
	if err != nil {
		return err
	}

	sample.Threads = value
	fields.haveThreads = true
	return nil
}

func parseFirstUint32Field(value string) (uint32, error) {
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return 0, errors.New("missing numeric field")
	}

	parsed, err := strconv.ParseUint(fields[0], 10, 32)
	if err != nil {
		return 0, err
	}

	return uint32(parsed), nil
}

func readOptionalCmdline(path string) string {
	data, err := os.ReadFile(path) // #nosec G304 -- procfs file path is derived from runtime-owned proc root.
	if err != nil || len(data) == 0 {
		return ""
	}

	parts := strings.Split(string(bytes.TrimRight(data, "\x00")), "\x00")
	filtered := parts[:0]
	for _, part := range parts {
		if part != "" {
			filtered = append(filtered, part)
		}
	}

	return strings.Join(filtered, " ")
}

func readOptionalLink(path string) string {
	target, err := os.Readlink(path)
	if err != nil {
		return ""
	}

	return target
}

func readOptionalFDCount(path string) int {
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0
	}

	return len(entries)
}

func deriveStartTime(bootTime *time.Time, startTicks uint64, hz uint64) *time.Time {
	if bootTime == nil || hz == 0 {
		return nil
	}

	secondsUint := startTicks / hz
	if secondsUint > math.MaxInt64 {
		return nil
	}

	nanosecondsUint := (startTicks % hz) * uint64(time.Second) / hz
	if nanosecondsUint > math.MaxInt64 {
		return nil
	}

	seconds := int64(secondsUint)
	nanoseconds := int64(nanosecondsUint)
	start := bootTime.Add(time.Duration(seconds)*time.Second + time.Duration(nanoseconds))
	return &start
}

// rssPagesToBytes converts resident pages to bytes while clamping negative values.
func rssPagesToBytes(rssPages int64, pageSize uint64) uint64 {
	if rssPages <= 0 {
		return 0
	}

	return uint64(rssPages) * pageSize
}

// parseIntField parses one indexed decimal field from a procfs record.
func parseIntField(fields []string, idx int) (int, error) {
	return strconv.Atoi(fields[idx])
}

// parseUint64Field parses one indexed unsigned decimal field from a procfs record.
func parseUint64Field(fields []string, idx int) (uint64, error) {
	return strconv.ParseUint(fields[idx], 10, 64)
}

// parseInt64Field parses one indexed signed decimal field from a procfs record.
func parseInt64Field(fields []string, idx int) (int64, error) {
	return strconv.ParseInt(fields[idx], 10, 64)
}

func systemClockTicksPerSecond() uint64 {
	// Procfs CPU counters use Linux USER_HZ units. LiteNAS keeps that raw
	// representation stable in the snapshot contract and only needs a consistent
	// tick baseline when deriving optional start_time values.
	return 100
}
