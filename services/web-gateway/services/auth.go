package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	// AccessTokenCookieName is the browser cookie name used for access tokens.
	// #nosec G101 -- cookie identifier only, not a credential
	AccessTokenCookieName = "lite-nas-at"
	// RefreshTokenCookieName is the browser cookie name used for refresh tokens.
	// #nosec G101 -- cookie identifier only, not a credential
	RefreshTokenCookieName = "lite-nas-rt"

	accessTokenPrefix  = "AT-"
	refreshTokenPrefix = "RT-"
	defaultUserID      = "stub-user"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 24 * time.Hour
)

// Session contains the auth state returned by the gateway auth service.
//
// The current implementation is still a stub, so the values represent
// placeholder identity and token data rather than persisted session state.
type Session struct {
	UserID        string
	Login         string
	AccessToken   string
	RefreshToken  string
	AccessExpires time.Time
	RefreshExpiry time.Time
	AuthType      string
	Roles         []string
	Scopes        []string
}

var (
	errMissingCredentials  = errors.New("missing login credentials")
	errMissingRefreshToken = errors.New("missing refresh token")
)

// AuthService defines the auth flows used by the gateway service layer.
type AuthService interface {
	Login(now time.Time, login string, password string) (Session, error)
	Refresh(now time.Time, refreshToken string) (Session, error)
	Logout(now time.Time, refreshToken string) (Session, error)
	Me(now time.Time, accessToken string) (Session, error)
}

type authService struct{}

// NewAuthService creates the current auth service implementation.
//
// The returned service is still a stub and exists to unblock the gateway
// transport while the real auth backend is being integrated.
func NewAuthService() AuthService {
	return authService{}
}

// Login issues a new auth token pair when both credentials are present.
//
// Parameters:
//   - now: clock value used to stamp the generated token expirations
//   - login: submitted login identifier
//   - password: submitted password
//
// Intentional simplification:
//   - while the auth service is still a skeleton, login succeeds on the happy
//     path when both fields are present
//   - TODO: replace this with real credential verification once the auth
//     service is implemented
func (authService) Login(now time.Time, login string, password string) (Session, error) {
	if strings.TrimSpace(login) == "" || strings.TrimSpace(password) == "" {
		return Session{}, errMissingCredentials
	}

	return newSession(now, login)
}

// Refresh issues a new token pair when a refresh token is present.
//
// Parameters:
//   - now: clock value used to stamp the generated token expirations
//   - refreshToken: submitted refresh token value
//
// Intentional simplification:
//   - while the auth service is still a skeleton, refresh succeeds on the
//     happy path when a refresh token is present
//   - TODO: replace this with real refresh-token verification once the auth
//     service is implemented
func (authService) Refresh(now time.Time, refreshToken string) (Session, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return Session{}, errMissingRefreshToken
	}

	return newSession(now, defaultUserID)
}

// Logout returns a session payload whose cookie timestamps force browser-side
// token expiry.
//
// Parameters:
//   - now: clock value used to compute expired cookie timestamps
//   - refreshToken: submitted refresh token value
//
// Intentional simplification:
//   - while the auth service is still a skeleton, logout succeeds on the happy
//     path when a refresh token is present
//   - TODO: replace this with real session invalidation once the auth service
//     is implemented
func (authService) Logout(now time.Time, refreshToken string) (Session, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return Session{}, errMissingRefreshToken
	}

	return Session{
		UserID:        defaultUserID,
		AccessExpires: now.Add(-time.Hour),
		RefreshExpiry: now.Add(-time.Hour),
	}, nil
}

// Me validates the access token format and returns the current stub user
// session.
//
// Parameters:
//   - now: clock value used to compute access-token expiry in the response
//   - accessToken: caller-provided access token extracted by transport logic
func (authService) Me(now time.Time, accessToken string) (Session, error) {
	if !strings.HasPrefix(strings.TrimSpace(accessToken), accessTokenPrefix) {
		return Session{}, ErrUnauthorized
	}

	return Session{
		UserID:        defaultUserID,
		Login:         defaultUserID,
		AccessToken:   accessToken,
		AccessExpires: now.Add(accessTokenTTL),
		AuthType:      "jwt",
		Roles:         []string{"admin"},
		Scopes: []string{
			"auth.me.read",
			"system-metrics.snapshot.read",
			"system-metrics.history.read",
		},
	}, nil
}

// AccessCookie converts session access-token data into the browser cookie that
// the gateway writes on auth responses.
func AccessCookie(session Session) http.Cookie {
	return http.Cookie{
		Name:     AccessTokenCookieName,
		Value:    session.AccessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  session.AccessExpires,
	}
}

// RefreshCookie converts session refresh-token data into the browser cookie
// that the gateway writes on auth responses.
func RefreshCookie(session Session) http.Cookie {
	return http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    session.RefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  session.RefreshExpiry,
	}
}

// ClearAccessCookie returns an expired access-token cookie for logout flows.
func ClearAccessCookie(now time.Time) http.Cookie {
	return expiredCookie(AccessTokenCookieName, now)
}

// ClearRefreshCookie returns an expired refresh-token cookie for logout flows.
func ClearRefreshCookie(now time.Time) http.Cookie {
	return expiredCookie(RefreshTokenCookieName, now)
}

func newSession(now time.Time, login string) (Session, error) {
	accessToken, err := newToken(accessTokenPrefix)
	if err != nil {
		return Session{}, err
	}

	refreshToken, err := newToken(refreshTokenPrefix)
	if err != nil {
		return Session{}, err
	}

	return Session{
		UserID:        defaultUserID,
		Login:         login,
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		AccessExpires: now.Add(accessTokenTTL),
		RefreshExpiry: now.Add(refreshTokenTTL),
		AuthType:      "jwt",
		Roles:         []string{"admin"},
		Scopes: []string{
			"auth.me.read",
			"system-metrics.snapshot.read",
			"system-metrics.history.read",
		},
	}, nil
}

func newToken(prefix string) (string, error) {
	id, err := newUUID()
	if err != nil {
		return "", err
	}

	return prefix + id, nil
}

func newUUID() (string, error) {
	data := make([]byte, 16)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}

	data[6] = (data[6] & 0x0f) | 0x40
	data[8] = (data[8] & 0x3f) | 0x80

	return fmt.Sprintf(
		"%08x-%04x-%04x-%04x-%012x",
		data[0:4],
		data[4:6],
		data[6:8],
		data[8:10],
		data[10:16],
	), nil
}

func expiredCookie(name string, now time.Time) http.Cookie {
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
