package auth

// LogoutRequest requests refresh-session revocation for the submitted refresh
// token.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
	ClientIP     string `json:"client_ip,omitempty"`
	UserAgent    string `json:"user_agent,omitempty"`
}

// LogoutResponse confirms whether the submitted refresh session was revoked.
type LogoutResponse struct {
	LoggedOut bool `json:"logged_out"`
}
