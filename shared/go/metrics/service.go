package metrics

import "time"

// ServiceMetricsSnapshot is the top-level service metrics snapshot payload.
type ServiceMetricsSnapshot struct {
	// Timestamp is the time at which the snapshot was collected.
	Timestamp time.Time `json:"timestamp"`

	// Services contains one snapshot entry per discovered service unit.
	Services []ServiceUnitSnapshot `json:"services"`
}

// ServiceUnitSnapshot represents one systemd-managed service unit snapshot.
type ServiceUnitSnapshot struct {
	// Name is the systemd unit name, for example "nginx.service".
	Name string `json:"name"`

	// Manager identifies the service manager that owns the unit.
	Manager string `json:"manager"`

	// UnitType identifies the unit family within the manager.
	UnitType string `json:"unit_type"`

	// Description stores the unit description when it can be resolved.
	Description *string `json:"description,omitempty"`

	// LoadState stores the resolved unit load state.
	LoadState *string `json:"load_state,omitempty"`

	// ActiveState stores the resolved coarse runtime activity state.
	ActiveState *string `json:"active_state,omitempty"`

	// SubState stores the resolved detailed runtime activity state.
	SubState *string `json:"sub_state,omitempty"`

	// EnabledState stores the resolved unit-file enablement state.
	EnabledState *string `json:"enabled_state,omitempty"`

	// Result stores the unit result when it can be resolved.
	Result *string `json:"result,omitempty"`

	// MainPID stores the primary PID when one can be resolved.
	MainPID *uint32 `json:"main_pid,omitempty"`

	// ControlPID stores the control PID when one can be resolved.
	ControlPID *uint32 `json:"control_pid,omitempty"`

	// PIDs stores PIDs that belong to the unit cgroup.
	PIDs []uint32 `json:"pids,omitempty"`

	// CGroup stores the unit cgroup path relative to the cgroup root.
	CGroup *string `json:"cgroup,omitempty"`

	// Slice stores the owning slice when it can be resolved.
	Slice *string `json:"slice,omitempty"`

	// FragmentPath stores the resolved unit file path when available.
	FragmentPath *string `json:"fragment_path,omitempty"`

	// StartedAt stores the earliest resolved process start time when available.
	StartedAt *time.Time `json:"started_at,omitempty"`

	// Resources stores cgroup-scoped resource values when available.
	Resources *ServiceUnitResources `json:"resources,omitempty"`
}

// ServiceUnitResources stores cgroup-scoped service resource values.
type ServiceUnitResources struct {
	// MemoryBytes stores current cgroup memory usage in bytes when available.
	MemoryBytes *uint64 `json:"memory_bytes,omitempty"`

	// CPUUsageUsec stores cumulative cgroup CPU usage in microseconds when available.
	CPUUsageUsec *uint64 `json:"cpu_usage_usec,omitempty"`
}
