package httpcookie

import (
	"net/http"
	"time"
)

// Expired returns a secure HTTP-only cookie configured to delete an existing
// browser cookie with the provided name.
//
// Parameters:
//   - name: cookie name to expire
//   - now: clock value used to set an expiry timestamp in the past
func Expired(name string, now time.Time) http.Cookie {
	return http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  now.Add(-time.Hour),
		MaxAge:   -1,
	}
}
