package authtoken

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type accessClaimOptions struct {
	Issuer    string
	Audience  string
	IssuedAt  time.Time
	ExpiresAt time.Time
	TokenID   string
	Principal Principal
}

func newAccessClaims(options accessClaimOptions) AccessClaims {
	return AccessClaims{
		RegisteredClaims: newRegisteredClaims(options),
		Login:            options.Principal.Login,
		Scopes:           options.Principal.Scopes,
		Roles:            options.Principal.Roles,
	}
}

func newRegisteredClaims(options accessClaimOptions) jwt.RegisteredClaims {
	return jwt.RegisteredClaims{
		Issuer:    options.Issuer,
		Subject:   options.Principal.Subject,
		Audience:  jwt.ClaimStrings{options.Audience},
		ExpiresAt: jwt.NewNumericDate(options.ExpiresAt),
		NotBefore: jwt.NewNumericDate(options.IssuedAt),
		IssuedAt:  jwt.NewNumericDate(options.IssuedAt),
		ID:        options.TokenID,
	}
}
