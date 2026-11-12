// Package log_audit defines the browser-facing Log Audit rule list DTO.
package log_audit

import (
	"time"

	"lite-nas/services/web-gateway/dto"
	contract "lite-nas/shared/contracts/loganalysis"
)

// EmptyInput represents the bodyless catalog request.
type EmptyInput struct{}

// LogAuditRulesOutput wraps the rule catalog in the gateway response envelope.
type LogAuditRulesOutput struct {
	Body struct {
		dto.ResponseMeta
		Data contract.LogAnalysisRulesResponse `json:"data"`
	}
}

// NewRulesOutput builds a successful rule catalog response.
func NewRulesOutput(data contract.LogAnalysisRulesResponse) *LogAuditRulesOutput {
	out := &LogAuditRulesOutput{}
	out.Body.ResponseMeta = dto.ResponseMeta{Success: true, Timestamp: time.Now()}
	out.Body.Data = data
	return out
}
