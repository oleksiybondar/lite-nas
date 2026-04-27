package middlewares

import (
	"context"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

type accessTokenContextKey struct{}

// ExtractAuthentication extracts an access token from supported transports and
// stores it in the request context for downstream middleware and handlers.
//
// Parameters:
//   - cookieName: name of the HTTP-only access-token cookie to inspect
//
// Supported access-token transport policies:
//   - Authorization bearer header
//   - HTTP-only access-token cookie
//
// Intentional simplification:
//   - for now, token presence is treated as sufficient
//   - TODO: replace this with real JWT validation once token issuance is backed
//     by a real auth service
func ExtractAuthentication(cookieName string) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		token := extractAccessToken(ctx, cookieName)
		if token == "" {
			next(ctx)
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
func RequireAuthentication(api huma.API) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		if !hasAccessToken(ctx.Context()) {
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
