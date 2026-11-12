// Package mountmonitor contains the versioned mount lifecycle contract shared
// by Mount Monitor and Antivirus.
package mountmonitor

const (
	LifecycleSubject = "mount.events.lifecycle"
	CurrentVersion   = 1
)

// Action names one observed mount lifecycle transition.
type Action string

const (
	Mounted   Action = "mounted"
	Unmounted Action = "unmounted"
	Changed   Action = "changed"
)

// LifecycleEvent carries factual mount metadata. MountID is the kernel's
// mountinfo identifier within one boot; consumers must recheck live mounts
// before sensitive work and must not treat this message as authorization.
type LifecycleEvent struct {
	Version    int    `json:"version" validate:"eq=1"`
	EventID    string `json:"event_id" validate:"required,min=1,max=64,printascii"`
	Action     Action `json:"action" validate:"required,oneof=mounted unmounted changed"`
	MountID    uint64 `json:"mount_id" validate:"gt=0"`
	MountPoint string `json:"mount_point" validate:"required,min=1,max=4096"`
	Source     string `json:"source" validate:"max=4096"`
	FSType     string `json:"fs_type" validate:"required,min=1,max=128,printascii"`
	ObservedAt string `json:"observed_at" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
}
