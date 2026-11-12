// Package security_automation defines validated browser-facing operator DTOs.
package security_automation

import (
	"time"

	"lite-nas/services/web-gateway/dto"
	contract "lite-nas/shared/contracts/securityautomation"
)

// EmptyInput represents a bodyless read request.
type EmptyInput struct{}

// IncidentInput identifies one incident and optional operator reason.
type IncidentInput struct {
	IncidentID string `path:"incident_id" minLength:"1" maxLength:"64" pattern:"^[A-Za-z0-9]+$"`
	Body       struct {
		Reason string `json:"reason,omitempty" maxLength:"256"`
	}
}

// InstanceInput identifies one rule instance for a detection-state reset.
type InstanceInput struct {
	RuleID string `path:"rule_id" minLength:"1" maxLength:"64" pattern:"^[A-Za-z0-9_]+$"`
	Body   struct {
		CorrelationKey string `json:"correlation_key" minLength:"1" maxLength:"256"`
		Reason         string `json:"reason,omitempty" maxLength:"256"`
	}
}

// RuleStateInput changes a configured rule's temporary effective state.
type RuleStateInput struct {
	RuleID string `path:"rule_id" minLength:"1" maxLength:"64" pattern:"^[A-Za-z0-9_]+$"`
	Body   struct {
		Enabled bool   `json:"enabled"`
		Reason  string `json:"reason,omitempty" maxLength:"256"`
	}
}

// ForceMitigateInput supplies explicit rule operands for an operator request.
type ForceMitigateInput struct {
	RuleID string `path:"rule_id" minLength:"1" maxLength:"64" pattern:"^[A-Za-z0-9_]+$"`
	Body   struct {
		Operands map[string]string `json:"operands" required:"true" maxProperties:"16"`
		Reason   string            `json:"reason,omitempty" maxLength:"256"`
	}
}

// RulesOutput wraps rule metadata in the standard gateway envelope.
type RulesOutput struct {
	Body struct {
		dto.ResponseMeta
		Data contract.RulesListResponse `json:"data"`
	}
}

// ActiveOutput wraps active incidents in the standard gateway envelope.
type ActiveOutput struct {
	Body struct {
		dto.ResponseMeta
		Data contract.ActiveListResponse `json:"data"`
	}
}

// OperationOutput wraps one mutation outcome in the standard gateway envelope.
type OperationOutput struct {
	Body struct {
		dto.ResponseMeta
		Data contract.OperationResponse `json:"data"`
	}
}

// NewRulesOutput constructs a successful rule-list response.
func NewRulesOutput(data contract.RulesListResponse) *RulesOutput {
	out := &RulesOutput{}
	out.Body.ResponseMeta = dto.ResponseMeta{Success: true, Timestamp: time.Now()}
	out.Body.Data = data
	return out
}

// NewActiveOutput constructs a successful active-incident response.
func NewActiveOutput(data contract.ActiveListResponse) *ActiveOutput {
	out := &ActiveOutput{}
	out.Body.ResponseMeta = dto.ResponseMeta{Success: true, Timestamp: time.Now()}
	out.Body.Data = data
	return out
}

// NewOperationOutput constructs a successful operation response.
func NewOperationOutput(data contract.OperationResponse) *OperationOutput {
	out := &OperationOutput{}
	out.Body.ResponseMeta = dto.ResponseMeta{Success: true, Timestamp: time.Now()}
	out.Body.Data = data
	return out
}
