package middlewares

import (
	"context"
	"strings"
	"time"

	"lite-nas/shared/roleauth"

	"github.com/danielgtaylor/huma/v2"

	"lite-nas/shared/authtoken"
	"lite-nas/shared/httpcookie"
)

type (
	accessTokenContextKey           struct{}
	accessClaimsContextKey          struct{}
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

		claims, err := options.Verifier.Verify(token)
		if err != nil {
			next(withAuthenticationFailure(ctx, err))
			return
		}

		next(withAuthenticatedContext(ctx, token, claims))
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

// RequireOperator rejects authenticated callers that do not hold the operator
// role or an administrator-equivalent role.
//
// Parameters:
//   - api: Huma API instance used to render transport-level auth errors
func RequireOperator(api huma.API) func(huma.Context, func(huma.Context)) {
	return RequireRole(api, roleauth.RequirementOperator)
}

// RequireAdministrator rejects authenticated callers that do not hold an
// administrator-equivalent role.
//
// Parameters:
//   - api: Huma API instance used to render transport-level auth errors
func RequireAdministrator(api huma.API) func(huma.Context, func(huma.Context)) {
	return RequireRole(api, roleauth.RequirementAdministrator)
}

// RequireSecurity rejects authenticated callers that do not hold the security
// role or an administrator-equivalent role.
//
// Parameters:
//   - api: Huma API instance used to render transport-level auth errors
func RequireSecurity(api huma.API) func(huma.Context, func(huma.Context)) {
	return RequireRole(api, roleauth.RequirementSecurity)
}

// RequireRole rejects authenticated callers unless they satisfy the shared
// coarse-grained authorization requirement.
//
// Parameters:
//   - api: Huma API instance used to render transport-level auth errors
//   - requirement: shared coarse authorization requirement to enforce
func RequireRole(
	api huma.API,
	requirement roleauth.Requirement,
) func(huma.Context, func(huma.Context)) {
	return RequireAnyRole(api, roleauth.AllowedRoles(requirement))
}

// RequireAnyRole rejects authenticated callers unless they hold at least one
// accepted role.
//
// Parameters:
//   - api: Huma API instance used to render transport-level auth errors
//   - acceptedRoles: role names accepted for the protected area
func RequireAnyRole(
	api huma.API,
	acceptedRoles []string,
) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		claims, ok := accessClaimsFromContext(ctx.Context())
		if !ok {
			_ = huma.WriteErr(api, ctx, 401, "missing or invalid access token")
			return
		}

		if roleauth.HasAnyRole(claims.Roles, acceptedRoles) {
			next(ctx)
			return
		}

		_ = huma.WriteErr(api, ctx, 403, "insufficient role")
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

// AccessTokenFromContext returns the access token previously extracted by
// authentication middleware when one is available.
func AccessTokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(accessTokenContextKey{}).(string)
	if !ok || !isAcceptedAccessToken(token) {
		return "", false
	}
	return token, true
}

// AccessClaimsFromContext returns the verified JWT claims previously extracted
// by authentication middleware when they are available.
func AccessClaimsFromContext(ctx context.Context) (authtoken.AccessClaims, bool) {
	return accessClaimsFromContext(ctx)
}

// NewAuthenticatedContext returns a plain context populated with the access
// token and verified claims that downstream handlers expect after successful
// authentication extraction.
func NewAuthenticatedContext(ctx context.Context, token string, claims authtoken.AccessClaims) context.Context {
	return context.WithValue(
		context.WithValue(ctx, accessTokenContextKey{}, token),
		accessClaimsContextKey{},
		claims,
	)
}

func accessClaimsFromContext(ctx context.Context) (authtoken.AccessClaims, bool) {
	claims, ok := ctx.Value(accessClaimsContextKey{}).(authtoken.AccessClaims)
	return claims, ok
}

func withAuthenticatedContext(
	ctx huma.Context,
	token string,
	claims authtoken.AccessClaims,
) huma.Context {
	return huma.WithValue(
		huma.WithValue(ctx, accessTokenContextKey{}, token),
		accessClaimsContextKey{},
		claims,
	)
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
