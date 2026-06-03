package middlewares

import (
	"context"
	"crypto/tls"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/golang-jwt/jwt/v5"

	"lite-nas/shared/authtoken"
)

type stubHumaContext struct {
	baseContext context.Context
	request     *http.Request
	status      int
	headers     http.Header
	body        io.Writer
}

func newStubHumaContext(request *http.Request) *stubHumaContext {
	return &stubHumaContext{
		baseContext: request.Context(),
		request:     request,
		headers:     make(http.Header),
		body:        io.Discard,
	}
}

func (c *stubHumaContext) Operation() *huma.Operation { return &huma.Operation{} }
func (c *stubHumaContext) Context() context.Context   { return c.baseContext }
func (c *stubHumaContext) TLS() *tls.ConnectionState  { return nil }
func (c *stubHumaContext) Version() huma.ProtoVersion { return huma.ProtoVersion{} }
func (c *stubHumaContext) Method() string             { return c.request.Method }
func (c *stubHumaContext) Host() string               { return c.request.Host }
func (c *stubHumaContext) RemoteAddr() string         { return c.request.RemoteAddr }
func (c *stubHumaContext) URL() url.URL               { return *c.request.URL }
func (c *stubHumaContext) Param(string) string        { return "" }
func (c *stubHumaContext) Query(name string) string   { return c.request.URL.Query().Get(name) }
func (c *stubHumaContext) Header(name string) string  { return c.request.Header.Get(name) }
func (c *stubHumaContext) EachHeader(cb func(name, value string)) {
	for name, values := range c.request.Header {
		for _, value := range values {
			cb(name, value)
		}
	}
}
func (c *stubHumaContext) BodyReader() io.Reader { return c.request.Body }
func (c *stubHumaContext) GetMultipartForm() (*multipart.Form, error) {
	return nil, http.ErrNotMultipart
}
func (c *stubHumaContext) SetReadDeadline(time.Time) error { return nil }
func (c *stubHumaContext) SetStatus(code int)              { c.status = code }
func (c *stubHumaContext) Status() int                     { return c.status }
func (c *stubHumaContext) SetHeader(name, value string)    { c.headers.Set(name, value) }
func (c *stubHumaContext) AppendHeader(name, value string) { c.headers.Add(name, value) }
func (c *stubHumaContext) BodyWriter() io.Writer           { return c.body }

type stubAPI struct{}

func (stubAPI) Adapter() huma.Adapter                                   { return nil }
func (stubAPI) OpenAPI() *huma.OpenAPI                                  { return &huma.OpenAPI{} }
func (stubAPI) Negotiate(string) (string, error)                        { return "application/json", nil }
func (stubAPI) Transform(_ huma.Context, _ string, v any) (any, error)  { return v, nil }
func (stubAPI) Marshal(io.Writer, string, any) error                    { return nil }
func (stubAPI) Unmarshal(string, []byte, any) error                     { return nil }
func (stubAPI) UseMiddleware(...func(huma.Context, func(huma.Context))) {}
func (stubAPI) Middlewares() huma.Middlewares                           { return nil }

type authenticationVerifierStub struct {
	err    error
	claims authtoken.AccessClaims
}

func (v authenticationVerifierStub) Verify(string) (authtoken.AccessClaims, error) {
	if v.err != nil {
		return authtoken.AccessClaims{}, v.err
	}

	return v.claims, nil
}

func TestExtractAuthenticationPrefersBearerHeader(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer AT-header")
	request.AddCookie(authenticationCookieFixture("lite-nas-at", "AT-cookie"))

	ctx := newStubHumaContext(request)
	middleware := ExtractAuthentication(authenticationOptionsFixture())

	nextCalled := false
	middleware(ctx, func(nextCtx huma.Context) {
		nextCalled = true
		assertAuthenticatedContext(t, nextCtx, []string{"lite-nas-operator"})
	})

	if !nextCalled {
		t.Fatal("expected next middleware to be called")
	}
}

func TestExtractAuthenticationFallsBackToCookie(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.AddCookie(authenticationCookieFixture("lite-nas-at", "AT-cookie"))

	ctx := newStubHumaContext(request)
	middleware := ExtractAuthentication(authenticationOptionsFixture())

	middleware(ctx, func(nextCtx huma.Context) {
		if !hasAccessToken(nextCtx.Context()) {
			t.Fatal("expected access token in context")
		}
	})
}

