package cache

import "time"

// Entry stores a cached authorization decision and its absolute expiry time.
type Entry struct {
	Allowed   bool
	ExpiresAt time.Time
}

// IsValid reports whether the cached entry is still valid at the provided time.
func (entry Entry) IsValid(now time.Time) bool {
	return entry.ExpiresAt.After(now)
}
