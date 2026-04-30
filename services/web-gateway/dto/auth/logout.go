package auth

import (
	"net/http"
	"time"

	"lite-nas/services/web-gateway/dto"
)

// LogoutInput accepts auth token material from explicit payload fields for
// non-cookie clients.
type LogoutInput struct {
	Authorization      string `header:"Authorization" doc:"Bearer access token header for explicit REST-style clients."`
	AccessTokenCookie  string `cookie:"lite-nas-at" doc:"Access token cookie."`
	RefreshTokenCookie string `cookie:"lite-nas-rt" doc:"Refresh token cookie."`
	UserAgent          string `header:"User-Agent" doc:"Client user agent bound to the refresh session."`
	Body               LogoutRequestBody
}

// LogoutRequestBody defines the explicit payload-based logout transport.
type LogoutRequestBody struct {
	AccessToken  string `json:"access_token,omitempty" pattern:"^AT-[A-Za-z0-9-]+$" doc:"Explicit access token payload for non-cookie clients."`
	RefreshToken string `json:"refresh_token,omitempty" pattern:"^RT-[A-Za-z0-9-]+$" doc:"Explicit refresh token payload for non-cookie clients."`
}

// LogoutOutput clears auth cookies and reports logout completion.
type LogoutOutput struct {
	SetCookie []http.Cookie `header:"Set-Cookie"`
	Body      LogoutBody
}

// LogoutBody defines the browser-facing logout response.
type LogoutBody struct {
	dto.ResponseMeta
}

// NewLogoutBody creates a logout response body with the common metadata set.
func NewLogoutBody(now time.Time, success bool, message string) LogoutBody {
	return LogoutBody{
		ResponseMeta: dto.ResponseMeta{
			Success:   success,
			Timestamp: now,
			Message:   message,
		},
	}
}
