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

var (
	operatorRoles      = []string{"lite-nas-operator"}
	administratorRoles = []string{"admin", "sudo"}
	securityRoles      = []string{"lite-nas-security"}
)

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
	return RequireAnyRole(api, operatorRoles, administratorRoles)
}

// RequireAdministrator rejects authenticated callers that do not hold an
// administrator-equivalent role.
//
// Parameters:
//   - api: Huma API instance used to render transport-level auth errors
func RequireAdministrator(api huma.API) func(huma.Context, func(huma.Context)) {
	return RequireAnyRole(api, nil, administratorRoles)
}

// RequireSecurity rejects authenticated callers that do not hold the security
// role or an administrator-equivalent role.
//
// Parameters:
//   - api: Huma API instance used to render transport-level auth errors
func RequireSecurity(api huma.API) func(huma.Context, func(huma.Context)) {
	return RequireAnyRole(api, securityRoles, administratorRoles)
}

// RequireAnyRole rejects authenticated callers unless they hold at least one
// target role or one administrator-equivalent role.
//
// Parameters:
//   - api: Huma API instance used to render transport-level auth errors
//   - targetRoles: role names accepted for the target protected area
//   - elevatedRoles: administrator-equivalent role names accepted globally
func RequireAnyRole(
	api huma.API,
	targetRoles []string,
	elevatedRoles []string,
) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		claims, ok := accessClaimsFromContext(ctx.Context())
		if !ok {
			_ = huma.WriteErr(api, ctx, 401, "missing or invalid access token")
			return
		}

		if hasAnyRole(claims.Roles, targetRoles) || hasAnyRole(claims.Roles, elevatedRoles) {
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

func hasAnyRole(subjectRoles []string, acceptedRoles []string) bool {
	if len(acceptedRoles) == 0 {
		return false
	}

	roleSet := buildNormalizedRoleSet(subjectRoles)
	for _, role := range acceptedRoles {
		key := normalizeRole(role)
		if key == "" {
			continue
		}
		if _, ok := roleSet[key]; ok {
			return true
		}
	}

	return false
}

func buildNormalizedRoleSet(roles []string) map[string]struct{} {
	roleSet := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		key := normalizeRole(role)
		if key == "" {
			continue
		}
		roleSet[key] = struct{}{}
	}
	return roleSet
}

func normalizeRole(role string) string {
	return strings.ToLower(strings.TrimSpace(role))
}