func TestExtractAuthenticationLeavesContextUntouchedWithoutToken(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := newStubHumaContext(request)
	middleware := ExtractAuthentication(authenticationOptionsFixture())

	middleware(ctx, func(nextCtx huma.Context) {
		if hasAccessToken(nextCtx.Context()) {
			t.Fatal("expected no access token in context")
		}
	})
}

func TestRequireAuthenticationRejectsMissingToken(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := newStubHumaContext(request)
	middleware := RequireAuthentication(stubAPI{}, authenticationOptionsFixture())

	nextCalled := false
	middleware(ctx, func(huma.Context) {
		nextCalled = true
	})

	if nextCalled {
		t.Fatal("expected request to be rejected")
	}
}

func TestRequireAuthenticationAllowsRequestWithToken(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := huma.WithValue(newStubHumaContext(request), accessTokenContextKey{}, "AT-token")
	middleware := RequireAuthentication(stubAPI{}, authenticationOptionsFixture())

	nextCalled := false
	middleware(ctx, func(huma.Context) {
		nextCalled = true
	})

	if !nextCalled {
		t.Fatal("expected request to continue")
	}
}

func TestRequireAuthenticationClearsCookiesForExpiredToken(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	baseCtx := newStubHumaContext(request)
	ctx := huma.WithValue(
		baseCtx,
		authenticationFailureContextKey{},
		authenticationFailureExpired,
	)
	middleware := RequireAuthentication(stubAPI{}, authenticationOptionsFixture())

	middleware(ctx, func(huma.Context) {
		t.Fatal("expected request to be rejected")
	})

	cookies := baseCtx.headers.Values("Set-Cookie")
	if len(cookies) != 2 {
		t.Fatalf("Set-Cookie count = %d, want 2", len(cookies))
	}
}

func TestExtractAuthenticationMarksExpiredToken(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.AddCookie(authenticationCookieFixture("lite-nas-at", "expired-token"))
	ctx := newStubHumaContext(request)
	options := authenticationOptionsFixture()
	options.Verifier = authenticationVerifierStub{err: jwt.ErrTokenExpired}

	ExtractAuthentication(options)(ctx, func(nextCtx huma.Context) {
		if !hasExpiredAuthentication(nextCtx.Context()) {
			t.Fatal("expected expired authentication marker")
		}
	})
}

func authenticationOptionsFixture() AuthenticationOptions {
	return AuthenticationOptions{
		AccessCookieName:  "lite-nas-at",
		RefreshCookieName: "lite-nas-rt",
		Verifier: authenticationVerifierStub{
			claims: authtoken.AccessClaims{Roles: []string{"lite-nas-operator"}},
		},
	}
}

// authenticationCookieFixture returns a browser-equivalent auth cookie for middleware tests.
func authenticationCookieFixture(name string, value string) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}

func TestExtractAccessTokenReturnsEmptyWhenCookieMissing(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := newStubHumaContext(request)

	if got := extractAccessToken(ctx, "lite-nas-at"); got != "" {
		t.Fatalf("extractAccessToken() = %q, want empty string", got)
	}
}

func TestIsAcceptedAccessTokenRequiresNonEmptyValue(t *testing.T) {
	t.Parallel()

	if isAcceptedAccessToken("   ") {
		t.Fatal("expected whitespace token to be rejected")
	}

	if !isAcceptedAccessToken("AT-token") {
		t.Fatal("expected token to be accepted")
	}
}

func TestRequireOperatorAllowsOperatorRole(t *testing.T) {
	t.Parallel()

	assertRoleMiddlewareAllowed(t, RequireOperator(stubAPI{}), []string{"lite-nas-operator"})
}

func TestRequireOperatorAllowsAdministratorRole(t *testing.T) {
	t.Parallel()

	assertRoleMiddlewareAllowed(t, RequireOperator(stubAPI{}), []string{"sudo"})
}

func TestRequireOperatorRejectsAuthenticatedNonOperator(t *testing.T) {
	t.Parallel()

	assertRoleMiddlewareRejected(t, RequireOperator(stubAPI{}), []string{"lite-nas-security"}, http.StatusForbidden)
}

