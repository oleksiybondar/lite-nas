package auth

import "time"

// ListSessionsRequest supplies the administrator credential for session inventory.
type ListSessionsRequest struct {
	AccessToken string `json:"access_token" validate:"required,min=1,max=8192"`
}

// SessionSummary identifies one active LiteNAS login session.
type SessionSummary struct {
	SessionID string    `json:"session_id"`
	Subject   string    `json:"subject"`
	Login     string    `json:"login"`
	Source    string    `json:"source,omitempty"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ListSessionsResponse returns active LiteNAS login sessions.
type ListSessionsResponse struct {
	Sessions []SessionSummary `json:"sessions"`
}
