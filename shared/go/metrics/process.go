package metrics

import "time"

// ProcessMetricsSnapshot is the top-level live process metrics snapshot payload.
type ProcessMetricsSnapshot struct {
	// Timestamp is the time at which the snapshot was collected.
	Timestamp time.Time `json:"timestamp"`

	// Processes contains one entry per process that could be collected during
	// the snapshot cycle.
	Processes []ProcessMetric `json:"processes"`
}

// ProcessMetric represents one Linux process snapshot collected from procfs.
type ProcessMetric struct {
	// PID stores the process identifier.
	PID int `json:"pid"`

	// PPID stores the parent process identifier.
	PPID int `json:"ppid"`

	// Name stores the kernel process name from /proc/[pid]/stat.
	Name string `json:"name"`

	// Cmdline stores the NUL-delimited command line rendered as a shell-like
	// space-separated string when available.
	Cmdline string `json:"cmdline,omitempty"`

	// Exe stores the resolved executable symlink target when available.
	Exe string `json:"exe,omitempty"`

	// Cwd stores the resolved working-directory symlink target when available.
	Cwd string `json:"cwd,omitempty"`

	// State stores the raw single-letter Linux process state.
	State string `json:"state"`

	// UID stores the real user identifier.
	UID uint32 `json:"uid"`

	// GID stores the real group identifier.
	GID uint32 `json:"gid"`

	// Username stores the resolved username when it can be determined.
	Username string `json:"username,omitempty"`

	// CPU stores raw process CPU counters.
	CPU ProcessCPU `json:"cpu"`

	// Memory stores raw process memory counters.
	Memory ProcessMemory `json:"memory"`

	// Threads stores the current thread count.
	Threads int `json:"threads"`

	// OpenFDs stores the number of open file descriptors when it can be read.
	OpenFDs int `json:"open_fds"`

	// StartTime stores the process start time when it can be derived from
	// procfs and the system boot time.
	StartTime *time.Time `json:"start_time,omitempty"`
}

// ProcessCPU stores raw per-process CPU counters in kernel clock ticks.
type ProcessCPU struct {
	// UserTicks stores cumulative user-mode clock ticks.
	UserTicks uint64 `json:"user_ticks"`

	// SystemTicks stores cumulative kernel-mode clock ticks.
	SystemTicks uint64 `json:"system_ticks"`

	// TotalTicks stores the sum of user and system ticks.
	TotalTicks uint64 `json:"total_ticks"`
}

// ProcessMemory stores raw per-process memory counters in bytes.
type ProcessMemory struct {
	// RSSBytes stores resident memory in bytes.
	RSSBytes uint64 `json:"rss_bytes"`

	// VMSBytes stores virtual memory size in bytes.
	VMSBytes uint64 `json:"vms_bytes"`
}
