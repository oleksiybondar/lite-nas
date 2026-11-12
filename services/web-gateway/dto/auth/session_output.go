package auth

import (
	"net/http"
	"time"

	"lite-nas/services/web-gateway/dto"
	authcontract "lite-nas/shared/contracts/auth"
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

// SessionData contains the current browser session state returned by
// session-issuing endpoints. Token material is transported only through
// HTTP-only cookies at the gateway boundary.
type SessionData struct {
	Authenticated bool                               `json:"authenticated"`
	User          AuthUser                           `json:"user"`
	RedirectTo    string                             `json:"redirect_to,omitempty"`
	Status        authcontract.Status                `json:"status,omitempty"`
	Conversation  *authcontract.PasswordConversation `json:"conversation,omitempty"`
	Messages      []authcontract.Message             `json:"messages,omitempty"`
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