func TestRequireAdministratorAllowsAdministratorRole(t *testing.T) {
	t.Parallel()

	assertRoleMiddlewareAllowed(t, RequireAdministrator(stubAPI{}), []string{"admin"})
}

func TestRequireAdministratorRejectsOperatorRole(t *testing.T) {
	t.Parallel()

	assertRoleMiddlewareRejected(t, RequireAdministrator(stubAPI{}), []string{"lite-nas-operator"}, http.StatusForbidden)
}

func TestRequireSecurityAllowsSecurityRole(t *testing.T) {
	t.Parallel()

	assertRoleMiddlewareAllowed(t, RequireSecurity(stubAPI{}), []string{"lite-nas-security"})
}

func TestRequireSecurityAllowsAdministratorRole(t *testing.T) {
	t.Parallel()

	assertRoleMiddlewareAllowed(t, RequireSecurity(stubAPI{}), []string{"sudo"})
}

func TestRequireSecurityRejectsAuthenticatedNonSecurity(t *testing.T) {
	t.Parallel()

	assertRoleMiddlewareRejected(t, RequireSecurity(stubAPI{}), []string{"lite-nas-operator"}, http.StatusForbidden)
}

func TestRequireAnyRoleRejectsMissingClaims(t *testing.T) {
	t.Parallel()

	assertRoleMiddlewareRejected(t, RequireAnyRole(stubAPI{}, operatorRoles, administratorRoles), nil, http.StatusUnauthorized)
}

func TestHasAnyRoleMatchesCaseInsensitiveRole(t *testing.T) {
	t.Parallel()

	if !hasAnyRole([]string{" SUDO "}, administratorRoles) {
		t.Fatal("expected normalized administrator role to match")
	}
}

func assertRoleMiddlewareAllowed(
	t *testing.T,
	middleware func(huma.Context, func(huma.Context)),
	roles []string,
) {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := authenticatedContextFixture(request, roles)

	nextCalled := false
	middleware(ctx, func(huma.Context) {
		nextCalled = true
	})

	if !nextCalled {
		t.Fatal("expected request to continue")
	}
}

func assertRoleMiddlewareRejected(
	t *testing.T,
	middleware func(huma.Context, func(huma.Context)),
	roles []string,
	wantStatus int,
) {
	t.Helper()

	ctx, statusCtx := roleContextFixture(roles)

	nextCalled := false
	middleware(ctx, func(huma.Context) {
		nextCalled = true
	})

	if nextCalled {
		t.Fatal("expected request to be rejected")
	}

	assertResponseStatus(t, statusCtx, wantStatus)
}

func authenticatedContextFixture(request *http.Request, roles []string) huma.Context {
	baseCtx := newStubHumaContext(request)
	return withAuthenticatedContext(
		baseCtx,
		"AT-token",
		authtoken.AccessClaims{Roles: roles},
	)
}

func assertAuthenticatedContext(t *testing.T, ctx huma.Context, wantRoles []string) {
	t.Helper()

	if !hasAccessToken(ctx.Context()) {
		t.Fatal("expected access token in context")
	}

	claims, ok := accessClaimsFromContext(ctx.Context())
	if !ok {
		t.Fatal("expected access claims in context")
	}
	if len(claims.Roles) != len(wantRoles) {
		t.Fatalf("claims roles = %#v, want %#v", claims.Roles, wantRoles)
	}
	for index, wantRole := range wantRoles {
		if claims.Roles[index] != wantRole {
			t.Fatalf("claims roles = %#v, want %#v", claims.Roles, wantRoles)
		}
	}
}

func roleContextFixture(roles []string) (huma.Context, *stubHumaContext) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	baseCtx := newStubHumaContext(request)
	if roles == nil {
		return baseCtx, baseCtx
	}

	return withAuthenticatedContext(
		baseCtx,
		"AT-token",
		authtoken.AccessClaims{Roles: roles},
	), baseCtx
}

func assertResponseStatus(t *testing.T, ctx *stubHumaContext, wantStatus int) {
	t.Helper()

	if ctx.status != wantStatus {
		t.Fatalf("status = %d, want %d", ctx.status, wantStatus)
	}
}
