package auth

// LoginRequest requests host-backed authentication for the submitted user
// credentials.
type LoginRequest struct {
	Username  string `json:"username" validate:"required,min=1,max=128"`
	Password  string `json:"password" validate:"required,min=1,max=4096"`
	ClientIP  string `json:"client_ip,omitempty" validate:"omitempty,ip"`
	UserAgent string `json:"user_agent,omitempty" validate:"omitempty,max=512"`
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
