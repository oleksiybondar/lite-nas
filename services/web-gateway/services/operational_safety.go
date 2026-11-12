package services

import (
	"context"

	contract "lite-nas/shared/contracts/operationalsafety"
	"lite-nas/shared/messaging"
)

// OperationalSafetyService calls the operational safety operator RPCs.
type OperationalSafetyService interface {
	Rules(context.Context, string) (contract.OperationalSafetyRulesResponse, error)
	Tasks(context.Context, string) (contract.OperationalSafetyTasksResponse, error)
	SetEnabled(context.Context, contract.OperationalSafetyRuleSetEnabledRequest) (contract.OperationalSafetyOperationResponse, error)
	ForceSet(context.Context, contract.OperationalSafetyTaskForceSetRequest) (contract.OperationalSafetyOperationResponse, error)
	ForceStop(context.Context, contract.OperationalSafetyTaskForceStopRequest) (contract.OperationalSafetyOperationResponse, error)
}

type operationalSafetyService struct{ client messaging.Client }

// NewOperationalSafetyService binds the gateway to its backend NATS client.
func NewOperationalSafetyService(client messaging.Client) OperationalSafetyService {
	return operationalSafetyService{client: client}
}

// Rules returns configured rules and their effective enabled state.
func (s operationalSafetyService) Rules(ctx context.Context, token string) (contract.OperationalSafetyRulesResponse, error) {
	var out contract.OperationalSafetyRulesResponse
	err := s.client.Request(ctx, contract.RulesListSubject, contract.OperationalSafetyListRequest{AccessToken: token}, &out)
	return out, err
}

// Tasks returns currently active in-memory tasks.
func (s operationalSafetyService) Tasks(ctx context.Context, token string) (contract.OperationalSafetyTasksResponse, error) {
	var out contract.OperationalSafetyTasksResponse
	err := s.client.Request(ctx, contract.TasksListSubject, contract.OperationalSafetyListRequest{AccessToken: token}, &out)
	return out, err
}

// SetEnabled changes one rule's temporary enabled state.
func (s operationalSafetyService) SetEnabled(ctx context.Context, request contract.OperationalSafetyRuleSetEnabledRequest) (contract.OperationalSafetyOperationResponse, error) {
	var out contract.OperationalSafetyOperationResponse
	err := s.client.Request(ctx, contract.RuleSetEnabledSubject, request, &out)
	return out, err
}

// ForceSet requests a verified rule-declared level on an existing task.
func (s operationalSafetyService) ForceSet(ctx context.Context, request contract.OperationalSafetyTaskForceSetRequest) (contract.OperationalSafetyOperationResponse, error) {
	var out contract.OperationalSafetyOperationResponse
	err := s.client.Request(ctx, contract.TaskForceSetSubject, request, &out)
	return out, err
}

// ForceStop starts staged recovery on an existing task.
func (s operationalSafetyService) ForceStop(ctx context.Context, request contract.OperationalSafetyTaskForceStopRequest) (contract.OperationalSafetyOperationResponse, error) {
	var out contract.OperationalSafetyOperationResponse
	err := s.client.Request(ctx, contract.TaskForceStopSubject, request, &out)
	return out, err
}
