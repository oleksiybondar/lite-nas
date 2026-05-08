package sessions

import "errors"

var (
	ErrUnknownRefreshToken         = errors.New("unknown refresh token")
	ErrExpiredRefreshToken         = errors.New("expired refresh token")
	ErrRefreshTokenContextMismatch = errors.New("refresh token context mismatch")
	ErrInvalidRefreshSession       = errors.New("invalid refresh session")
)
