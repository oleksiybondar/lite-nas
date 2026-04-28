package sessions

import "time"

// RefreshSession is auth-service-owned state for opaque refresh-token
// rotation. Refresh tokens are not JWTs and are not validated outside this
// service.
type RefreshSession struct {
	ID            string
	Subject       string
	Login         string
	TokenHash     []byte
	ExpiresAt     time.Time
	CreatedAt     time.Time
	LastRotatedAt time.Time
	RevokedAt     *time.Time
}
