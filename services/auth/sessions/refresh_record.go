package sessions

import "time"

// RefreshRecord is active auth-service-owned state for one opaque refresh
// token. The store keeps only currently usable refresh tokens.
type RefreshRecord struct {
	TokenHash string
	ExpiresAt time.Time

	SessionID string
	Subject   string
	Login     string

	ClientIP  string
	UserAgent string
}

// RefreshToken is the raw token value returned to a caller with the absolute
// expiry inherited by its refresh session.
type RefreshToken struct {
	Value     string
	ExpiresAt time.Time
}

// RefreshContext binds a refresh token to observable request context.
type RefreshContext struct {
	ClientIP  string
	UserAgent string
}

// CreateInput contains identity, context, and lifetime data needed to create a
// refresh session.
type CreateInput struct {
	SessionID string
	Subject   string
	Login     string
	Context   RefreshContext
	TTL       time.Duration
}
