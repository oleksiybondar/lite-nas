package services

import "errors"

// ErrUnauthorized indicates that the caller did not provide a valid auth token.
var ErrUnauthorized = errors.New("unauthorized")

// ErrMissingRefreshToken indicates that a required refresh token was not
// provided to the auth service skeleton.
func ErrMissingRefreshToken() error {
	return errMissingRefreshToken
}

// ErrMissingCredentials indicates that required login credentials were not
// provided to the auth service skeleton.
func ErrMissingCredentials() error {
	return errMissingCredentials
}
