// Package securityautomation defines the operator NATS request/reply surface.
package securityautomation

const (
	// RulesListSubject lists configured rules, including disabled definitions.
	RulesListSubject = "security-automation.rpc.rules.list"
	// ActiveListSubject lists incidents whose mitigation or restore is pending.
	ActiveListSubject = "security-automation.rpc.active.list"
	// RollbackSubject requests restoration of one owned incident.
	RollbackSubject = "security-automation.rpc.rollback"
	// CooldownResetSubject clears one rule and correlation key's volatile detection state.
	CooldownResetSubject = "security-automation.rpc.cooldown.reset"
	// RuleSetEnabledSubject changes one rule's effective state until restart.
	RuleSetEnabledSubject = "security-automation.rpc.rule.set-enabled"
	// ForceNormalizeSubject retries owned restoration, then closes only a verified incident.
	ForceNormalizeSubject = "security-automation.rpc.force-normalize"
	// ForceMitigateSubject creates an operator-requested incident after operand validation.
	ForceMitigateSubject = "security-automation.rpc.force-mitigate"
)

// MitigationState is the service-owned lifecycle state shown to operators.
type MitigationState string

const (
	// StateNormal means this rule has no active mitigation for the reported target.
	StateNormal MitigationState = "normal"
	// StatePending means a mitigation decision is recorded but not yet verified.
	StatePending MitigationState = "pending"
	// StateMitigating means the effect was applied and verified.
	StateMitigating MitigationState = "mitigating"
	// StateRestoring means an owned effect is being removed.
	StateRestoring MitigationState = "restoring"
	// StateFailed means an effect or restore failed and needs attention.
	StateFailed MitigationState = "failed"
)

// ListRequest carries the operator access token for a read RPC.
type ListRequest struct {
	AccessToken string `json:"access_token" validate:"required,min=1,max=8192"`
}

// Rule describes one configured rule and its temporary effective state.
type Rule struct {
	ID                string          `json:"id"`
	Name              string          `json:"name"`
	Description       string          `json:"description"`
	Explanation       string          `json:"explanation"`
	ActionKind        string          `json:"action_kind"`
	ActionName        string          `json:"action_name,omitempty"`
	Enabled           bool            `json:"enabled"`
	ConfiguredEnabled bool            `json:"configured_enabled"`
	RequiredOperands  []string        `json:"required_operands,omitempty"`
	State             MitigationState `json:"state"`
	ActiveCount       int             `json:"active_count"`
	LastOutcome       string          `json:"last_outcome"`
}

// RulesListResponse contains all configured rule definitions.
type RulesListResponse struct {
	Rules []Rule `json:"rules"`
}

// ActiveAutomation describes one incident and its last verified effect state.
type ActiveAutomation struct {
	IncidentID  string          `json:"incident_id"`
	Rule        Rule            `json:"rule"`
	Target      string          `json:"target"`
	Mitigation  string          `json:"mitigation"`
	StartedAt   string          `json:"started_at"`
	ExpiresAt   string          `json:"expires_at,omitempty"`
	State       MitigationState `json:"state"`
	LastOutcome string          `json:"last_outcome"`
}

// ActiveListResponse contains incidents still needing operator visibility.
type ActiveListResponse struct {
	Items []ActiveAutomation `json:"items"`
}

// RollbackRequest identifies one incident for an authorized restore attempt.
type RollbackRequest struct {
	AccessToken string `json:"access_token" validate:"required,min=1,max=8192"`
	IncidentID  string `json:"incident_id" validate:"required,min=1,max=64,alphanum"`
	Reason      string `json:"reason,omitempty" validate:"omitempty,max=256"`
}

// ForceMitigateRequest bypasses detection but still requires rule-owned validation and incident handling.
type ForceMitigateRequest struct {
	AccessToken string            `json:"access_token" validate:"required,min=1,max=8192"`
	RuleID      string            `json:"rule_id" validate:"required,min=1,max=64"`
	Operands    map[string]string `json:"operands" validate:"max=16,dive,keys,required,max=64,endkeys,required,max=256"`
	Reason      string            `json:"reason,omitempty" validate:"omitempty,max=256"`
}

// CooldownResetRequest identifies one isolated volatile detector instance.
type CooldownResetRequest struct {
	AccessToken    string `json:"access_token" validate:"required,min=1,max=8192"`
	RuleID         string `json:"rule_id" validate:"required,min=1,max=64"`
	CorrelationKey string `json:"correlation_key" validate:"required,min=1,max=256"`
	Reason         string `json:"reason,omitempty" validate:"omitempty,max=256"`
}

// RuleSetEnabledRequest changes only the in-memory effective enabled state.
type RuleSetEnabledRequest struct {
	AccessToken string `json:"access_token" validate:"required,min=1,max=8192"`
	RuleID      string `json:"rule_id" validate:"required,min=1,max=64"`
	Enabled     bool   `json:"enabled"`
	Reason      string `json:"reason,omitempty" validate:"omitempty,max=256"`
}

// OperationResponse reports an accepted operation and any remaining effects.
type OperationResponse struct {
	Completed      bool   `json:"completed"`
	PendingEffects int    `json:"pending_effects"`
	Outcome        string `json:"outcome"`
}
