// Package servicemonitor contains factual systemd state notifications shared
// by System Service Monitor and Watchdog.
package servicemonitor

const (
	StateSubject   = "service.monitor.events.state"
	CurrentVersion = 1
)

// Condition is the monitor's classification of observed systemd properties.
type Condition string

const (
	Active            Condition = "active"
	Inactive          Condition = "inactive"
	Failed            Condition = "failed"
	Restarting        Condition = "restarting"
	PersistentFailure Condition = "persistent_failure"
	Restored          Condition = "restored"
)

// StateEvent reports a transition; Watchdog must independently inspect the
// configured protected unit before deciding whether to act.
type StateEvent struct {
	Version      int       `json:"version" validate:"eq=1"`
	EventID      string    `json:"event_id" validate:"required,min=1,max=64,printascii"`
	Unit         string    `json:"unit" validate:"required,min=9,max=256,endswith=.service"`
	Condition    Condition `json:"condition" validate:"required,oneof=active inactive failed restarting persistent_failure restored"`
	ActiveState  string    `json:"active_state" validate:"required,min=1,max=64,printascii"`
	SubState     string    `json:"sub_state" validate:"required,min=1,max=64,printascii"`
	Result       string    `json:"result" validate:"max=128,printascii"`
	RestartCount uint64    `json:"restart_count"`
	ObservedAt   string    `json:"observed_at" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
}
