package middlewares

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
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

var errStubTransformUnsupported = errors.New("stub transform unsupported")

func (stubAPI) Adapter() huma.Adapter            { return nil }
func (stubAPI) OpenAPI() *huma.OpenAPI           { return &huma.OpenAPI{} }
func (stubAPI) Negotiate(string) (string, error) { return "application/json", nil }
func (stubAPI) Transform(huma.Context, string, any) (any, error) {
	return nil, errStubTransformUnsupported
}
func (stubAPI) Marshal(io.Writer, string, any) error                    { return nil }
func (stubAPI) Unmarshal(string, []byte, any) error                     { return nil }
func (stubAPI) UseMiddleware(...func(huma.Context, func(huma.Context))) {}
func (stubAPI) Middlewares() huma.Middlewares                           { return nil }

func TestExtractAuthenticationPrefersBearerHeader(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer AT-header")
	request.AddCookie(&http.Cookie{Name: "lite-nas-at", Value: "AT-cookie"})

	ctx := newStubHumaContext(request)
	middleware := ExtractAuthentication("lite-nas-at")

	nextCalled := false
	middleware(ctx, func(nextCtx huma.Context) {
		nextCalled = true
		if !hasAccessToken(nextCtx.Context()) {
			t.Fatal("expected access token in context")
		}
	})

	if !nextCalled {
		t.Fatal("expected next middleware to be called")
	}
}

func TestExtractAuthenticationFallsBackToCookie(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.AddCookie(&http.Cookie{Name: "lite-nas-at", Value: "AT-cookie"})

	ctx := newStubHumaContext(request)
	middleware := ExtractAuthentication("lite-nas-at")

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
	middleware := ExtractAuthentication("lite-nas-at")

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
	middleware := RequireAuthentication(stubAPI{})

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
	middleware := RequireAuthentication(stubAPI{})

	nextCalled := false
	middleware(ctx, func(huma.Context) {
		nextCalled = true
	})

	if !nextCalled {
		t.Fatal("expected request to continue")
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
