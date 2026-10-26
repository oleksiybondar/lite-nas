package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"lite-nas/shared/browserauth"
)

// authStub supplies controlled shared auth outcomes for router boundary tests.
type authStub struct {
	session browserauth.Session
	err     error
}

// Login returns the configured session or error for credential route tests.
func (stub authStub) Login(context.Context, time.Time, string, string, browserauth.AuthRequestContext) (browserauth.Session, error) {
	return stub.session, stub.err
}

// Refresh returns the configured session or error for cookie rotation tests.
func (stub authStub) Refresh(context.Context, time.Time, string, browserauth.AuthRequestContext) (browserauth.Session, error) {
	return stub.session, stub.err
}

// Logout returns the configured session or error for revocation tests.
func (stub authStub) Logout(context.Context, time.Time, string, browserauth.AuthRequestContext) (browserauth.Session, error) {
	return stub.session, stub.err
}

// Me rejects missing principals and returns configured authenticated profiles.
func (stub authStub) Me(time.Time, string) (browserauth.Session, error) {
	if stub.err != nil {
		return browserauth.Session{}, stub.err
	}
	if stub.session.UserID == "" {
		return browserauth.Session{}, browserauth.ErrUnauthorized
	}
	return stub.session, nil
}

// Requirements: moments/FR-001
func TestMeRequiresAuthentication(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	response := httptest.NewRecorder()
	newRouter(authStub{}, "").ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", response.Code)
	}
}

// TestSwaggerUIUsesGeneratedMomentsSchema covers moments-backend/IR-002.
func TestSwaggerUIUsesGeneratedMomentsSchema(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/openapi.json", nil)
	response := httptest.NewRecorder()
	store, err := openGalleryStore(filepath.Join(t.TempDir(), "moments.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.close()
	newRouterWithStore(authStub{}, "", store).ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "/api/galleries") {
		t.Fatalf("OpenAPI response = %d %q", response.Code, response.Body.String())
	}
}

// Requirements: moments/FR-002
func TestAuthMutationRejectsForeignOrigin(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	request.Header.Set("Origin", "https://foreign.example")
	response := httptest.NewRecorder()
	newRouter(authStub{}, "").ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", response.Code)
	}
}

// Requirements: moments/FR-001, browserauth/FR-003
func TestLoginUsesConfiguredCookieDomain(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"login":"alice","password":"secret"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	session := browserauth.Session{UserID: "alice", AccessToken: "access", RefreshToken: "refresh", AccessExpires: time.Now().Add(time.Hour), RefreshExpiry: time.Now().Add(24 * time.Hour)}
	newRouter(authStub{session: session}, "example.com").ServeHTTP(response, request)
	if !strings.Contains(response.Header().Get("Set-Cookie"), "Domain=example.com") {
		t.Fatalf("Set-Cookie lacks configured domain: %q", response.Header().Get("Set-Cookie"))
	}
}

