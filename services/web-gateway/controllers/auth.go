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

	"lite-nas/shared/browserauth"
	authcontract "lite-nas/shared/contracts/auth"
)

// AuthService defines the auth behavior required by the browser-facing auth
// controller.
type AuthService interface {
	Login(context.Context, time.Time, string, string, services.AuthRequestContext) (services.Session, error)
	BeginLogin(context.Context, time.Time, string, string, services.AuthRequestContext) (browserauth.PasswordResult, error)
	BeginPasswordChange(context.Context, string) (browserauth.PasswordResult, error)
	ReplyPasswordChange(context.Context, time.Time, string, int32, string) (browserauth.PasswordResult, error)
	Refresh(context.Context, time.Time, string, services.AuthRequestContext) (services.Session, error)
	Logout(context.Context, time.Time, string, services.AuthRequestContext) (services.Session, error)
	Me(now time.Time, accessToken string) (services.Session, error)
}

// AuthController translates auth HTTP requests into service calls and shapes
// the browser-facing responses.
type AuthController struct {
	service      AuthService
	cookieDomain string
}

// NewAuthController creates an AuthController.
//
// Parameters:
//   - service: auth service implementation used to execute auth flows
func NewAuthController(service AuthService, cookieDomain ...string) AuthController {
	controller := AuthController{service: service}
	if len(cookieDomain) > 0 {
		controller.cookieDomain = cookieDomain[0]
	}
	return controller
}

// Login authenticates the submitted credentials and returns session metadata
// plus browser cookies for the stub auth flow.
//
// Parameters:
//   - input: validated login payload containing the submitted credentials
func (c AuthController) Login(
	ctx context.Context,
	input *authdto.LoginInput,
) (*authdto.SessionOutput, error) {
	now := time.Now()
	result, err := c.service.BeginLogin(ctx, now, input.Body.Login, input.Body.Password, authRequestContext(input.UserAgent))
	if err != nil {
		return nil, mapAuthError(err, "invalid login or password")
	}
	if result.Status == authcontract.StatusDenied {
		return nil, huma.Error401Unauthorized("invalid login or password")
	}
	return c.passwordOutput(now, result, browserauth.RedirectPath(input.RedirectTo)), nil
}

// StartPasswordChange begins a PAM conversation for the signed-in principal.
func (c AuthController) StartPasswordChange(ctx context.Context, input *authdto.PasswordChangeStartInput) (*authdto.SessionOutput, error) {
	token := resolveAccessToken(input.Authorization, input.AccessTokenCookie)
	result, err := c.service.BeginPasswordChange(ctx, token)
	if err != nil {
		return nil, mapAuthError(err, "password change failed")
	}
	return c.passwordOutput(time.Now(), result, ""), nil
}

// ReplyPasswordChange forwards one answer to the pending PAM conversation.
func (c AuthController) ReplyPasswordChange(ctx context.Context, input *authdto.PasswordChangeReplyInput) (*authdto.SessionOutput, error) {
	result, err := c.service.ReplyPasswordChange(ctx, time.Now(), input.Body.ConversationID, input.Body.Sequence, input.Body.Answer)
	if err != nil {
		return nil, mapAuthError(err, "password change failed")
	}
	return c.passwordOutput(time.Now(), result, ""), nil
}

// passwordOutput returns PAM state and only sets cookies for a new session.
func (c AuthController) passwordOutput(now time.Time, result browserauth.PasswordResult, redirectTo string) *authdto.SessionOutput {
	output := &authdto.SessionOutput{Body: authdto.NewSessionBody(now, authdto.SessionData{
		Authenticated: result.Status == authcontract.StatusAuthenticated, User: authdto.AuthUser{ID: result.Session.UserID}, RedirectTo: redirectTo,
		Status: result.Status, Conversation: result.Conversation, Messages: result.Messages,
	})}
	output.Body.Success = result.Status != authcontract.StatusDenied
	if result.Session.AccessToken != "" {
		output.SetCookie = []http.Cookie{browserauth.AccessCookie(result.Session, c.cookieDomain), browserauth.RefreshCookie(result.Session, c.cookieDomain)}
	}
	return output
}

