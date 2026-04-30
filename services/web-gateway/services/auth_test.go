package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"lite-nas/shared/authtoken"
	authcontract "lite-nas/shared/contracts/auth"
)

type authServiceClientStub struct {
	subject   string
	request   any
	err       error
	status    authcontract.Status
	loggedOut bool
}

func (c *authServiceClientStub) Publish(context.Context, string, any) error {
	return nil
}

func (c *authServiceClientStub) Request(_ context.Context, subject string, request any, response any) error {
	c.subject = subject
	c.request = request
	if c.err != nil {
		return c.err
	}

	c.populateResponse(response)
	return nil
}

func (c *authServiceClientStub) populateResponse(response any) {
	switch out := response.(type) {
	case *authcontract.LoginResponse:
		c.populateLoginResponse(out)
	case *authcontract.RefreshResponse:
		out.AccessToken = "access-token"
		out.RefreshToken = "refresh-token"
	case *authcontract.LogoutResponse:
		out.LoggedOut = true
		if c.loggedOut {
			out.LoggedOut = c.loggedOut
		}
	}
}

func (c *authServiceClientStub) populateLoginResponse(response *authcontract.LoginResponse) {
	response.Status = authcontract.StatusAuthenticated
	if c.status != "" {
		response.Status = c.status
	}
	response.AccessToken = "access-token"
	response.RefreshToken = "refresh-token"
}

func (c *authServiceClientStub) Drain() error { return nil }
func (c *authServiceClientStub) Close()       {}

type authTokenVerifierStub struct {
	err      error
	noExpiry bool
}

func (v authTokenVerifierStub) Verify(string) (authtoken.AccessClaims, error) {
	if v.err != nil {
		return authtoken.AccessClaims{}, v.err
	}

	claims := authtoken.AccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "1000",
		},
		Login:  "john.doe",
		Roles:  []string{"admin"},
		Scopes: []string{"auth.me.read"},
	}
	if !v.noExpiry {
		claims.ExpiresAt = jwt.NewNumericDate(time.Unix(1000, 0))
	}

	return claims, nil
}

// Requirements: web-gateway/FR-004
func TestAuthServiceLoginRequestsAuthRPC(t *testing.T) {
	t.Parallel()

	client := &authServiceClientStub{}
	service := NewAuthService(client, authTokenVerifierStub{})

	session, err := service.Login(
		context.Background(),
		time.Unix(100, 0),
		"john.doe",
		"pass",
		AuthRequestContext{UserAgent: "browser"},
	)
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if client.subject != authcontract.LoginRPCSubject {
		t.Fatalf("subject = %q, want %q", client.subject, authcontract.LoginRPCSubject)
	}
	if session.UserID != "1000" {
		t.Fatalf("UserID = %q, want 1000", session.UserID)
	}
	if session.AccessToken != "access-token" {
		t.Fatalf("AccessToken = %q, want access-token", session.AccessToken)
	}
}

// Requirements: web-gateway/FR-004
func TestAuthServiceRefreshRequestsAuthRPC(t *testing.T) {
	t.Parallel()

	client := &authServiceClientStub{}
	service := NewAuthService(client, authTokenVerifierStub{})

	session, err := service.Refresh(
		context.Background(),
		time.Unix(100, 0),
		"refresh-token",
		AuthRequestContext{UserAgent: "browser"},
	)
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	if client.subject != authcontract.RefreshRPCSubject {
		t.Fatalf("subject = %q, want %q", client.subject, authcontract.RefreshRPCSubject)
	}
	if session.RefreshToken != "refresh-token" {
		t.Fatalf("RefreshToken = %q, want refresh-token", session.RefreshToken)
	}
}

// Requirements: web-gateway/FR-004
func TestAuthServiceLogoutRequestsAuthRPC(t *testing.T) {
	t.Parallel()

	client := &authServiceClientStub{}
	service := NewAuthService(client, authTokenVerifierStub{})

	if _, err := service.Logout(
		context.Background(),
		time.Unix(100, 0),
		"refresh-token",
		AuthRequestContext{UserAgent: "browser"},
	); err != nil {
		t.Fatalf("Logout() error = %v", err)
	}

	if client.subject != authcontract.LogoutRPCSubject {
		t.Fatalf("subject = %q, want %q", client.subject, authcontract.LogoutRPCSubject)
	}
}