// Requirements: moments/FR-001
func TestLoginReturnsGatewayCompatibleSessionData(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"login":"alice","password":"secret"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	newRouter(authStub{session: browserauth.Session{UserID: "alice"}}, "").ServeHTTP(response, request)
	var body struct {
		Data struct {
			User struct {
				ID string `json:"id"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data.User.ID != "alice" {
		t.Fatalf("user ID = %q, want alice", body.Data.User.ID)
	}
}

// Requirements: moments/IR-002
func TestListenerRejectsExternalBind(t *testing.T) {
	if err := validateListener("0.0.0.0:9091"); err == nil {
		t.Fatal("expected external listener to be rejected")
	}
}

// mustMomentsSession supplies the cookie material expected by auth route tests.
func mustMomentsSession() browserauth.Session {
	return browserauth.Session{
		UserID: "alice", Login: "alice", AuthType: "password",
		Roles: []string{"owner"}, Scopes: []string{"read"},
		AccessToken: "access", RefreshToken: "refresh",
		AccessExpires: time.Now().Add(time.Hour), RefreshExpiry: time.Now().Add(24 * time.Hour),
	}
}

// requestMomentsAuth sends a JSON auth request through the real Huma router.
func requestMomentsAuth(stub authStub, method, path, cookieName, cookieValue string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader("{}"))
	request.Header.Set("Content-Type", "application/json")
	if cookieName != "" {
		request.AddCookie(&http.Cookie{Name: cookieName, Value: cookieValue})
	}
	response := httptest.NewRecorder()
	newRouter(stub, "").ServeHTTP(response, request)
	return response
}

// Requirements: moments/FR-001. Refresh requires a cookie and rotates a valid session.
func TestRefreshCookieBoundary(t *testing.T) {
	for _, test := range []struct {
		name, cookie string
		stub         authStub
		want         int
	}{
		{"missing", "", authStub{}, http.StatusUnauthorized},
		{"valid", "refresh", authStub{session: mustMomentsSession()}, http.StatusOK},
		{"denied", "refresh", authStub{err: browserauth.ErrUnauthorized}, http.StatusUnauthorized},
		{"unavailable", "refresh", authStub{err: errors.New("RPC unavailable")}, http.StatusInternalServerError},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := requestMomentsAuth(test.stub, http.MethodPost, "/api/auth/refresh", "lite-nas-rt", test.cookie)
			if response.Code != test.want {
				t.Fatalf("refresh status = %d, want %d", response.Code, test.want)
			}
		})
	}
}

// Requirements: moments/FR-001. Logout revokes an existing refresh session.
func TestLogoutCookieBoundary(t *testing.T) {
	for _, test := range []struct {
		name, cookie string
		stub         authStub
		want         int
	}{
		{"missing", "", authStub{}, http.StatusUnauthorized},
		{"valid", "refresh", authStub{session: mustMomentsSession()}, http.StatusOK},
		{"denied", "refresh", authStub{err: browserauth.ErrUnauthorized}, http.StatusUnauthorized},
		{"unavailable", "refresh", authStub{err: errors.New("RPC unavailable")}, http.StatusInternalServerError},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := requestMomentsAuth(test.stub, http.MethodPost, "/api/auth/logout", "lite-nas-rt", test.cookie)
			if response.Code != test.want {
				t.Fatalf("logout status = %d, want %d", response.Code, test.want)
			}
		})
	}
}

// Requirements: moments/FR-001. Refresh issues a new browser access cookie.
func TestRefreshRotatesAccessCookie(t *testing.T) {
	response := requestMomentsAuth(authStub{session: mustMomentsSession()}, http.MethodPost, "/api/auth/refresh", "lite-nas-rt", "refresh")
	if !strings.Contains(response.Header().Get("Set-Cookie"), "lite-nas-at=access") {
		t.Fatalf("refresh access cookie = %q", response.Header().Get("Set-Cookie"))
	}
}

// Requirements: moments/FR-001. Logout clears the browser access cookie.
func TestLogoutClearsAccessCookie(t *testing.T) {
	response := requestMomentsAuth(authStub{session: mustMomentsSession()}, http.MethodPost, "/api/auth/logout", "lite-nas-rt", "refresh")
	if !strings.Contains(response.Header().Get("Set-Cookie"), "lite-nas-at=;") {
		t.Fatalf("logout access cookie = %q", response.Header().Get("Set-Cookie"))
	}
}

// Requirements: moments/FR-001. Me reports the verified browser principal.
func TestMeReturnsVerifiedProfile(t *testing.T) {
	response := requestMomentsAuth(authStub{session: mustMomentsSession()}, http.MethodGet, "/api/auth/me", "lite-nas-at", "access")
	var body struct {
		Data struct {
			User struct {
				ID string `json:"id"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data.User.ID != "alice" {
		t.Fatalf("me user ID = %q", body.Data.User.ID)
	}
}

// Requirements: moments/FR-001. Auth service outages fail closed.
func TestMeReportsAuthFailure(t *testing.T) {
	response := requestMomentsAuth(authStub{err: errors.New("RPC unavailable")}, http.MethodGet, "/api/auth/me", "lite-nas-at", "access")
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("me status = %d", response.Code)
	}
}

// Requirements: moments/FR-001. Login rejects an unavailable auth provider.
func TestLoginReportsAuthFailure(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"login":"alice","password":"secret"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	newRouter(authStub{err: errors.New("RPC unavailable")}, "").ServeHTTP(response, request)
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("login status = %d", response.Code)
	}
}

// TestAccessTokenPrefersBearerHeader checks explicit API token precedence.
func TestAccessTokenPrefersBearerHeader(t *testing.T) {
	for _, test := range []struct {
		name, header, cookie, want string
	}{
		{"bearer", "Bearer explicit", "cookie", "explicit"},
		{"blank bearer", "Bearer  ", "cookie", "cookie"},
		{"other scheme", "Basic ignored", "cookie", "cookie"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := accessToken(test.header, test.cookie); got != test.want {
				t.Fatalf("access token = %q, want %q", got, test.want)
			}
		})
	}
}

