package auth

// LogoutRequest requests refresh-session revocation for the submitted refresh
// token.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required,min=1,max=8192"`
	ClientIP     string `json:"client_ip,omitempty" validate:"omitempty,ip"`
	UserAgent    string `json:"user_agent,omitempty" validate:"omitempty,max=512"`
}

// LogoutResponse confirms whether the submitted refresh session was revoked.
type LogoutResponse struct {
	LoggedOut bool `json:"logged_out"`
}
