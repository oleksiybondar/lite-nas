package auth

import "time"

// ServiceTokenLoginRequest requests a long-lived service-to-service token pair.
type ServiceTokenLoginRequest struct {
	Service string `json:"service" validate:"required,min=1,max=128"`
}

// ServiceTokenLoginResponse returns issued service-to-service token material.
type ServiceTokenLoginResponse struct {
	AccessToken  string    `json:"access_token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
}
