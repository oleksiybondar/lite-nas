package main

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"lite-nas/shared/browserauth"
)

// authController connects Moments routes to the shared browser auth lifecycle.
type authController struct {
	service      browserauth.AuthService
	cookieDomain string
}

// loginInput validates submitted credentials at the HTTP boundary.
type loginInput struct {
	UserAgent string `header:"User-Agent"`
	Body      struct {
		Login    string `json:"login" required:"true" pattern:"^(?:[A-Za-z0-9._-]+|[^@\\s]+@[^@\\s]+\\.[^@\\s]+)$"`
		Password string `json:"password" required:"true" minLength:"4"`
	}
}

// refreshInput reads the HTTP-only refresh cookie.
type refreshInput struct {
	RefreshToken string `cookie:"lite-nas-rt"`
	UserAgent    string `header:"User-Agent"`
	Body         struct{}
}

// meInput accepts the access cookie or an explicit bearer token.
type meInput struct {
	Authorization string `header:"Authorization"`
	AccessToken   string `cookie:"lite-nas-at"`
}

// responseMeta is the browser auth response metadata shared with the gateway.
type responseMeta struct {
	Success   bool      `json:"success"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message,omitempty"`
}

// sessionOutput issues access and refresh cookies after login or rotation.
type sessionOutput struct {
	SetCookie []http.Cookie `header:"Set-Cookie"`
	Body      sessionBody
}

// sessionBody carries the gateway-compatible session envelope.
type sessionBody struct {
	responseMeta
	Data sessionData `json:"data"`
}

// sessionData reports the issued session without exposing token material.
type sessionData struct {
	Authenticated bool        `json:"authenticated"`
	User          sessionUser `json:"user"`
}

// sessionUser identifies the authenticated principal.
type sessionUser struct {
	ID string `json:"id"`
}

// logoutOutput clears cookies after the refresh token is revoked.
type logoutOutput struct {
	SetCookie []http.Cookie `header:"Set-Cookie"`
	Body      responseMeta
}

// meOutput returns the verified principal to the browser.
type meOutput struct{ Body meBody }

// meBody carries the current-user response envelope.
type meBody struct {
	responseMeta
	Data meData `json:"data"`
}

// meData reports the verified user's identity and permissions.
type meData struct {
	Authenticated bool     `json:"authenticated"`
	AuthType      string   `json:"auth_type"`
	User          meUser   `json:"user"`
	Roles         []string `json:"roles"`
	Scopes        []string `json:"scopes"`
}

// meUser reports the current principal identity.
type meUser struct {
	ID    string `json:"id"`
	Login string `json:"login"`
}

// login delegates credential verification to Auth Service.
func (c authController) login(ctx context.Context, input *loginInput) (*sessionOutput, error) {
	session, err := c.service.Login(ctx, time.Now(), input.Body.Login, input.Body.Password, browserauth.AuthRequestContext{UserAgent: strings.TrimSpace(input.UserAgent)})
	if err != nil {
		return nil, authError(err)
	}
	return c.sessionResponse(session), nil
}

// refresh rotates the cookie session through Auth Service.
func (c authController) refresh(ctx context.Context, input *refreshInput) (*sessionOutput, error) {
	if strings.TrimSpace(input.RefreshToken) == "" {
		return nil, huma.Error401Unauthorized("missing refresh token")
	}
	session, err := c.service.Refresh(ctx, time.Now(), input.RefreshToken, browserauth.AuthRequestContext{UserAgent: strings.TrimSpace(input.UserAgent)})
	if err != nil {
		return nil, authError(err)
	}
	return c.sessionResponse(session), nil
}

// logout revokes the refresh token and expires the issued cookie scope.
func (c authController) logout(ctx context.Context, input *refreshInput) (*logoutOutput, error) {
	if strings.TrimSpace(input.RefreshToken) == "" {
		return nil, huma.Error401Unauthorized("missing refresh token")
	}
	session, err := c.service.Logout(ctx, time.Now(), input.RefreshToken, browserauth.AuthRequestContext{UserAgent: strings.TrimSpace(input.UserAgent)})
	if err != nil {
		return nil, authError(err)
	}
	return &logoutOutput{SetCookie: []http.Cookie{browserauth.ClearAccessCookie(session.AccessExpires, c.cookieDomain), browserauth.ClearRefreshCookie(session.RefreshExpiry, c.cookieDomain)}, Body: responseMeta{Success: true, Timestamp: time.Now(), Message: "logged out"}}, nil
}

// me returns the verified user profile from the shared auth client.
func (c authController) me(_ context.Context, input *meInput) (*meOutput, error) {
	session, err := c.service.Me(time.Now(), accessToken(input.Authorization, input.AccessToken))
	if err != nil {
		return nil, authError(err)
	}
	output := &meOutput{}
	output.Body.responseMeta = responseMeta{Success: true, Timestamp: time.Now()}
	output.Body.Data.Authenticated = true
	output.Body.Data.AuthType = session.AuthType
	output.Body.Data.User.ID = session.UserID
	output.Body.Data.User.Login = session.Login
	output.Body.Data.Roles = session.Roles
	output.Body.Data.Scopes = session.Scopes
	return output, nil
}

// sessionResponse shapes the gateway-compatible login and refresh response.
func (c authController) sessionResponse(session browserauth.Session) *sessionOutput {
	output := &sessionOutput{SetCookie: []http.Cookie{browserauth.AccessCookie(session, c.cookieDomain), browserauth.RefreshCookie(session, c.cookieDomain)}}
	output.Body.responseMeta = responseMeta{Success: true, Timestamp: time.Now()}
	output.Body.Data.Authenticated = true
	output.Body.Data.User.ID = session.UserID
	return output
}

// accessToken applies the gateway's header-before-cookie precedence.
func accessToken(authorization, cookie string) string {
	if strings.HasPrefix(authorization, "Bearer ") {
		if token := strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer ")); token != "" {
			return token
		}
	}
	return strings.TrimSpace(cookie)
}

// authError maps shared auth failures to browser-facing HTTP statuses.
func authError(err error) error {
	if errors.Is(err, browserauth.ErrUnauthorized) || errors.Is(err, browserauth.ErrMissingCredentials()) || errors.Is(err, browserauth.ErrMissingRefreshToken()) {
		return huma.Error401Unauthorized("invalid authentication")
	}
	return huma.Error500InternalServerError("failed to process auth token")
}
