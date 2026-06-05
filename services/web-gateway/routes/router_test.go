package routes

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"lite-nas/services/web-gateway/controllers"
	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/services/web-gateway/modules"
	"lite-nas/services/web-gateway/services"
	webtest "lite-nas/services/web-gateway/testutil"
	"lite-nas/shared/authtoken"
	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
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

func (s routeAuthService) Login(
	_ context.Context,
	now time.Time,
	login string,
	password string,
	_ services.AuthRequestContext,
) (services.Session, error) {
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

func (s routeAuthService) Refresh(
	_ context.Context,
	now time.Time,
	refreshToken string,
	_ services.AuthRequestContext,
) (services.Session, error) {
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

func (s routeAuthService) Logout(
	_ context.Context,
	now time.Time,
	refreshToken string,
	_ services.AuthRequestContext,
) (services.Session, error) {
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

type routeZFSMetricsService struct{}

func (routeZFSMetricsService) GetSnapshot(context.Context) (metrics.ZFSSnapshot, error) {
	return metrics.ZFSSnapshot{Timestamp: time.Unix(100, 0).UTC()}, nil
}

func (routeZFSMetricsService) GetHistory(context.Context) ([]metrics.ZFSSnapshot, error) {
	return []metrics.ZFSSnapshot{
		{Timestamp: time.Unix(100, 0).UTC()},
		{Timestamp: time.Unix(101, 0).UTC()},
	}, nil
}

type routeAlertsService struct{}

func (routeAlertsService) List(context.Context, services.AlertListInput) (services.AlertListPage, error) {
	return services.AlertListPage{
		Items:      []loggingmanagercontract.ListAlertItem{{EventID: "evt-1"}},
		TotalCount: 1,
	}, nil
}

func (routeAlertsService) ListActive(context.Context, services.AlertListInput) (services.AlertListPage, error) {
	return services.AlertListPage{
		Items:      []loggingmanagercontract.ListAlertItem{{EventID: "evt-1"}},
		TotalCount: 1,
	}, nil
}

func (routeAlertsService) ListUnacknowledged(context.Context, services.AlertListInput) (services.AlertListPage, error) {
	return services.AlertListPage{
		Items:      []loggingmanagercontract.ListAlertItem{{EventID: "evt-1"}},
		TotalCount: 1,
	}, nil
}

func (routeAlertsService) Get(context.Context, services.AlertGetInput) (loggingmanagercontract.ListAlertItem, bool, error) {
	return loggingmanagercontract.ListAlertItem{EventID: "evt-1"}, true, nil
}

func (routeAlertsService) Acknowledge(context.Context, services.AlertActionInput) error {
	return nil
}

func (routeAlertsService) Mute(context.Context, services.AlertActionInput) error {
	return nil
}

type routeAuthVerifier struct {
	claims authtoken.AccessClaims
}

func (v routeAuthVerifier) Verify(string) (authtoken.AccessClaims, error) {
	claims := v.claims
	if claims.Login == "" {
		claims.Login = "john.doe"
	}
	if len(claims.Roles) == 0 {
		claims.Roles = []string{"admin"}
	}
	return claims, nil
}

func routerFixture(authService controllers.AuthService) http.Handler {
	return routerFixtureWithVerifier(authService, routeAuthVerifier{})
}

func routerFixtureWithVerifier(authService controllers.AuthService, verifier routeAuthVerifier) http.Handler {
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
		SystemAlerts:   controllers.NewSystemAlertsController(routeAlertsService{}),
		SecurityAlerts: controllers.NewSecurityAlertsController(routeAlertsService{}),
		SystemMetrics:  controllers.NewSystemMetricsController(routeSystemMetricsService{}),
		ZFSMetrics:     controllers.NewZFSMetricsController(routeZFSMetricsService{}),
	}

	return NewRouter("web-gateway", "0.1.0", controllerModule, middlewares.AuthenticationOptions{
		AccessCookieName:  services.AccessTokenCookieName,
		RefreshCookieName: services.RefreshTokenCookieName,
		Verifier:          verifier,
	})
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
		{path: "/dashboard", wantContentType: "text/html; charset=utf-8"},
		{path: "/settings/users", wantContentType: "text/html; charset=utf-8"},
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
		webtest.NewRequest(http.MethodPost, "/api/auth/login", []byte(`{"login":"john.doe","password":"pass"}`)),
	)

	webtest.AssertStatus(t, recorder, http.StatusOK)
	webtest.AssertCookieCount(t, recorder, 2)
}

// Requirements: web-gateway/FR-004, web-gateway/TR-001
func TestRouterProtectedAuthEndpointRejectsMissingToken(t *testing.T) {
	t.Parallel()

	handler := routerFixture(nil)
	recorder := webtest.ServeRequest(handler, webtest.NewRequest(http.MethodGet, "/api/auth/me", nil))
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

	assertAuthenticatedRouteStatus(t, routerFixture(nil), http.MethodGet, "/api/auth/me", nil, http.StatusOK)
}

// Requirements: web-gateway/FR-004, web-gateway/TR-001
func TestRouterRefreshRejectsMissingRefreshToken(t *testing.T) {
	t.Parallel()

	handler := routerFixture(nil)
	recorder := webtest.ServeRequest(handler, webtest.NewRequest(http.MethodPost, "/api/auth/refresh", []byte(`{}`)))
	webtest.AssertStatus(t, recorder, http.StatusUnauthorized)
}

// Requirements: web-gateway/FR-004, web-gateway/TR-001
func TestRouterLogoutAcceptsRefreshTokenCookie(t *testing.T) {
	t.Parallel()

	request := webtest.NewRequest(http.MethodPost, "/api/auth/logout", []byte(`{}`))
	request.AddCookie(&http.Cookie{
		Name:     services.RefreshTokenCookieName,
		Value:    "RT-logout",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	recorder := webtest.ServeRequest(
		routerFixture(nil),
		request,
	)
	webtest.AssertStatus(t, recorder, http.StatusOK)
}

// Requirements: web-gateway/FR-003, web-gateway/TR-001
func TestRouterSystemMetricsHistoryRequiresAuthentication(t *testing.T) {
	t.Parallel()

	handler := routerFixture(nil)
	recorder := webtest.ServeRequest(handler, webtest.NewRequest(http.MethodGet, "/api/system-metrics/history", nil))
	webtest.AssertStatus(t, recorder, http.StatusUnauthorized)
}

// Requirements: web-gateway/FR-003, web-gateway/TR-001
func TestRouterSystemMetricsHistoryReturnsJSONWhenAuthenticated(t *testing.T) {
	t.Parallel()

	assertAuthenticatedRouteStatus(t, routerFixture(nil), http.MethodGet, "/api/system-metrics/history", nil, http.StatusOK)
}

// Requirements: web-gateway/FR-003, web-gateway/TR-001
func TestRouterZFSMetricsHistoryRequiresAuthentication(t *testing.T) {
	t.Parallel()

	handler := routerFixture(nil)
	recorder := webtest.ServeRequest(handler, webtest.NewRequest(http.MethodGet, "/api/zfs-metrics/history", nil))
	webtest.AssertStatus(t, recorder, http.StatusUnauthorized)
}

// Requirements: web-gateway/FR-003, web-gateway/TR-001
func TestRouterZFSMetricsHistoryReturnsJSONWhenAuthenticated(t *testing.T) {
	t.Parallel()

	assertAuthenticatedRouteStatus(t, routerFixture(nil), http.MethodGet, "/api/zfs-metrics/history", nil, http.StatusOK)
}

// Requirements: web-gateway/FR-005, web-gateway/TR-001
func TestRouterSystemAlertsRequireAuthentication(t *testing.T) {
	t.Parallel()

	handler := routerFixture(nil)
	recorder := webtest.ServeRequest(handler, webtest.NewRequest(http.MethodGet, "/api/alerts/system", nil))
	webtest.AssertStatus(t, recorder, http.StatusUnauthorized)
}

// Requirements: web-gateway/FR-005, web-gateway/TR-001
func TestRouterSystemAlertsAcceptOperatorRole(t *testing.T) {
	t.Parallel()

	handler := routerFixtureWithVerifier(nil, routeAuthVerifier{
		claims: authtoken.AccessClaims{Login: "john.doe", Roles: []string{"lite-nas-operator"}},
	})
	recorder := webtest.ServeRequest(handler, webtest.NewAuthenticatedRequest(http.MethodGet, "/api/alerts/system/active", nil))
	webtest.AssertStatus(t, recorder, http.StatusOK)
}

// Requirements: web-gateway/FR-005, web-gateway/TR-001
func TestRouterSecurityAlertsRejectOperatorRole(t *testing.T) {
	t.Parallel()

	handler := routerFixtureWithVerifier(nil, routeAuthVerifier{
		claims: authtoken.AccessClaims{Login: "john.doe", Roles: []string{"lite-nas-operator"}},
	})
	recorder := webtest.ServeRequest(handler, webtest.NewAuthenticatedRequest(http.MethodGet, "/api/alerts/security", nil))
	webtest.AssertStatus(t, recorder, http.StatusForbidden)
}

// Requirements: web-gateway/FR-004, web-gateway/TR-001
func TestRouterMapsUnauthorizedAuthServiceError(t *testing.T) {
	t.Parallel()

	handler := routerFixture(routeAuthService{meErr: services.ErrUnauthorized})
	recorder := webtest.ServeRequest(handler, webtest.NewAuthenticatedRequest(http.MethodGet, "/api/auth/me", nil))
	webtest.AssertStatus(t, recorder, http.StatusUnauthorized)
}

// Requirements: web-gateway/TR-001
func TestRouterMapsUnexpectedAuthServiceError(t *testing.T) {
	t.Parallel()

	handler := routerFixture(routeAuthService{refreshErr: errors.New("boom")})
	request := webtest.NewRequest(http.MethodPost, "/api/auth/refresh", []byte(`{}`))
	request.AddCookie(&http.Cookie{
		Name:     services.RefreshTokenCookieName,
		Value:    "RT-refresh",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	recorder := webtest.ServeRequest(
		handler,
		request,
	)
	webtest.AssertStatus(t, recorder, http.StatusInternalServerError)
}

// Requirements: web-gateway/IR-001
func TestRouterExposesOpenAPIDocs(t *testing.T) {
	t.Parallel()

	handler := routerFixture(nil)
	recorder := webtest.ServeRequest(handler, webtest.NewRequest(http.MethodGet, "/api/openapi.json", nil))
	webtest.AssertStatus(t, recorder, http.StatusOK)
}

// Requirements: web-gateway/IR-001
func TestRouterDocumentsAlertFiltersQueryParameter(t *testing.T) {
	t.Parallel()

	handler := routerFixture(nil)
	recorder := webtest.ServeRequest(handler, webtest.NewRequest(http.MethodGet, "/api/openapi.json", nil))
	webtest.AssertStatus(t, recorder, http.StatusOK)

	var spec map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &spec); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	assertDocumentedQueryParameter(t, spec, "/alerts/system", "get", "filters")
	assertDocumentedQueryParameter(t, spec, "/alerts/system/active", "get", "filters")
	assertDocumentedQueryParameter(t, spec, "/alerts/system/unacknowledged", "get", "filters")
}

// Requirements: web-gateway/IR-001
func TestRouterDocsPageReferencesMountedOpenAPIPath(t *testing.T) {
	t.Parallel()

	handler := routerFixture(nil)
	recorder := webtest.ServeRequest(handler, webtest.NewRequest(http.MethodGet, "/api/docs", nil))
	webtest.AssertStatus(t, recorder, http.StatusOK)

	if !strings.Contains(recorder.Body.String(), `apiDescriptionUrl="/api/openapi.yaml"`) {
		t.Fatalf("docs html = %q, want mounted OpenAPI path", recorder.Body.String())
	}
}

// Requirements: web-gateway/FR-001, web-gateway/FR-002
func TestRouterDoesNotServeSPAForAPIMisses(t *testing.T) {
	t.Parallel()

	handler := routerFixture(nil)
	recorder := webtest.ServeRequest(handler, webtest.NewRequest(http.MethodGet, "/api/unknown", nil))
	webtest.AssertStatus(t, recorder, http.StatusNotFound)
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

func assertDocumentedQueryParameter(
	t *testing.T,
	spec map[string]any,
	path string,
	method string,
	parameterName string,
) {
	t.Helper()

	operation := mustOpenAPIOperation(t, spec, path, method)
	parameter := mustFindOpenAPIParameter(t, operation, path, method, parameterName)
	if parameter["description"] == "" {
		t.Fatalf("parameter %q description is empty for %s %s", parameterName, strings.ToUpper(method), path)
	}
}

func mustOpenAPIOperation(t *testing.T, spec map[string]any, path string, method string) map[string]any {
	t.Helper()

	paths := mustMapField(t, spec, "paths")
	pathItem, ok := paths[path].(map[string]any)
	if !ok {
		t.Fatalf("openapi path %q missing", path)
	}
	operation, ok := pathItem[method].(map[string]any)
	if !ok {
		t.Fatalf("openapi operation %s %s missing", strings.ToUpper(method), path)
	}
	return operation
}

func mustFindOpenAPIParameter(
	t *testing.T,
	operation map[string]any,
	path string,
	method string,
	parameterName string,
) map[string]any {
	t.Helper()

	parameters, ok := operation["parameters"].([]any)
	if !ok {
		t.Fatalf("openapi parameters for %s %s = %T, want []any", strings.ToUpper(method), path, operation["parameters"])
	}

	for _, parameter := range parameters {
		typed, ok := parameter.(map[string]any)
		if ok && typed["name"] == parameterName {
			return typed
		}
	}

	t.Fatalf("parameter %q not documented for %s %s", parameterName, strings.ToUpper(method), path)
	return nil
}

func mustMapField(t *testing.T, source map[string]any, field string) map[string]any {
	t.Helper()

	value, ok := source[field].(map[string]any)
	if !ok {
		t.Fatalf("openapi %s = %T, want map[string]any", field, source[field])
	}
	return value
}
