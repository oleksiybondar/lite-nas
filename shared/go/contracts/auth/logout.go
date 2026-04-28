package auth

// LogoutRequest requests refresh-session revocation for the submitted refresh
// token.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// LogoutResponse confirms whether the submitted refresh session was revoked.
type LogoutResponse struct {
	LoggedOut bool `json:"logged_out"`
}
