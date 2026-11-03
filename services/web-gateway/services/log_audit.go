package services

import (
	"context"

	contract "lite-nas/shared/contracts/loganalysis"
	"lite-nas/shared/messaging"
)

// LogAuditService reads configured Log Analysis rules over NATS.
type LogAuditService interface {
	Rules(context.Context, string) (contract.LogAnalysisRulesResponse, error)
}

type logAuditService struct{ client messaging.Client }

// NewLogAuditService binds the gateway to the Log Analysis catalog RPC.
func NewLogAuditService(client messaging.Client) LogAuditService {
	return logAuditService{client: client}
}

// Rules fetches the catalog for the authenticated operator.
func (s logAuditService) Rules(ctx context.Context, token string) (contract.LogAnalysisRulesResponse, error) {
	var response contract.LogAnalysisRulesResponse
	err := s.client.Request(ctx, contract.RulesListSubject, contract.RulesListRequest{AccessToken: token}, &response)
	return response, err
}
