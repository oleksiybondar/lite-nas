package controllers

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	authdto "lite-nas/services/web-gateway/dto/auth"
	"lite-nas/services/web-gateway/services"
)

type stubAuthService struct {
	loginResult   services.Session
	refreshResult services.Session
	logoutResult  services.Session
	meResult      services.Session
	loginErr      error
	refreshErr    error
	logoutErr     error
	meErr         error
}

func (s stubAuthService) Login(
	context.Context,
	time.Time,
	string,
	string,
	services.AuthRequestContext,
) (services.Session, error) {
	return s.loginResult, s.loginErr
}

func (s stubAuthService) Refresh(
	context.Context,
	time.Time,
	string,
	services.AuthRequestContext,
) (services.Session, error) {
	return s.refreshResult, s.refreshErr
}

func (s stubAuthService) Logout(
	context.Context,
	time.Time,
	string,
	services.AuthRequestContext,
) (services.Session, error) {
	return s.logoutResult, s.logoutErr
}

func (s stubAuthService) Me(time.Time, string) (services.Session, error) {
	return s.meResult, s.meErr
}

func authSessionFixture() services.Session {
	now := time.Unix(100, 0)

	return services.Session{
		UserID:        "stub-user",
		Login:         "john.doe",
		AccessToken:   "AT-token",
		RefreshToken:  "RT-token",
		AccessExpires: now.Add(time.Minute),
		RefreshExpiry: now.Add(2 * time.Minute),
		AuthType:      "jwt",
		Roles:         []string{"admin"},
		Scopes:        []string{"auth.me.read"},
	}
}

func newAuthControllerWithLoginResult() AuthController {
	return NewAuthController(stubAuthService{loginResult: authSessionFixture()})
}

func newAuthControllerWithRefreshResult() AuthController {
	return NewAuthController(stubAuthService{refreshResult: authSessionFixture()})
}

func newAuthControllerWithLogoutResult() AuthController {
	return NewAuthController(stubAuthService{logoutResult: authSessionFixture()})
}

func newAuthControllerWithMeResult() AuthController {
	return NewAuthController(stubAuthService{meResult: authSessionFixture()})
}

func authLoginInput() *authdto.LoginInput {
	return &authdto.LoginInput{
		Body: authdto.LoginRequestBody{
			Login:    "john.doe",
			Password: "pass",
		},
	}
}

func authRefreshInput(token string) *authdto.RefreshInput {
	return &authdto.RefreshInput{
		Body: authdto.RefreshRequestBody{RefreshToken: token},
	}
}

func authLogoutInput(token string) *authdto.LogoutInput {
	return &authdto.LogoutInput{
		Body: authdto.LogoutRequestBody{RefreshToken: token},
	}
}

func mustLogin(t *testing.T, controller AuthController) *authdto.SessionOutput {
	t.Helper()

	got, err := controller.Login(context.Background(), authLoginInput())
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	return got
}

func mustRefresh(t *testing.T, controller AuthController, input *authdto.RefreshInput) *authdto.SessionOutput {
	t.Helper()

	got, err := controller.Refresh(context.Background(), input)
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}

	return got
}

func mustLogout(t *testing.T, controller AuthController, input *authdto.LogoutInput) *authdto.LogoutOutput {
	t.Helper()

	got, err := controller.Logout(context.Background(), input)
	if err != nil {
		t.Fatalf("Logout() error = %v", err)
	}

	return got
}

func mustMe(t *testing.T, controller AuthController, accessToken string) *authdto.MeOutput {
	t.Helper()

	got, err := controller.Me(context.Background(), &authdto.MeInput{AccessTokenCookie: accessToken})
	if err != nil {
		t.Fatalf("Me() error = %v", err)
	}

	return got
}

// Requirements: web-gateway/FR-004, web-gateway/TR-001
func TestAuthControllerLoginReturnsSessionAndCookies(t *testing.T) {
	t.Parallel()

	controller := newAuthControllerWithLoginResult()
	got := mustLogin(t, controller)

	if len(got.SetCookie) != 2 {
		t.Fatalf("cookie count = %d, want 2", len(got.SetCookie))
	}

	if got.Body.Data.AccessToken != "AT-token" {
		t.Fatalf("AccessToken = %q, want %q", got.Body.Data.AccessToken, "AT-token")
	}
}

