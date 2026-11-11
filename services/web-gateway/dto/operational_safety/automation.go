// Package operationalsafety defines validated browser-facing operator DTOs.
package operationalsafety

import (
	"time"

	"lite-nas/services/web-gateway/dto"
	contract "lite-nas/shared/contracts/operationalsafety"
)

// EmptyInput represents a bodyless read request.
type EmptyInput struct{}

// RuleStateInput changes the temporary enabled state of one configured rule.
type RuleStateInput struct {
	RuleID string `path:"rule_id" minLength:"1" maxLength:"64" pattern:"^[a-z][a-z0-9_]*$"`
	Body   OperationalSafetyRuleStateBody
}

// OperationalSafetyRuleStateBody requires an explicit desired enabled state.
type OperationalSafetyRuleStateBody struct {
	Enabled *bool  `json:"enabled" required:"true"`
	Reason  string `json:"reason,omitempty" maxLength:"256"`
}

// ForceSetInput selects a rule-declared level for an existing task.
type ForceSetInput struct {
	TaskID string `path:"task_id" minLength:"1" maxLength:"64" pattern:"^[A-Za-z0-9]+$"`
	Body   struct {
		Level  contract.OperationalSafetyLevel `json:"level" enum:"normal,warning,error,critical" required:"true"`
		Reason string                          `json:"reason,omitempty" maxLength:"256"`
	}
}

// ForceStopInput starts staged recovery of an existing task.
type ForceStopInput struct {
	TaskID string `path:"task_id" minLength:"1" maxLength:"64" pattern:"^[A-Za-z0-9]+$"`
	Body   struct {
		Reason string `json:"reason,omitempty" maxLength:"256"`
	}
}

// RulesOutput wraps configured rule metadata in the gateway envelope.
type RulesOutput struct{ Body RulesBody }

// RulesBody contains successful rule metadata.
type RulesBody struct {
	dto.ResponseMeta
	Data contract.OperationalSafetyRulesResponse `json:"data"`
}

// TasksOutput wraps active task metadata in the gateway envelope.
type TasksOutput struct{ Body TasksBody }

// TasksBody contains successful task metadata.
type TasksBody struct {
	dto.ResponseMeta
	Data contract.OperationalSafetyTasksResponse `json:"data"`
}

// OperationOutput wraps an operator outcome in the gateway envelope.
type OperationOutput struct{ Body OperationBody }

// OperationBody contains the service's verified or pending outcome.
type OperationBody struct {
	dto.ResponseMeta
	Data contract.OperationalSafetyOperationResponse `json:"data"`
}

// NewRulesOutput creates a successful rule-list response.
func NewRulesOutput(data contract.OperationalSafetyRulesResponse) *RulesOutput {
	return &RulesOutput{Body: RulesBody{ResponseMeta: dto.ResponseMeta{Success: true, Timestamp: time.Now()}, Data: data}}
}

// NewTasksOutput creates a successful task-list response.
func NewTasksOutput(data contract.OperationalSafetyTasksResponse) *TasksOutput {
	return &TasksOutput{Body: TasksBody{ResponseMeta: dto.ResponseMeta{Success: true, Timestamp: time.Now()}, Data: data}}
}

// NewOperationOutput creates a successful operator outcome response.
func NewOperationOutput(data contract.OperationalSafetyOperationResponse) *OperationOutput {
	return &OperationOutput{Body: OperationBody{ResponseMeta: dto.ResponseMeta{Success: true, Timestamp: time.Now()}, Data: data}}
}
