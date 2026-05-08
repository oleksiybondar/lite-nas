package auth

import (
	"net/http"
	"time"

	"lite-nas/services/web-gateway/dto"
)

// LogoutInput accepts auth token material from browser cookies.
type LogoutInput struct {
	Authorization      string `header:"Authorization" doc:"Bearer access token header for explicit REST-style clients."`
	AccessTokenCookie  string `cookie:"lite-nas-at" doc:"Access token cookie."`
	RefreshTokenCookie string `cookie:"lite-nas-rt" doc:"Refresh token cookie."`
	UserAgent          string `header:"User-Agent" doc:"Client user agent bound to the refresh session."`
	Body               LogoutRequestBody
}

// LogoutRequestBody is intentionally empty for the BFF browser auth flow.
// Refresh token material is read from the HTTP-only refresh-token cookie.
type LogoutRequestBody struct{}

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
