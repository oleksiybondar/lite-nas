package authtoken

import "github.com/golang-jwt/jwt/v5"

// AccessClaims is the signed JWT payload shared between the auth service and
// services that validate LiteNAS access tokens.
//
// RegisteredClaims carries the standard JWT fields used for bounded trust:
// issuer, subject, audience, expiry, not-before, issued-at, and token ID.
//
// Login is the authenticated host login name shown to users and services.
// Scopes and Roles are optional authorization labels; an empty set means the
// token proves identity only and does not grant any service capability by
// itself.
type AccessClaims struct {
	jwt.RegisteredClaims

	// Login is the authenticated host login name associated with Subject.
	Login string `json:"login"`
	// Scopes contains optional fine-grained permissions such as
	// "monitoring.read".
	Scopes []string `json:"scopes,omitempty"`
	// Roles contains optional coarse authorization labels such as "operator".
	Roles []string `json:"roles,omitempty"`
}
