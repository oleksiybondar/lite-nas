package routes

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"lite-nas/services/web-gateway/controllers"
	"lite-nas/services/web-gateway/modules"
	"lite-nas/services/web-gateway/services"
	webtest "lite-nas/services/web-gateway/testutil"
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
	recorder := webtest.ServeRequest(
		handler,
		webtest.NewRequest(http.MethodPost, "/auth/login", []byte(`{"login":"john.doe","password":"pass"}`)),
	)

	webtest.AssertStatus(t, recorder, http.StatusOK)
	webtest.AssertCookieCount(t, recorder, 2)
}

// Requirements: web-gateway/FR-004, web-gateway/TR-001
func TestRouterProtectedAuthEndpointRejectsMissingToken(t *testing.T) {
	t.Parallel()

	handler := routerFixture(nil)
	recorder := webtest.ServeRequest(handler, webtest.NewRequest(http.MethodGet, "/auth/me", nil))
	webtest.AssertStatus(t, recorder, http.StatusUnauthorized)
}

func assertStaticRouteResponse(t *testing.T, handler http.Handler, path string, wantContentType string) {
	t.Helper()

	recorder := webtest.ServeRequest(handler, webtest.NewRequest(http.MethodGet, path, nil))
	webtest.AssertStatus(t, recorder, http.StatusOK)
	webtest.AssertContentType(t, recorder, wantContentType)
}

// Requirements: web-gateway/FR-004, web-gateway/TR-001
func TestRouterProtectedAuthEndpointAcceptsAccessTokenCookie(t *testing.T) {
	t.Parallel()

	assertAuthenticatedRouteStatus(t, routerFixture(nil), http.MethodGet, "/auth/me", nil, http.StatusOK)
}

// Requirements: web-gateway/FR-004, web-gateway/TR-001
func TestRouterRefreshRejectsMissingRefreshToken(t *testing.T) {
	t.Parallel()

	handler := routerFixture(nil)
	recorder := webtest.ServeRequest(handler, webtest.NewRequest(http.MethodPost, "/auth/refresh", []byte(`{}`)))
	webtest.AssertStatus(t, recorder, http.StatusUnauthorized)
}

// Requirements: web-gateway/FR-004, web-gateway/TR-001
func TestRouterLogoutAcceptsRefreshTokenPayload(t *testing.T) {
	t.Parallel()

	recorder := webtest.ServeRequest(
		routerFixture(nil),
		webtest.NewRequest(http.MethodPost, "/auth/logout", []byte(`{"refresh_token":"RT-logout"}`)),
	)
	webtest.AssertStatus(t, recorder, http.StatusOK)
}

// Requirements: web-gateway/FR-003, web-gateway/TR-001
func TestRouterSystemMetricsHistoryRequiresAuthentication(t *testing.T) {
	t.Parallel()

	handler := routerFixture(nil)
	recorder := webtest.ServeRequest(handler, webtest.NewRequest(http.MethodGet, "/system-metrics/history", nil))
	webtest.AssertStatus(t, recorder, http.StatusUnauthorized)
}

// Requirements: web-gateway/FR-003, web-gateway/TR-001
func TestRouterSystemMetricsHistoryReturnsJSONWhenAuthenticated(t *testing.T) {
	t.Parallel()

	assertAuthenticatedRouteStatus(t, routerFixture(nil), http.MethodGet, "/system-metrics/history", nil, http.StatusOK)
}

// Requirements: web-gateway/FR-004, web-gateway/TR-001
func TestRouterMapsUnauthorizedAuthServiceError(t *testing.T) {
	t.Parallel()

	handler := routerFixture(routeAuthService{meErr: services.ErrUnauthorized})
	recorder := webtest.ServeRequest(handler, webtest.NewAuthenticatedRequest(http.MethodGet, "/auth/me", nil))
	webtest.AssertStatus(t, recorder, http.StatusUnauthorized)
}

// Requirements: web-gateway/TR-001
func TestRouterMapsUnexpectedAuthServiceError(t *testing.T) {
	t.Parallel()

	handler := routerFixture(routeAuthService{refreshErr: errors.New("boom")})
	recorder := webtest.ServeRequest(
		handler,
		webtest.NewRequest(http.MethodPost, "/auth/refresh", []byte(`{"refresh_token":"RT-refresh"}`)),
	)
	webtest.AssertStatus(t, recorder, http.StatusInternalServerError)
}

// Requirements: web-gateway/IR-001
func TestRouterExposesOpenAPIDocs(t *testing.T) {
	t.Parallel()

	handler := routerFixture(nil)
	recorder := webtest.ServeRequest(handler, webtest.NewRequest(http.MethodGet, "/openapi.json", nil))
	webtest.AssertStatus(t, recorder, http.StatusOK)
}

func assertAuthenticatedRouteStatus(
	t *testing.T,
	handler http.Handler,
	method string,
	path string,
	body []byte,
	wantStatus int,
) {
	t.Helper()

	recorder := webtest.ServeRequest(handler, webtest.NewAuthenticatedRequest(method, path, body))
	webtest.AssertStatus(t, recorder, wantStatus)
}