// TestPackagedMomentsConfigLoads covers the shipped listener and auth settings.
func TestPackagedMomentsConfigLoads(t *testing.T) {
	config, err := loadConfig("../../configs/etc/lite-nas/moments.conf")
	if err != nil {
		t.Fatal(err)
	}
	if config.HTTP.Address != "127.0.0.1:9091" {
		t.Fatalf("HTTP address = %q", config.HTTP.Address)
	}
}

// TestMomentsConfigRejectsUnsafeSettings covers invalid listener and cookie scopes.
func TestMomentsConfigRejectsUnsafeSettings(t *testing.T) {
	payload, err := os.ReadFile("../../configs/etc/lite-nas/moments.conf")
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name, old, replacement string
	}{
		{"external listener", "address=127.0.0.1:9091", "address=0.0.0.0:9091"},
		{"malformed listener", "address=127.0.0.1:9091", "address=invalid"},
		{"invalid cookie domain", "cookie_domain=", "cookie_domain=/invalid"},
		{"missing token issuer", "issuer=lite-nas-auth", "issuer="},
		{"missing HTTP address", "address=127.0.0.1:9091", "address="},
		{"external file service", "file_service_url=http://127.0.0.1:8082", "file_service_url=http://example.com:8082"},
		{"relative database path", "database_path=/var/lib/lite-nas/moments/moments.db", "database_path=relative.db"},
		{"relative thumbnail path", "thumbnail_dir=/var/lib/lite-nas/moments/thumbnails", "thumbnail_dir=thumbnails"},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "moments.conf")
			changed := strings.Replace(string(payload), test.old, test.replacement, 1)
			if err := os.WriteFile(path, []byte(changed), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := loadConfig(path); err == nil {
				t.Fatal("invalid Moments configuration was accepted")
			}
		})
	}
}

// TestMomentsConfigRejectsUnreadableInput checks failures before parsing settings.
func TestMomentsConfigRejectsUnreadableInput(t *testing.T) {
	if _, err := loadConfig(filepath.Join(t.TempDir(), "missing.conf")); err == nil {
		t.Fatal("missing Moments configuration was accepted")
	}
}

// mustMomentsRuntimeConfig points an otherwise valid service config at one cert.
func mustMomentsRuntimeConfig(t *testing.T, certificate string) string {
	t.Helper()
	payload, err := os.ReadFile("../../configs/etc/lite-nas/moments.conf")
	if err != nil {
		t.Fatal(err)
	}
	installedCertificate := "verification_cert=/etc/lite-nas/certificates/auth/token-signing.crt"
	if !strings.Contains(string(payload), installedCertificate) {
		t.Fatal("packaged Moments certificate setting is missing")
	}
	changed := strings.Replace(string(payload),
		installedCertificate,
		"verification_cert="+certificate, 1)
	path := filepath.Join(t.TempDir(), "moments.conf")
	if err := os.WriteFile(path, []byte(changed), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestMomentsRunFailsBeforeListenerOnInvalidCertificate checks startup trust.
func TestMomentsRunFailsBeforeListenerOnInvalidCertificate(t *testing.T) {
	for _, test := range []struct {
		name, certificate string
		create            bool
	}{
		{"missing", "missing.crt", false},
		{"malformed", "invalid.crt", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			certificate := filepath.Join(t.TempDir(), test.certificate)
			if test.create {
				if err := os.WriteFile(certificate, []byte("invalid certificate"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if err := run(t.Context(), mustMomentsRuntimeConfig(t, certificate)); err == nil {
				t.Fatal("Moments started without a trusted signing certificate")
			}
		})
	}
}
