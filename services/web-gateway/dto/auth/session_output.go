package auth

import (
	"net/http"
	"time"

	"lite-nas/services/web-gateway/dto"
)

// AuthUser represents the authenticated user payload returned by auth
// endpoints.
type AuthUser struct {
	ID string `json:"id" example:"stub-user"`
}

// SessionOutput returns session metadata and auth cookies for login and
// refresh flows.
type SessionOutput struct {
	SetCookie []http.Cookie `header:"Set-Cookie"`
	Body      SessionBody
}

// SessionBody defines the browser-facing session envelope returned by login
// and refresh.
type SessionBody struct {
	dto.ResponseMeta
	Data SessionData `json:"data"`
}

// SessionData contains the current session state and token payload returned by
// session-issuing endpoints.
type SessionData struct {
	Authenticated bool     `json:"authenticated"`
	User          AuthUser `json:"user"`
	AccessToken   string   `json:"access_token"`
	RefreshToken  string   `json:"refresh_token"`
}

// NewSessionBody creates the session response body with common metadata set.
func NewSessionBody(now time.Time, data SessionData) SessionBody {
	return SessionBody{
		ResponseMeta: dto.ResponseMeta{
			Success:   true,
			Timestamp: now,
		},
		Data: data,
	}
}
