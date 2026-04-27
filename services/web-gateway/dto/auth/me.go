package auth

import (
	"time"

	"lite-nas/services/web-gateway/dto"
)

// MeInput reads the access token from the configured cookie.
type MeInput struct {
	AccessToken string `cookie:"lite-nas-at" required:"true" doc:"Access token cookie."`
}

// MeOutput returns the current authenticated user state.
type MeOutput struct {
	Body MeBody
}

// MeBody defines the browser-facing current-session response.
type MeBody struct {
	dto.ResponseMeta
	Data MeData `json:"data"`
}

// MeData contains the authenticated user state returned by /auth/me.
type MeData struct {
	Authenticated bool     `json:"authenticated"`
	AuthType      string   `json:"auth_type"`
	User          MeUser   `json:"user"`
	Roles         []string `json:"roles"`
	Scopes        []string `json:"scopes"`
}

// MeUser contains the current authenticated principal identity.
type MeUser struct {
	ID    string `json:"id" example:"stub-user"`
	Login string `json:"login" example:"john.doe"`
}

// NewMeBody creates the current-session response body with common metadata set.
func NewMeBody(now time.Time, data MeData) MeBody {
	return MeBody{
		ResponseMeta: dto.ResponseMeta{
			Success:   true,
			Timestamp: now,
		},
		Data: data,
	}
}