// Refresh rotates the auth token pair using the HTTP-only refresh-token cookie.
//
// Parameters:
//   - input: validated refresh request plus the extracted refresh-token cookie
func (c AuthController) Refresh(
	ctx context.Context,
	input *authdto.RefreshInput,
) (*authdto.SessionOutput, error) {
	now := time.Now()
	refreshToken := resolveRefreshToken(input)
	if strings.TrimSpace(refreshToken) == "" {
		return nil, huma.Error401Unauthorized("missing refresh token")
	}

	session, err := c.service.Refresh(ctx, now, refreshToken, authRequestContext(input.UserAgent))
	if err != nil {
		return nil, mapAuthError(err, "invalid refresh token")
	}

	return &authdto.SessionOutput{
		SetCookie: []http.Cookie{
			browserauth.AccessCookie(session, c.cookieDomain),
			browserauth.RefreshCookie(session, c.cookieDomain),
		},
		Body: buildSessionBody(now, session, ""),
	}, nil
}

// Logout revokes an active refresh session and always expires browser auth cookies.
//
// Parameters:
//   - input: validated logout request plus any extracted refresh-token cookie
func (c AuthController) Logout(
	ctx context.Context,
	input *authdto.LogoutInput,
) (*authdto.LogoutOutput, error) {
	now := time.Now()
	refreshToken := resolveLogoutRefreshToken(input)
	if strings.TrimSpace(refreshToken) == "" {
		return c.logoutOutput(now), nil
	}

	_, err := c.service.Logout(ctx, now, refreshToken, authRequestContext(input.UserAgent))
	if err != nil {
		if errors.Is(err, services.ErrUnauthorized) || errors.Is(err, services.ErrMissingRefreshToken()) {
			return c.logoutOutput(now), nil
		}
		return nil, mapAuthError(err, "invalid refresh token")
	}
	return c.logoutOutput(now), nil
}

// logoutOutput clears both cookies in the configured browser scope.
func (c AuthController) logoutOutput(now time.Time) *authdto.LogoutOutput {
	return &authdto.LogoutOutput{
		SetCookie: []http.Cookie{
			browserauth.ClearAccessCookie(now, c.cookieDomain),
			browserauth.ClearRefreshCookie(now, c.cookieDomain),
		},
		Body: authdto.NewLogoutBody(now, true, "logged out"),
	}
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
	session, err := c.service.Me(now, resolveAccessToken(input.Authorization, input.AccessTokenCookie))
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

func buildSessionBody(now time.Time, session services.Session, redirectTo string) authdto.SessionBody {
	return authdto.NewSessionBody(now, authdto.SessionData{
		Authenticated: true,
		User: authdto.AuthUser{
			ID: session.UserID,
		},
		RedirectTo: redirectTo,
	})
}

func authRequestContext(userAgent string) services.AuthRequestContext {
	return services.AuthRequestContext{UserAgent: strings.TrimSpace(userAgent)}
}

func resolveAccessToken(authorization string, cookieValue string) string {
	if token := extractBearerToken(authorization); token != "" {
		return token
	}

	return strings.TrimSpace(cookieValue)
}

func resolveRefreshToken(input *authdto.RefreshInput) string {
	if input == nil {
		return ""
	}

	return strings.TrimSpace(input.RefreshTokenCookie)
}

func mapAuthError(err error, message string) error {
	if errors.Is(err, services.ErrUnauthorized) ||
		errors.Is(err, services.ErrMissingRefreshToken()) ||
		errors.Is(err, services.ErrMissingCredentials()) {
		return huma.Error401Unauthorized(message)
	}

	return huma.Error500InternalServerError("failed to process auth token")
}

func resolveLogoutRefreshToken(input *authdto.LogoutInput) string {
	if input == nil {
		return ""
	}

	return strings.TrimSpace(input.RefreshTokenCookie)
}

func extractBearerToken(header string) string {
	const bearerPrefix = "Bearer "

	if !strings.HasPrefix(header, bearerPrefix) {
		return ""
	}

	return strings.TrimSpace(strings.TrimPrefix(header, bearerPrefix))
}
