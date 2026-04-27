package controllers

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	authdto "lite-nas/services/web-gateway/dto/auth"
	"lite-nas/services/web-gateway/services"

	"github.com/danielgtaylor/huma/v2"
)

// AuthService defines the auth behavior required by the browser-facing auth
// controller.
type AuthService interface {
	Login(now time.Time, login string, password string) (services.Session, error)
	Refresh(now time.Time, refreshToken string) (services.Session, error)
	Logout(now time.Time, refreshToken string) (services.Session, error)
	Me(now time.Time, accessToken string) (services.Session, error)
}

// AuthController translates auth HTTP requests into service calls and shapes
// the browser-facing responses.
type AuthController struct {
	service AuthService
}

// NewAuthController creates an AuthController.
//
// Parameters:
//   - service: auth service implementation used to execute auth flows
func NewAuthController(service AuthService) AuthController {
	return AuthController{service: service}
}

// Login authenticates the submitted credentials and returns session metadata
// plus browser cookies for the stub auth flow.
//
// Parameters:
//   - input: validated login payload containing the submitted credentials
func (c AuthController) Login(
	_ context.Context,
	input *authdto.LoginInput,
) (*authdto.SessionOutput, error) {
	now := time.Now()
	session, err := c.service.Login(now, input.Body.Login, input.Body.Password)
	if err != nil {
		return nil, huma.Error401Unauthorized("invalid login or password")
	}

	return &authdto.SessionOutput{
		SetCookie: []http.Cookie{
			services.AccessCookie(session),
			services.RefreshCookie(session),
		},
		Body: buildSessionBody(now, session),
	}, nil
}

// Refresh rotates the auth token pair using a refresh token from the payload
// or cookie transport.
//
// Parameters:
//   - input: validated refresh request plus any extracted refresh-token cookie
func (c AuthController) Refresh(
	_ context.Context,
	input *authdto.RefreshInput,
) (*authdto.SessionOutput, error) {
	now := time.Now()
	refreshToken := resolveRefreshToken(input)
	if strings.TrimSpace(refreshToken) == "" {
		return nil, huma.Error401Unauthorized("missing refresh token")
	}

	session, err := c.service.Refresh(now, refreshToken)
	if err != nil {
		return nil, mapAuthError(err, "invalid refresh token")
	}

	return &authdto.SessionOutput{
		SetCookie: []http.Cookie{
			services.AccessCookie(session),
			services.RefreshCookie(session),
		},
		Body: buildSessionBody(now, session),
	}, nil
}

// Logout expires the browser auth cookies when a refresh token is present.
//
// Parameters:
//   - input: validated logout request plus any extracted refresh-token cookie
func (c AuthController) Logout(
	_ context.Context,
	input *authdto.LogoutInput,
) (*authdto.LogoutOutput, error) {
	now := time.Now()
	refreshToken := resolveLogoutRefreshToken(input)
	if strings.TrimSpace(refreshToken) == "" {
		return nil, huma.Error401Unauthorized("missing refresh token")
	}

	session, err := c.service.Logout(now, refreshToken)
	if err != nil {
		return nil, mapAuthError(err, "invalid refresh token")
	}

	return &authdto.LogoutOutput{
		SetCookie: []http.Cookie{
			services.ClearAccessCookie(session.AccessExpires),
			services.ClearRefreshCookie(session.RefreshExpiry),
		},
		Body: authdto.NewLogoutBody(now, true, "logged out"),
	}, nil
}

// Me returns the currently authenticated stub user represented by the access
// token extracted by middleware.
//
// Parameters:
//   - input: request context populated with the resolved access token
func (c AuthController) Me(
	_ context.Context,
	input *authdto.MeInput,
) (*authdto.MeOutput, error) {
	now := time.Now()
	session, err := c.service.Me(now, input.AccessToken)
	if err != nil {
		return nil, mapAuthError(err, "missing or invalid access token")
	}

	return &authdto.MeOutput{
		Body: authdto.NewMeBody(now, authdto.MeData{
			Authenticated: true,
			AuthType:      session.AuthType,
			User: authdto.MeUser{
				ID:    session.UserID,
				Login: session.Login,
			},
			Roles:  session.Roles,
			Scopes: session.Scopes,
		}),
	}, nil
}

func buildSessionBody(now time.Time, session services.Session) authdto.SessionBody {
	return authdto.NewSessionBody(now, authdto.SessionData{
		Authenticated: true,
		User: authdto.AuthUser{
			ID: session.UserID,
		},
		AccessToken:  session.AccessToken,
		RefreshToken: session.RefreshToken,
	})
}

func resolveRefreshToken(input *authdto.RefreshInput) string {
	if input == nil {
		return ""
	}

	if token := strings.TrimSpace(input.Body.RefreshToken); token != "" {
		return token
	}

	return strings.TrimSpace(input.RefreshTokenCookie)
}

func mapAuthError(err error, message string) error {
	if errors.Is(err, services.ErrUnauthorized) ||
		errors.Is(err, services.ErrMissingRefreshToken()) ||
		errors.Is(err, services.ErrMissingCredentials()) {
		return huma.Error401Unauthorized(message)
	}

	return huma.Error500InternalServerError("failed to process auth token stub")
}

func resolveLogoutRefreshToken(input *authdto.LogoutInput) string {
	if input == nil {
		return ""
	}

	if token := strings.TrimSpace(input.Body.RefreshToken); token != "" {
		return token
	}

	return strings.TrimSpace(input.RefreshTokenCookie)
}
