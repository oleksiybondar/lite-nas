package middlewares

import (
	"context"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"lite-nas/shared/authtoken"
	"lite-nas/shared/httpcookie"
)

type (
	accessTokenContextKey           struct{}
	authenticationFailureContextKey struct{}
)

// AccessTokenVerifier defines the local JWT verification behavior needed by
// authentication middleware.
type AccessTokenVerifier interface {
	Verify(tokenText string) (authtoken.AccessClaims, error)
}

// AuthenticationOptions configures gateway authentication extraction and
// enforcement middleware.
type AuthenticationOptions struct {
	AccessCookieName  string
	RefreshCookieName string
	Verifier          AccessTokenVerifier
}

type authenticationFailure string

const authenticationFailureExpired authenticationFailure = "expired"

// ExtractAuthentication extracts an access token from supported transports and
// stores it in the request context for downstream middleware and handlers.
//
// Parameters:
//   - options: cookie names and JWT verifier used by the middleware
//
// Supported access-token transport policies:
//   - Authorization bearer header
//   - HTTP-only access-token cookie
func ExtractAuthentication(options AuthenticationOptions) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		token := extractAccessToken(ctx, options.AccessCookieName)
		if token == "" {
			next(ctx)
			return
		}

		if options.Verifier == nil {
			next(ctx)
			return
		}

		if _, err := options.Verifier.Verify(token); err != nil {
			next(withAuthenticationFailure(ctx, err))
			return
		}

		next(huma.WithValue(ctx, accessTokenContextKey{}, token))
	}
}

// RequireAuthentication rejects protected endpoints unless an access token has
// already been extracted into the request context.
//
// Parameters:
//   - api: Huma API instance used to render transport-level auth errors
//   - options: cookie names used to clear expired browser auth sessions
func RequireAuthentication(api huma.API, options AuthenticationOptions) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		if !hasAccessToken(ctx.Context()) {
			if hasExpiredAuthentication(ctx.Context()) {
				clearAuthCookies(ctx, options, time.Now())
			}

			_ = huma.WriteErr(api, ctx, 401, "missing or invalid access token")
			return
		}

		next(ctx)
	}
}

func extractAccessToken(ctx huma.Context, cookieName string) string {
	headerToken := extractBearerToken(ctx.Header("Authorization"))
	if isAcceptedAccessToken(headerToken) {
		return headerToken
	}

	cookie, err := huma.ReadCookie(ctx, cookieName)
	if err != nil {
		return ""
	}

	cookieToken := strings.TrimSpace(cookie.Value)
	if isAcceptedAccessToken(cookieToken) {
		return cookieToken
	}

	return ""
}

func extractBearerToken(header string) string {
	const bearerPrefix = "Bearer "

	if !strings.HasPrefix(header, bearerPrefix) {
		return ""
	}

	return strings.TrimSpace(strings.TrimPrefix(header, bearerPrefix))
}

func isAcceptedAccessToken(token string) bool {
	return strings.TrimSpace(token) != ""
}

func hasAccessToken(ctx context.Context) bool {
	token, _ := ctx.Value(accessTokenContextKey{}).(string)
	return isAcceptedAccessToken(token)
}

func withAuthenticationFailure(ctx huma.Context, err error) huma.Context {
	if authtoken.IsExpiredError(err) {
		return huma.WithValue(ctx, authenticationFailureContextKey{}, authenticationFailureExpired)
	}

	return ctx
}

func hasExpiredAuthentication(ctx context.Context) bool {
	failure, _ := ctx.Value(authenticationFailureContextKey{}).(authenticationFailure)
	return failure == authenticationFailureExpired
}

func clearAuthCookies(ctx huma.Context, options AuthenticationOptions, now time.Time) {
	accessCookie := httpcookie.Expired(options.AccessCookieName, now)
	refreshCookie := httpcookie.Expired(options.RefreshCookieName, now)
	ctx.AppendHeader("Set-Cookie", accessCookie.String())
	ctx.AppendHeader("Set-Cookie", refreshCookie.String())
}
