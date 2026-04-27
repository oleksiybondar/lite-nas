package routes

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"lite-nas/services/web-gateway/controllers"
	"lite-nas/services/web-gateway/modules"
	"lite-nas/services/web-gateway/services"
	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/metrics"
)

type routeStubReader struct {
	data []byte
}

func (r routeStubReader) Read() ([]byte, error) {
	return r.data, nil
}

type routeAuthService struct {
	loginErr   error
	refreshErr error
	logoutErr  error
	meErr      error
}

func (s routeAuthService) Login(now time.Time, login string, password string) (services.Session, error) {
	if s.loginErr != nil {
		return services.Session{}, s.loginErr
	}

	return services.Session{
		UserID:        "stub-user",
		Login:         login,
		AccessToken:   "AT-login",
		RefreshToken:  "RT-login",
		AccessExpires: now.Add(time.Minute),
		RefreshExpiry: now.Add(2 * time.Minute),
	}, nil
}

func (s routeAuthService) Refresh(now time.Time, refreshToken string) (services.Session, error) {
	if s.refreshErr != nil {
		return services.Session{}, s.refreshErr
	}

	return services.Session{
		UserID:        "stub-user",
		Login:         "john.doe",
		AccessToken:   "AT-refresh",
		RefreshToken:  refreshToken,
		AccessExpires: now.Add(time.Minute),
		RefreshExpiry: now.Add(2 * time.Minute),
	}, nil
}

func (s routeAuthService) Logout(now time.Time, refreshToken string) (services.Session, error) {
	if s.logoutErr != nil {
		return services.Session{}, s.logoutErr
	}

	return services.Session{
		AccessExpires: now.Add(-time.Minute),
		RefreshExpiry: now.Add(-time.Minute),
	}, nil
}

func (s routeAuthService) Me(now time.Time, accessToken string) (services.Session, error) {
	if s.meErr != nil {
		return services.Session{}, s.meErr
	}

	return services.Session{
		UserID:        "stub-user",
		Login:         "john.doe",
		AccessToken:   accessToken,
		AccessExpires: now.Add(time.Minute),
		AuthType:      "jwt",
		Roles:         []string{"admin"},
		Scopes:        []string{"auth.me.read"},
	}, nil
}

type routeSystemMetricsService struct{}

func (routeSystemMetricsService) GetSnapshot(context.Context) (metrics.SystemSnapshot, error) {
	return metrics.SystemSnapshot{Timestamp: time.Unix(100, 0).UTC()}, nil
}

func (routeSystemMetricsService) GetHistory(context.Context) ([]metrics.SystemSnapshot, error) {
	return []metrics.SystemSnapshot{
		{Timestamp: time.Unix(100, 0).UTC()},
		{Timestamp: time.Unix(101, 0).UTC()},
	}, nil
}

func routerFixture(authService controllers.AuthService) http.Handler {
	if authService == nil {
		authService = routeAuthService{}
	}

	controllerModule := modules.Controllers{
		Auth: controllers.NewAuthController(authService),
		Static: controllers.NewStaticController(
			controllers.StaticFiles{
				IndexHTML: routeStubReader{data: []byte("<html>ok</html>")},
				IndexCSS:  routeStubReader{data: []byte("body {}")},
				IndexJS:   routeStubReader{data: []byte("console.log('ok')")},
				Favicon:   routeStubReader{data: []byte{0x00, 0x00, 0x01, 0x00}},
			},
			sharedlogger.NewNop(),
		),
		SystemMetrics: controllers.NewSystemMetricsController(routeSystemMetricsService{}),
	}

	return NewRouter("web-gateway", "0.1.0", controllerModule)
}

// Requirements: web-gateway/FR-001, web-gateway/OR-002
func TestRouterServesStaticAssets(t *testing.T) {
	t.Parallel()

	handler := routerFixture(nil)

	testCases := []struct {
		path            string
		wantContentType string
	}{
		{path: "/", wantContentType: "text/html; charset=utf-8"},
		{path: "/assets/index.css", wantContentType: "text/css; charset=utf-8"},
		{path: "/assets/index.js", wantContentType: "application/javascript; charset=utf-8"},
		{path: "/favicon.ico", wantContentType: "image/x-icon"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.path, func(t *testing.T) {
			t.Parallel()
			assertStaticRouteResponse(t, handler, testCase.path, testCase.wantContentType)
		})
	}
}

