package auth

// LoginRequest requests host-backed authentication for the submitted user
// credentials.
type LoginRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	ClientIP  string `json:"client_ip,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// LoginResponse returns the normalized auth outcome and issued session
// material when authentication succeeds.
type LoginResponse struct {
	Status            Status    `json:"status"`
	Username          string    `json:"username"`
	Messages          []Message `json:"messages,omitempty"`
	CanChangePassword bool      `json:"can_change_password,omitempty"`
	AccessToken       string    `json:"access_token,omitempty"`
	RefreshToken      string    `json:"refresh_token,omitempty"`
}
