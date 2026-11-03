// Package loganalysis defines the read-only Log Analysis rule catalog RPC.
package loganalysis

// RulesListSubject returns the configured detection rules.
const RulesListSubject = "log-analysis.rpc.rules.list"

// RulesListRequest carries the operator token for live authorization.
type RulesListRequest struct {
	AccessToken string `json:"access_token" validate:"required,min=1,max=8192"`
}

// LogAnalysisRule presents one configured detection and its operator guidance.
type LogAnalysisRule struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Explanation       string   `json:"explanation"`
	PossibleReactions []string `json:"possible_reactions"`
	Enabled           bool     `json:"enabled"`
	Source            string   `json:"source"`
	Category          string   `json:"category"`
	Severity          string   `json:"severity"`
	Instance          string   `json:"instance"`
	Threshold         int      `json:"threshold"`
	Window            string   `json:"window"`
	EmailOnCritical   bool     `json:"email_on_critical"`
	Lockdown          bool     `json:"lockdown"`
}

// LogAnalysisRulesResponse includes enabled and disabled configured rules.
type LogAnalysisRulesResponse struct {
	Rules []LogAnalysisRule `json:"rules"`
}
