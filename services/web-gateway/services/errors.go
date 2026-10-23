package services

import "lite-nas/shared/browserauth"

// ErrUnauthorized indicates that a caller did not provide a valid auth token.
var ErrUnauthorized = browserauth.ErrUnauthorized

// ErrMissingRefreshToken identifies an absent refresh credential.
func ErrMissingRefreshToken() error { return browserauth.ErrMissingRefreshToken() }

// ErrMissingCredentials identifies absent login credentials.
func ErrMissingCredentials() error { return browserauth.ErrMissingCredentials() }