// Requirements: web-gateway/FR-002, web-gateway/FR-004
func TestRouterLoginReturnsCookies(t *testing.T) {
	t.Parallel()

	handler := routerFixture(nil)
	body := []byte(`{"login":"john.doe","password":"pass"}`)
	request := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}

	if len(recorder.Result().Cookies()) != 2 {
		t.Fatalf("cookie count = %d, want 2", len(recorder.Result().Cookies()))
	}
}

// Requirements: web-gateway/FR-004, web-gateway/TR-001
func TestRouterProtectedAuthEndpointRejectsMissingToken(t *testing.T) {
	t.Parallel()

	handler := routerFixture(nil)
	request := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", recorder.Code)
	}
}

func assertStaticRouteResponse(t *testing.T, handler http.Handler, path string, wantContentType string) {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, path, nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}

	if got := recorder.Header().Get("Content-Type"); got != wantContentType {
		t.Fatalf("Content-Type = %q, want %q", got, wantContentType)
	}
}

// Requirements: web-gateway/FR-004, web-gateway/TR-001
func TestRouterProtectedAuthEndpointAcceptsAccessTokenCookie(t *testing.T) {
	t.Parallel()

	handler := routerFixture(nil)
	request := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	request.AddCookie(&http.Cookie{Name: services.AccessTokenCookieName, Value: "AT-cookie"})
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
}

// Requirements: web-gateway/FR-004, web-gateway/TR-001
func TestRouterRefreshRejectsMissingRefreshToken(t *testing.T) {
	t.Parallel()

	handler := routerFixture(nil)
	request := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader([]byte(`{}`)))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", recorder.Code)
	}
}

// Requirements: web-gateway/FR-004, web-gateway/TR-001
func TestRouterLogoutAcceptsRefreshTokenPayload(t *testing.T) {
	t.Parallel()

	handler := routerFixture(nil)
	request := httptest.NewRequest(http.MethodPost, "/auth/logout", bytes.NewReader([]byte(`{"refresh_token":"RT-logout"}`)))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
}

// Requirements: web-gateway/FR-003, web-gateway/TR-001
func TestRouterSystemMetricsHistoryRequiresAuthentication(t *testing.T) {
	t.Parallel()

	handler := routerFixture(nil)
	request := httptest.NewRequest(http.MethodGet, "/system-metrics/history", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", recorder.Code)
	}
}

// Requirements: web-gateway/FR-003, web-gateway/TR-001
func TestRouterSystemMetricsHistoryReturnsJSONWhenAuthenticated(t *testing.T) {
	t.Parallel()

	handler := routerFixture(nil)
	request := httptest.NewRequest(http.MethodGet, "/system-metrics/history", nil)
	request.AddCookie(&http.Cookie{Name: services.AccessTokenCookieName, Value: "AT-cookie"})
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
}

// Requirements: web-gateway/FR-004, web-gateway/TR-001
func TestRouterMapsUnauthorizedAuthServiceError(t *testing.T) {
	t.Parallel()

	handler := routerFixture(routeAuthService{meErr: services.ErrUnauthorized})
	request := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	request.AddCookie(&http.Cookie{Name: services.AccessTokenCookieName, Value: "AT-cookie"})
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", recorder.Code)
	}
}

// Requirements: web-gateway/TR-001
func TestRouterMapsUnexpectedAuthServiceError(t *testing.T) {
	t.Parallel()

	handler := routerFixture(routeAuthService{refreshErr: errors.New("boom")})
	request := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader([]byte(`{"refresh_token":"RT-refresh"}`)))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", recorder.Code)
	}
}

// Requirements: web-gateway/IR-001
func TestRouterExposesOpenAPIDocs(t *testing.T) {
	t.Parallel()

	handler := routerFixture(nil)
	request := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
}
