package auth

// RevokeSessionRequest identifies a login session and an administrator credential.
type RevokeSessionRequest struct {
	AccessToken string `json:"access_token" validate:"required,min=1,max=8192"`
	SessionID   string `json:"session_id" validate:"required,min=1,max=128"`
}

// RevokeSessionResponse confirms whether an active refresh session was revoked.
type RevokeSessionResponse struct {
	Revoked bool `json:"revoked"`
}
