package configtest

import (
	"time"

	sharedconfig "lite-nas/shared/config"
	"lite-nas/shared/testutil/testcasetest"
)

// AuthTokenFieldCases returns common field assertions for a loaded config value
// that exposes shared auth-token settings.
func AuthTokenFieldCases[T any](
	getAuthTokens func(T) sharedconfig.AuthTokenConfig,
) []testcasetest.FieldCase[T] {
	return []testcasetest.FieldCase[T]{
		{Name: "issuer", Got: func(cfg T) any { return getAuthTokens(cfg).Issuer }, Want: "lite-nas-auth"},
		{Name: "audience", Got: func(cfg T) any { return getAuthTokens(cfg).Audience }, Want: "lite-nas-services"},
		{Name: "access lifetime", Got: func(cfg T) any { return getAuthTokens(cfg).AccessLifetime }, Want: 15 * time.Minute},
		{Name: "clock skew", Got: func(cfg T) any { return getAuthTokens(cfg).ClockSkew }, Want: 30 * time.Second},
	}
}
