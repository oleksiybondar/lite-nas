package auth

import "time"

// ServiceTokenRefreshRequest requests service-to-service token rotation.
type ServiceTokenRefreshRequest struct {
	Service      string `json:"service" validate:"required,min=1,max=128"`
	RefreshToken string `json:"refresh_token" validate:"required,min=1,max=8192"`
}

// ServiceTokenRefreshResponse returns rotated service-to-service token material.
type ServiceTokenRefreshResponse struct {
	AccessToken  string    `json:"access_token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
}
