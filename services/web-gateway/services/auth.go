package services

import (
	"net/http"
	"time"

	"lite-nas/shared/authtoken"
	"lite-nas/shared/browserauth"
	"lite-nas/shared/messaging"
)

// Session is the shared browser session used by gateway auth routes.
type Session = browserauth.Session

// AuthRequestContext carries caller metadata to the Auth Service.
type AuthRequestContext = browserauth.AuthRequestContext

// AccessTokenVerifier verifies the access token returned by Auth Service.
type AccessTokenVerifier interface {
	Verify(string) (authtoken.AccessClaims, error)
}

// AuthService exposes the shared browser auth lifecycle to the gateway.
type AuthService = browserauth.AuthService

// AccessTokenCookieName is the shared access cookie name.
const AccessTokenCookieName = browserauth.AccessTokenCookieName

// RefreshTokenCookieName is the shared refresh cookie name.
const RefreshTokenCookieName = browserauth.RefreshTokenCookieName

// NewAuthService creates a browser auth client backed by Auth Service RPCs.
func NewAuthService(client messaging.Client, verifier AccessTokenVerifier) AuthService {
	return browserauth.NewAuthService(client, verifier)
}

// AccessCookie issues a host-only access cookie for legacy gateway callers.
func AccessCookie(session Session) http.Cookie { return browserauth.AccessCookie(session, "") }

// RefreshCookie issues a host-only refresh cookie for legacy gateway callers.
func RefreshCookie(session Session) http.Cookie { return browserauth.RefreshCookie(session, "") }

// ClearAccessCookie expires a host-only access cookie.
func ClearAccessCookie(now time.Time) http.Cookie { return browserauth.ClearAccessCookie(now, "") }

// ClearRefreshCookie expires a host-only refresh cookie.
func ClearRefreshCookie(now time.Time) http.Cookie { return browserauth.ClearRefreshCookie(now, "") }
