package services

import (
	"context"

	contract "lite-nas/shared/contracts/securityautomation"
	"lite-nas/shared/messaging"
)

// SecurityAutomationService calls the automation service's operator RPCs.
type SecurityAutomationService interface {
	Rules(context.Context, string) (contract.RulesListResponse, error)
	Active(context.Context, string) (contract.ActiveListResponse, error)
	Rollback(context.Context, contract.RollbackRequest) (contract.OperationResponse, error)
	ForceNormalize(context.Context, contract.RollbackRequest) (contract.OperationResponse, error)
	ForceMitigate(context.Context, contract.ForceMitigateRequest) (contract.OperationResponse, error)
	ResetCooldown(context.Context, contract.CooldownResetRequest) (contract.OperationResponse, error)
	SetEnabled(context.Context, contract.RuleSetEnabledRequest) (contract.OperationResponse, error)
}

type securityAutomationService struct{ client messaging.Client }

// NewSecurityAutomationService binds the gateway to the shared NATS client.
func NewSecurityAutomationService(client messaging.Client) SecurityAutomationService {
	return securityAutomationService{client: client}
}

// Rules returns configured rule metadata and effective enabled states.
func (s securityAutomationService) Rules(ctx context.Context, token string) (contract.RulesListResponse, error) {
	var out contract.RulesListResponse
	err := s.client.Request(ctx, contract.RulesListSubject, contract.ListRequest{AccessToken: token}, &out)
	return out, err
}

// Active returns incidents still needing operator visibility.
func (s securityAutomationService) Active(ctx context.Context, token string) (contract.ActiveListResponse, error) {
	var out contract.ActiveListResponse
	err := s.client.Request(ctx, contract.ActiveListSubject, contract.ListRequest{AccessToken: token}, &out)
	return out, err
}

// Rollback requests restoration of one service-owned incident.
func (s securityAutomationService) Rollback(ctx context.Context, request contract.RollbackRequest) (contract.OperationResponse, error) {
	var out contract.OperationResponse
	err := s.client.Request(ctx, contract.RollbackSubject, request, &out)
	return out, err
}

// ForceNormalize requests restore and normalization after effect verification.
func (s securityAutomationService) ForceNormalize(ctx context.Context, request contract.RollbackRequest) (contract.OperationResponse, error) {
	var out contract.OperationResponse
	err := s.client.Request(ctx, contract.ForceNormalizeSubject, request, &out)
	return out, err
}

// ForceMitigate requests a rule-owned mitigation with explicit operands.
func (s securityAutomationService) ForceMitigate(ctx context.Context, request contract.ForceMitigateRequest) (contract.OperationResponse, error) {
	var out contract.OperationResponse
	err := s.client.Request(ctx, contract.ForceMitigateSubject, request, &out)
	return out, err
}

// ResetCooldown resets one keyed detector's cooldown.
func (s securityAutomationService) ResetCooldown(ctx context.Context, request contract.CooldownResetRequest) (contract.OperationResponse, error) {
	var out contract.OperationResponse
	err := s.client.Request(ctx, contract.CooldownResetSubject, request, &out)
	return out, err
}

// SetEnabled changes one rule's in-memory effective state.
func (s securityAutomationService) SetEnabled(ctx context.Context, request contract.RuleSetEnabledRequest) (contract.OperationResponse, error) {
	var out contract.OperationResponse
	err := s.client.Request(ctx, contract.RuleSetEnabledSubject, request, &out)
	return out, err
}
