package configtest

import (
	"time"

	sharedconfig "lite-nas/shared/config"
	"lite-nas/shared/testutil/testcasetest"
)

// AuthTokenExpectedPaths describes service-specific certificate and key paths
// expected from an auth-token config fixture.
type AuthTokenExpectedPaths struct {
	SigningKey       string
	SigningCert      string
	VerificationCert string
}

// AuthTokenFieldCases returns common field assertions for a loaded config value
// that exposes shared auth-token settings.
func AuthTokenFieldCases[T any](
	getAuthTokens func(T) sharedconfig.AuthTokenConfig,
	expectedPaths AuthTokenExpectedPaths,
) []testcasetest.FieldCase[T] {
	return []testcasetest.FieldCase[T]{
		{Name: "issuer", Got: func(cfg T) any { return getAuthTokens(cfg).Issuer }, Want: "lite-nas-auth"},
		{Name: "audience", Got: func(cfg T) any { return getAuthTokens(cfg).Audience }, Want: "lite-nas-management-api"},
		{Name: "access lifetime", Got: func(cfg T) any { return getAuthTokens(cfg).AccessLifetime }, Want: 15 * time.Minute},
		{Name: "clock skew", Got: func(cfg T) any { return getAuthTokens(cfg).ClockSkew }, Want: 30 * time.Second},
		{Name: "signing key", Got: func(cfg T) any { return getAuthTokens(cfg).SigningKey }, Want: expectedPaths.SigningKey},
		{Name: "signing cert", Got: func(cfg T) any { return getAuthTokens(cfg).SigningCert }, Want: expectedPaths.SigningCert},
		{Name: "verification cert", Got: func(cfg T) any { return getAuthTokens(cfg).VerificationCert }, Want: expectedPaths.VerificationCert},
		{Name: "enforce refresh client IP", Got: func(cfg T) any { return getAuthTokens(cfg).EnforceRefreshClientIP }, Want: false},
	}
}