// Requirements: web-gateway/FR-004, web-gateway/TR-001
func TestAuthControllerRefreshPrefersPayloadToken(t *testing.T) {
	t.Parallel()

	controller := newAuthControllerWithRefreshResult()
	got := mustRefresh(t, controller, &authdto.RefreshInput{
		RefreshTokenCookie: "RT-cookie",
		Body: authdto.RefreshRequestBody{
			RefreshToken: "RT-body",
		},
	})

	if got.Body.Data.RefreshToken != "RT-token" {
		t.Fatalf("RefreshToken = %q, want %q", got.Body.Data.RefreshToken, "RT-token")
	}
}

// Requirements: web-gateway/FR-004, web-gateway/TR-001
func TestAuthControllerRefreshRejectsMissingToken(t *testing.T) {
	t.Parallel()

	controller := NewAuthController(stubAuthService{})

	got, err := controller.Refresh(context.Background(), &authdto.RefreshInput{})
	if err == nil {
		t.Fatal("Refresh() error = nil, want error")
	}

	if got != nil {
		t.Fatalf("Refresh() result = %#v, want nil", got)
	}
}

// Requirements: web-gateway/FR-004, web-gateway/TR-001
func TestAuthControllerLogoutClearsCookies(t *testing.T) {
	t.Parallel()

	controller := newAuthControllerWithLogoutResult()
	got := mustLogout(t, controller, authLogoutInput("RT-body"))

	if len(got.SetCookie) != 2 {
		t.Fatalf("cookie count = %d, want 2", len(got.SetCookie))
	}

	for _, cookie := range got.SetCookie {
		if cookie.MaxAge != -1 {
			t.Fatalf("cookie MaxAge = %d, want -1", cookie.MaxAge)
		}
	}
}

// Requirements: web-gateway/FR-004, web-gateway/TR-001
func TestAuthControllerMeReturnsAuthenticatedUser(t *testing.T) {
	t.Parallel()

	controller := newAuthControllerWithMeResult()
	got := mustMe(t, controller, "AT-token")

	if !got.Body.Data.Authenticated {
		t.Fatalf("Authenticated = false, want true")
	}

	if got.Body.Data.User.Login != "john.doe" {
		t.Fatalf("Login = %q, want %q", got.Body.Data.User.Login, "john.doe")
	}
}

// Requirements: web-gateway/TR-001
func TestAuthControllerLoginMapsMissingCredentialsToUnauthorized(t *testing.T) {
	t.Parallel()

	assertAuthControllerUnauthorizedError(t, stubAuthService{loginErr: services.ErrMissingCredentials()}, func(controller AuthController) error {
		_, err := controller.Login(context.Background(), authLoginInput())
		return err
	})
}

// Requirements: web-gateway/TR-001
func TestAuthControllerRefreshMapsMissingRefreshTokenToUnauthorized(t *testing.T) {
	t.Parallel()

	assertAuthControllerUnauthorizedError(t, stubAuthService{refreshErr: services.ErrMissingRefreshToken()}, func(controller AuthController) error {
		_, err := controller.Refresh(context.Background(), authRefreshInput("RT-token"))
		return err
	})
}

// Requirements: web-gateway/TR-001
func TestAuthControllerLogoutMapsMissingRefreshTokenToUnauthorized(t *testing.T) {
	t.Parallel()

	assertAuthControllerUnauthorizedError(t, stubAuthService{logoutErr: services.ErrMissingRefreshToken()}, func(controller AuthController) error {
		_, err := controller.Logout(context.Background(), authLogoutInput("RT-token"))
		return err
	})
}

// Requirements: web-gateway/TR-001
func TestAuthControllerMeMapsUnauthorizedToUnauthorized(t *testing.T) {
	t.Parallel()

	assertAuthControllerUnauthorizedError(t, stubAuthService{meErr: services.ErrUnauthorized}, func(controller AuthController) error {
		_, err := controller.Me(context.Background(), &authdto.MeInput{AccessTokenCookie: "AT-token"})
		return err
	})
}

