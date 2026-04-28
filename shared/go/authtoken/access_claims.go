package authtoken

import "github.com/golang-jwt/jwt/v5"

// AccessClaims is the signed JWT payload shared between the auth service and
// services that validate LiteNAS access tokens.
type AccessClaims struct {
	jwt.RegisteredClaims

	Login  string   `json:"login"`
	Scopes []string `json:"scopes,omitempty"`
	Roles  []string `json:"roles,omitempty"`
}
