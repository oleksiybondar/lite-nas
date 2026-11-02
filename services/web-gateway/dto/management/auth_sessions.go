package management

import (
	"time"

	"lite-nas/services/web-gateway/dto"
	authcontract "lite-nas/shared/contracts/auth"
)

// AuthSessionActionInput selects an authentication session from the inventory.
type AuthSessionActionInput struct {
	SessionID string `path:"session_id" minLength:"1" maxLength:"128" pattern:"^[A-Za-z0-9_-]+$" doc:"LiteNAS Auth session ID returned by the inventory."`
}

// AuthSession describes an active LiteNAS authentication session.
type AuthSession struct {
	SessionID string    `json:"session_id"`
	Subject   string    `json:"subject"`
	Login     string    `json:"login"`
	ExpiresAt time.Time `json:"expires_at"`
}

// AuthSessionsOutput returns the current LiteNAS authentication inventory.
type AuthSessionsOutput struct{ Body AuthSessionsBody }

// AuthSessionsBody contains session identities and common response metadata.
type AuthSessionsBody struct {
	dto.ResponseMeta
	Data []AuthSession `json:"data"`
}

// AuthSessionRevokeOutput reports whether a session was revoked.
type AuthSessionRevokeOutput struct{ Body AuthSessionRevokeBody }

// AuthSessionRevokeBody contains the revoke result and common response metadata.
type AuthSessionRevokeBody struct {
	dto.ResponseMeta
	Revoked bool `json:"revoked"`
}

// NewAuthSessionsBody maps Auth Service results into browser-facing data.
func NewAuthSessionsBody(items []authcontract.SessionSummary) AuthSessionsBody {
	data := make([]AuthSession, 0, len(items))
	for _, item := range items {
		data = append(data, AuthSession{SessionID: item.SessionID, Subject: item.Subject, Login: item.Login, ExpiresAt: item.ExpiresAt})
	}
	return AuthSessionsBody{ResponseMeta: dto.ResponseMeta{Success: true, Timestamp: time.Now()}, Data: data}
}

// NewAuthSessionRevokeBody wraps the backend's revocation result.
func NewAuthSessionRevokeBody(revoked bool) AuthSessionRevokeBody {
	return AuthSessionRevokeBody{ResponseMeta: dto.ResponseMeta{Success: true, Timestamp: time.Now()}, Revoked: revoked}
}