// Requirements: web-gateway/TR-001
func TestAuthControllerMapsUnexpectedServiceErrorToInternalError(t *testing.T) {
	t.Parallel()

	controller := NewAuthController(stubAuthService{refreshErr: errors.New("boom")})

	_, err := controller.Refresh(context.Background(), authRefreshInput("RT-token"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBuildSessionBodyReturnsAuthenticatedEnvelope(t *testing.T) {
	t.Parallel()

	body := buildSessionBody(time.Unix(100, 0), authSessionFixture())
	if !body.Data.Authenticated {
		t.Fatalf("Authenticated = false, want true")
	}
}

func TestResolveRefreshTokenPrefersPayloadThenCookie(t *testing.T) {
	t.Parallel()

	if got := resolveRefreshToken(nil); got != "" {
		t.Fatalf("resolveRefreshToken(nil) = %q, want empty string", got)
	}

	payload := resolveRefreshToken(&authdto.RefreshInput{
		RefreshTokenCookie: "RT-cookie",
		Body:               authdto.RefreshRequestBody{RefreshToken: "RT-body"},
	})
	if payload != "RT-body" {
		t.Fatalf("resolveRefreshToken(payload) = %q, want %q", payload, "RT-body")
	}

	cookie := resolveRefreshToken(&authdto.RefreshInput{RefreshTokenCookie: "RT-cookie"})
	if cookie != "RT-cookie" {
		t.Fatalf("resolveRefreshToken(cookie) = %q, want %q", cookie, "RT-cookie")
	}
}

func TestResolveLogoutRefreshTokenPrefersPayloadThenCookie(t *testing.T) {
	t.Parallel()

	if got := resolveLogoutRefreshToken(nil); got != "" {
		t.Fatalf("resolveLogoutRefreshToken(nil) = %q, want empty string", got)
	}

	payload := resolveLogoutRefreshToken(&authdto.LogoutInput{
		RefreshTokenCookie: "RT-cookie",
		Body:               authdto.LogoutRequestBody{RefreshToken: "RT-body"},
	})
	if payload != "RT-body" {
		t.Fatalf("resolveLogoutRefreshToken(payload) = %q, want %q", payload, "RT-body")
	}

	cookie := resolveLogoutRefreshToken(&authdto.LogoutInput{RefreshTokenCookie: "RT-cookie"})
	if cookie != "RT-cookie" {
		t.Fatalf("resolveLogoutRefreshToken(cookie) = %q, want %q", cookie, "RT-cookie")
	}
}

func TestMapAuthErrorReturnsUnauthorizedForUnauthorized(t *testing.T) {
	t.Parallel()

	if err := mapAuthError(services.ErrUnauthorized, "nope"); err == nil {
		t.Fatal("expected unauthorized mapping")
	}
}

func TestMapAuthErrorReturnsUnauthorizedForMissingRefreshToken(t *testing.T) {
	t.Parallel()

	if err := mapAuthError(services.ErrMissingRefreshToken(), "nope"); err == nil {
		t.Fatal("expected unauthorized mapping")
	}
}

func TestMapAuthErrorReturnsUnauthorizedForMissingCredentials(t *testing.T) {
	t.Parallel()

	if err := mapAuthError(services.ErrMissingCredentials(), "nope"); err == nil {
		t.Fatal("expected unauthorized mapping")
	}
}

func TestMapAuthErrorReturnsInternalErrorForUnexpectedErrors(t *testing.T) {
	t.Parallel()

	if err := mapAuthError(errors.New("boom"), "nope"); err == nil {
		t.Fatal("expected internal error mapping")
	}
}

func TestAuthControllerLogoutSetsCookieHeaders(t *testing.T) {
	t.Parallel()

	controller := newAuthControllerWithLogoutResult()
	got := mustLogout(t, controller, authLogoutInput("RT-body"))

	assertSecureCookieHeaders(t, got.SetCookie)
}

func TestAuthControllerLoginSetsCookieHeaders(t *testing.T) {
	t.Parallel()

	controller := newAuthControllerWithLoginResult()
	got := mustLogin(t, controller)

	assertSecureCookieHeaders(t, got.SetCookie)
}

func TestAuthControllerMeDoesNotSetCookies(t *testing.T) {
	t.Parallel()

	controller := newAuthControllerWithMeResult()
	got := mustMe(t, controller, "AT-token")

	if got.Body.Success != true {
		t.Fatalf("Success = %v, want true", got.Body.Success)
	}
}

func TestAuthControllerRefreshSetsCookieHeaders(t *testing.T) {
	t.Parallel()

	controller := newAuthControllerWithRefreshResult()
	got := mustRefresh(t, controller, authRefreshInput("RT-token"))

	for _, cookie := range got.SetCookie {
		if cookie.Path != "/" {
			t.Fatalf("cookie path = %q, want /", cookie.Path)
		}
	}
}

func TestAuthControllerLoginMapsUnauthorizedStatus(t *testing.T) {
	t.Parallel()

	controller := NewAuthController(stubAuthService{loginErr: services.ErrMissingCredentials()})
	_, err := controller.Login(context.Background(), authLoginInput())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAuthControllerMeUsesAccessTokenInput(t *testing.T) {
	t.Parallel()

	session := authSessionFixture()
	session.AccessToken = "AT-custom"
	controller := NewAuthController(stubAuthService{meResult: session})
	got := mustMe(t, controller, "AT-custom")

	if got.Body.Data.AuthType != "jwt" {
		t.Fatalf("AuthType = %q, want jwt", got.Body.Data.AuthType)
	}
}

func assertAuthControllerUnauthorizedError(t *testing.T, service stubAuthService, invoker func(AuthController) error) {
	t.Helper()

	controller := NewAuthController(service)

	if err := invoker(controller); err == nil {
		t.Fatal("expected error")
	}
}

func assertSecureCookieHeaders(t *testing.T, cookies []http.Cookie) {
	t.Helper()

	for _, cookie := range cookies {
		assertCookiePathRoot(t, cookie)
		assertCookieHTTPOnly(t, cookie)
		assertCookieSecure(t, cookie)
	}
}

func assertCookiePathRoot(t *testing.T, cookie http.Cookie) {
	t.Helper()

	if cookie.Path != "/" {
		t.Fatalf("cookie path = %q, want /", cookie.Path)
	}
}

func assertCookieHTTPOnly(t *testing.T, cookie http.Cookie) {
	t.Helper()

	if !cookie.HttpOnly {
		t.Fatal("expected HttpOnly cookie")
	}
}

func assertCookieSecure(t *testing.T, cookie http.Cookie) {
	t.Helper()

	if !cookie.Secure {
		t.Fatal("expected Secure cookie")
	}
}

func TestAuthControllerLogoutResponseMessage(t *testing.T) {
	t.Parallel()

	controller := newAuthControllerWithLogoutResult()
	got := mustLogout(t, controller, authLogoutInput("RT-token"))

	if got.Body.Message != "logged out" {
		t.Fatalf("Message = %q, want %q", got.Body.Message, "logged out")
	}
}

func TestAccessCookieAndRefreshCookieSetExpectedNames(t *testing.T) {
	t.Parallel()

	session := authSessionFixture()
	accessCookie := services.AccessCookie(session)
	refreshCookie := services.RefreshCookie(session)

	if accessCookie.Name != services.AccessTokenCookieName {
		t.Fatalf("access cookie name = %q, want %q", accessCookie.Name, services.AccessTokenCookieName)
	}

	if refreshCookie.Name != services.RefreshTokenCookieName {
		t.Fatalf("refresh cookie name = %q, want %q", refreshCookie.Name, services.RefreshTokenCookieName)
	}
}

func TestClearCookiesSetExpiredState(t *testing.T) {
	t.Parallel()

	now := time.Unix(100, 0)
	for _, cookie := range []http.Cookie{
		services.ClearAccessCookie(now),
		services.ClearRefreshCookie(now),
	} {
		if cookie.MaxAge != -1 {
			t.Fatalf("MaxAge = %d, want -1", cookie.MaxAge)
		}
	}
}
