package auth

// RefreshRequest requests a rotated access and refresh token pair for the
// submitted refresh token.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required,min=1,max=8192"`
	ClientIP     string `json:"client_ip,omitempty" validate:"omitempty,ip"`
	UserAgent    string `json:"user_agent,omitempty" validate:"omitempty,max=512"`
}

// RefreshResponse returns the rotated session material.
type RefreshResponse struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}
