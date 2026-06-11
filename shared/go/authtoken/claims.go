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
	scopes := options.Principal.Scopes
	if scopes == nil {
		scopes = []string{}
	}

	roles := options.Principal.Roles
	if roles == nil {
		roles = []string{}
	}

	return AccessClaims{
		RegisteredClaims: newRegisteredClaims(options),
		Login:            options.Principal.Login,
		Scopes:           scopes,
		Roles:            roles,
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