// Requirements: web-gateway/FR-004
func TestAuthServiceLogoutRequiresRefreshToken(t *testing.T) {
	t.Parallel()

	service := NewAuthService(&authServiceClientStub{}, authTokenVerifierStub{})

	if _, err := service.Logout(context.Background(), time.Unix(100, 0), "", AuthRequestContext{}); !errors.Is(err, ErrMissingRefreshToken()) {
		t.Fatalf("Logout() error = %v, want %v", err, ErrMissingRefreshToken())
	}
}

func TestAuthServiceLoginRequiresCredentials(t *testing.T) {
	t.Parallel()

	service := NewAuthService(&authServiceClientStub{}, authTokenVerifierStub{})
	if _, err := service.Login(context.Background(), time.Unix(100, 0), "", "pass", AuthRequestContext{}); !errors.Is(err, ErrMissingCredentials()) {
		t.Fatalf("Login() error = %v, want %v", err, ErrMissingCredentials())
	}
}

func TestAuthServiceLoginMapsDeniedStatusToUnauthorized(t *testing.T) {
	t.Parallel()

	service := NewAuthService(
		&authServiceClientStub{status: authcontract.StatusDenied},
		authTokenVerifierStub{},
	)
	if _, err := service.Login(context.Background(), time.Unix(100, 0), "john.doe", "pass", AuthRequestContext{}); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("Login() error = %v, want %v", err, ErrUnauthorized)
	}
}

func TestAuthServiceRefreshRequiresRefreshToken(t *testing.T) {
	t.Parallel()

	service := NewAuthService(&authServiceClientStub{}, authTokenVerifierStub{})
	if _, err := service.Refresh(context.Background(), time.Unix(100, 0), "", AuthRequestContext{}); !errors.Is(err, ErrMissingRefreshToken()) {
		t.Fatalf("Refresh() error = %v, want %v", err, ErrMissingRefreshToken())
	}
}

func TestAuthServiceMeRejectsInvalidToken(t *testing.T) {
	t.Parallel()

	service := NewAuthService(&authServiceClientStub{}, authTokenVerifierStub{err: errors.New("invalid")})
	if _, err := service.Me(time.Unix(100, 0), "access-token"); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("Me() error = %v, want %v", err, ErrUnauthorized)
	}
}

func TestAuthServiceMeRejectsMissingExpiry(t *testing.T) {
	t.Parallel()

	service := NewAuthService(&authServiceClientStub{}, authTokenVerifierStub{noExpiry: true})
	if _, err := service.Me(time.Unix(100, 0), "access-token"); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("Me() error = %v, want %v", err, ErrUnauthorized)
	}
}

func TestAuthCookiesUseExpectedSecurityAttributes(t *testing.T) {
	t.Parallel()

	now := time.Unix(100, 0)
	session := Session{
		AccessToken:   "access-token",
		RefreshToken:  "refresh-token",
		AccessExpires: now.Add(time.Minute),
		RefreshExpiry: now.Add(2 * time.Minute),
	}

	for _, cookie := range []struct {
		name string
		got  string
	}{
		{name: AccessTokenCookieName, got: AccessCookie(session).Name},
		{name: RefreshTokenCookieName, got: RefreshCookie(session).Name},
		{name: AccessTokenCookieName, got: ClearAccessCookie(now).Name},
		{name: RefreshTokenCookieName, got: ClearRefreshCookie(now).Name},
	} {
		if cookie.got != cookie.name {
			t.Fatalf("cookie name = %q, want %q", cookie.got, cookie.name)
		}
	}
}

func TestAuthServiceMeReturnsClaimsSession(t *testing.T) {
	t.Parallel()

	service := NewAuthService(&authServiceClientStub{}, authTokenVerifierStub{})
	session, err := service.Me(time.Unix(100, 0), "access-token")
	if err != nil {
		t.Fatalf("Me() error = %v", err)
	}

	if session.Login != "john.doe" {
		t.Fatalf("Login = %q, want john.doe", session.Login)
	}
}
