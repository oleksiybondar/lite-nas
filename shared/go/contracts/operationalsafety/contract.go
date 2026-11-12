// Package operationalsafety defines the operator RPC surface for Operational Safety Automation.
package operationalsafety

const (
	// RulesListSubject lists configured rules and their temporary state.
	RulesListSubject = "operational-safety-automation.rpc.rules.list"
	// TasksListSubject lists active in-memory tasks.
	TasksListSubject = "operational-safety-automation.rpc.tasks.list"
	// RuleSetEnabledSubject changes one rule's state until restart.
	RuleSetEnabledSubject = "operational-safety-automation.rpc.rule.set-enabled"
	// TaskForceSetSubject requests one rule-declared level on an existing task.
	TaskForceSetSubject = "operational-safety-automation.rpc.task.force-set"
	// TaskForceStopSubject starts staged recovery on an existing task.
	TaskForceStopSubject = "operational-safety-automation.rpc.task.force-stop"
	// CPUSetCapSubject verifies a manual host cap when no in-memory task exists.
	CPUSetCapSubject = "operational-safety-automation.rpc.cpu.set-cap"
)

// OperationalSafetyLevel is a configured CPU effect level ordered by severity.
type OperationalSafetyLevel string

const (
	// LevelNormal is the rule's unthrottled or baseline cap.
	LevelNormal OperationalSafetyLevel = "normal"
	// LevelWarning is the first protective cap.
	LevelWarning OperationalSafetyLevel = "warning"
	// LevelError is the second protective cap.
	LevelError OperationalSafetyLevel = "error"
	// LevelCritical is the strongest protective cap.
	LevelCritical OperationalSafetyLevel = "critical"
)

// OperationalSafetyOutcomeState reports whether an operator request was blocked, pending, completed, or failed.
type OperationalSafetyOutcomeState string

const (
	// OutcomeBlocked means a safety condition prevents the requested transition.
	OutcomeBlocked OperationalSafetyOutcomeState = "blocked"
	// OutcomePending means an effect or cooldown is still in progress.
	OutcomePending OperationalSafetyOutcomeState = "pending"
	// OutcomeCompleted means the requested state was verified.
	OutcomeCompleted OperationalSafetyOutcomeState = "completed"
	// OutcomeFailed means the effect or its verification failed.
	OutcomeFailed OperationalSafetyOutcomeState = "failed"
)

// OperationalSafetyListRequest carries the operator token for a read-only RPC.
type OperationalSafetyListRequest struct {
	AccessToken string `json:"access_token" validate:"required,min=1,max=8192"`
}

// OperationalSafetyRule is the operator-visible configuration and temporary enabled state.
type OperationalSafetyRule struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Enabled           bool   `json:"enabled"`
	ConfiguredEnabled bool   `json:"configured_enabled"`
	OutputIDPrefix    string `json:"output_id_prefix"`
	Cooldown          string `json:"cooldown"`
}

// OperationalSafetyRulesResponse returns every configured rule, including disabled rules.
type OperationalSafetyRulesResponse struct {
	Rules []OperationalSafetyRule `json:"rules"`
}

// OperationalSafetyTask reports one in-memory task and its last verified effect.
type OperationalSafetyTask struct {
	ID             string                 `json:"id"`
	RuleID         string                 `json:"rule_id"`
	CorrelationKey string                 `json:"correlation_key"`
	Level          OperationalSafetyLevel `json:"level"`
	ActiveAlertIDs []string               `json:"active_alert_ids"`
	CooldownUntil  string                 `json:"cooldown_until,omitempty"`
	LastOutcome    string                 `json:"last_outcome"`
}

// OperationalSafetyTasksResponse returns active in-memory tasks.
type OperationalSafetyTasksResponse struct {
	Items []OperationalSafetyTask `json:"items"`
}

// OperationalSafetyRuleSetEnabledRequest changes the effective state until service restart.
type OperationalSafetyRuleSetEnabledRequest struct {
	AccessToken string `json:"access_token" validate:"required,min=1,max=8192"`
	RuleID      string `json:"rule_id" validate:"required,min=1,max=64"`
	Enabled     *bool  `json:"enabled" validate:"required"`
	Reason      string `json:"reason,omitempty" validate:"omitempty,max=256"`
}

// OperationalSafetyTaskForceSetRequest requests a rule-declared effect level on an existing task.
type OperationalSafetyTaskForceSetRequest struct {
	AccessToken string                 `json:"access_token" validate:"required,min=1,max=8192"`
	TaskID      string                 `json:"task_id" validate:"required,min=1,max=64,alphanum"`
	Level       OperationalSafetyLevel `json:"level" validate:"required,oneof=normal warning error critical"`
	Reason      string                 `json:"reason,omitempty" validate:"omitempty,max=256"`
}

// OperationalSafetyTaskForceStopRequest starts staged recovery on an existing task.
type OperationalSafetyTaskForceStopRequest struct {
	AccessToken string `json:"access_token" validate:"required,min=1,max=8192"`
	TaskID      string `json:"task_id" validate:"required,min=1,max=64,alphanum"`
	Reason      string `json:"reason,omitempty" validate:"omitempty,max=256"`
}

// OperationalSafetyCPUSetCapRequest requests a direct verified cap after a restart.
type OperationalSafetyCPUSetCapRequest struct {
	AccessToken string `json:"access_token" validate:"required,min=1,max=8192"`
	CapKHz      uint64 `json:"cap_khz" validate:"required,gt=0"`
	Reason      string `json:"reason,omitempty" validate:"omitempty,max=256"`
}

// OperationalSafetyCPUSetCapResponse reports the verified manual cap.
type OperationalSafetyCPUSetCapResponse struct {
	State   OperationalSafetyOutcomeState `json:"state"`
	CapKHz  uint64                        `json:"cap_khz"`
	Outcome string                        `json:"outcome"`
}

// OperationalSafetyOperationResponse reports one request's task and effect outcome.
type OperationalSafetyOperationResponse struct {
	TaskID  string                        `json:"task_id"`
	State   OperationalSafetyOutcomeState `json:"state"`
	Level   OperationalSafetyLevel        `json:"level"`
	Outcome string                        `json:"outcome"`
}
