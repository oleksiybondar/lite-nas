package auth

// RefreshRequest requests a rotated access and refresh token pair for the
// submitted refresh token.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
	ClientIP     string `json:"client_ip,omitempty"`
	UserAgent    string `json:"user_agent,omitempty"`
}

// RefreshResponse returns the rotated session material.
type RefreshResponse struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}
