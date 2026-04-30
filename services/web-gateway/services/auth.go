package services

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"lite-nas/shared/authtoken"
	authcontract "lite-nas/shared/contracts/auth"
	"lite-nas/shared/httpcookie"
	"lite-nas/shared/messaging"
)

const (
	// AccessTokenCookieName is the browser cookie name used for access tokens.
	// #nosec G101 -- cookie identifier only, not a credential
	AccessTokenCookieName = "lite-nas-at"
	// RefreshTokenCookieName is the browser cookie name used for refresh tokens.
	// #nosec G101 -- cookie identifier only, not a credential
	RefreshTokenCookieName = "lite-nas-rt"

	authTypeJWT = "jwt"
)

// Session contains the auth state returned by the gateway auth service.
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

// AuthRequestContext contains caller metadata passed through to auth-service
// token lifecycle RPCs.
type AuthRequestContext struct {
	ClientIP  string
	UserAgent string
}

// AccessTokenVerifier defines the local JWT verification behavior needed by
// the gateway service layer.
type AccessTokenVerifier interface {
	Verify(tokenText string) (authtoken.AccessClaims, error)
}

// AuthService defines the auth flows used by the gateway service layer.
type AuthService interface {
	Login(ctx context.Context, now time.Time, login string, password string, requestContext AuthRequestContext) (Session, error)
	Refresh(ctx context.Context, now time.Time, refreshToken string, requestContext AuthRequestContext) (Session, error)
	Logout(ctx context.Context, now time.Time, refreshToken string, requestContext AuthRequestContext) (Session, error)
	Me(now time.Time, accessToken string) (Session, error)
}

type authService struct {
	client   messaging.Client
	verifier AccessTokenVerifier
}

// NewAuthService creates an auth service backed by auth-service RPCs and local
// access-token verification.
//
// Parameters:
//   - client: messaging client used for auth-service request/reply calls
//   - verifier: local verifier for JWT access tokens returned by auth-service
func NewAuthService(client messaging.Client, verifier AccessTokenVerifier) AuthService {
	return authService{
		client:   client,
		verifier: verifier,
	}
}

// Login authenticates credentials through auth-service and maps the issued
// token pair into a browser session.
//
// Parameters:
//   - ctx: request-scoped context used for RPC cancellation
//   - now: clock value used for responses that do not carry token expirations
//   - login: submitted login identifier
//   - password: submitted password
//   - requestContext: caller metadata forwarded to auth-service
func (s authService) Login(
	ctx context.Context,
	now time.Time,
	login string,
	password string,
	requestContext AuthRequestContext,
) (Session, error) {
	if strings.TrimSpace(login) == "" || strings.TrimSpace(password) == "" {
		return Session{}, errMissingCredentials
	}

	var response authcontract.LoginResponse
	if err := s.client.Request(ctx, authcontract.LoginRPCSubject, authcontract.LoginRequest{
		Username:  strings.TrimSpace(login),
		Password:  password,
		ClientIP:  requestContext.ClientIP,
		UserAgent: requestContext.UserAgent,
	}, &response); err != nil {
		return Session{}, err
	}

	if response.Status != authcontract.StatusAuthenticated {
		return Session{}, ErrUnauthorized
	}

	return s.sessionFromTokens(now, response.AccessToken, response.RefreshToken)
}

// Refresh rotates the token pair through auth-service.
//
// Parameters:
//   - ctx: request-scoped context used for RPC cancellation
//   - now: clock value used for responses that do not carry token expirations
//   - refreshToken: submitted refresh token value
//   - requestContext: caller metadata forwarded to auth-service
func (s authService) Refresh(
	ctx context.Context,
	now time.Time,
	refreshToken string,
	requestContext AuthRequestContext,
) (Session, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return Session{}, errMissingRefreshToken
	}

	var response authcontract.RefreshResponse
	if err := s.client.Request(ctx, authcontract.RefreshRPCSubject, authcontract.RefreshRequest{
		RefreshToken: strings.TrimSpace(refreshToken),
		ClientIP:     requestContext.ClientIP,
		UserAgent:    requestContext.UserAgent,
	}, &response); err != nil {
		return Session{}, err
	}

	return s.sessionFromTokens(now, response.AccessToken, response.RefreshToken)
}

// Logout revokes the submitted refresh token through auth-service and returns
// cookie timestamps that force browser-side token expiry.
//
// Parameters:
//   - ctx: request-scoped context used for RPC cancellation
//   - now: clock value used to compute expired cookie timestamps
//   - refreshToken: submitted refresh token value
//   - requestContext: caller metadata forwarded to auth-service
func (s authService) Logout(
	ctx context.Context,
	now time.Time,
	refreshToken string,
	requestContext AuthRequestContext,
) (Session, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return Session{}, errMissingRefreshToken
	}

	var response authcontract.LogoutResponse
	if err := s.client.Request(ctx, authcontract.LogoutRPCSubject, authcontract.LogoutRequest{
		RefreshToken: strings.TrimSpace(refreshToken),
		ClientIP:     requestContext.ClientIP,
		UserAgent:    requestContext.UserAgent,
	}, &response); err != nil {
		return Session{}, err
	}
	if !response.LoggedOut {
		return Session{}, ErrUnauthorized
	}

	return Session{
		AccessExpires: now,
		RefreshExpiry: now,
	}, nil
}

// Me locally verifies the access token and returns the authenticated principal
// represented by its JWT claims.
//
// Parameters:
//   - now: clock value used to compute access-token expiry in the response
//   - accessToken: caller-provided access token extracted by transport logic
func (s authService) Me(now time.Time, accessToken string) (Session, error) {
	claims, err := s.verifyAccessToken(accessToken)
	if err != nil {
		return Session{}, ErrUnauthorized
	}

	return sessionFromClaims(now, strings.TrimSpace(accessToken), "", claims)
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
	return httpcookie.Expired(AccessTokenCookieName, now)
}

// ClearRefreshCookie returns an expired refresh-token cookie for logout flows.
func ClearRefreshCookie(now time.Time) http.Cookie {
	return httpcookie.Expired(RefreshTokenCookieName, now)
}

func (s authService) sessionFromTokens(now time.Time, accessToken string, refreshToken string) (Session, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return Session{}, errMissingRefreshToken
	}

	claims, err := s.verifyAccessToken(accessToken)
	if err != nil {
		return Session{}, ErrUnauthorized
	}

	return sessionFromClaims(now, strings.TrimSpace(accessToken), strings.TrimSpace(refreshToken), claims)
}

func (s authService) verifyAccessToken(accessToken string) (authtoken.AccessClaims, error) {
	token := strings.TrimSpace(accessToken)
	if token == "" {
		return authtoken.AccessClaims{}, ErrUnauthorized
	}
	if s.verifier == nil {
		return authtoken.AccessClaims{}, ErrUnauthorized
	}

	return s.verifier.Verify(token)
}

func sessionFromClaims(
	now time.Time,
	accessToken string,
	refreshToken string,
	claims authtoken.AccessClaims,
) (Session, error) {
	if claims.ExpiresAt == nil {
		return Session{}, ErrUnauthorized
	}

	return Session{
		UserID:        claims.Subject,
		Login:         claims.Login,
		AccessToken:   accessToken,
		RefreshToken:  refreshToken,
		AccessExpires: claims.ExpiresAt.Time,
		RefreshExpiry: now.Add(24 * time.Hour),
		AuthType:      authTypeJWT,
		Roles:         claims.Roles,
		Scopes:        claims.Scopes,
	}, nil
}
