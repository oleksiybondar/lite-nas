package browserauth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Requirements: browserauth/FR-004
func TestSameOrigin(t *testing.T) {
	for _, test := range []struct {
		name, method, origin, proto string
		status                      int
	}{
		{"same host", http.MethodPost, "https://moments.example.com", "", http.StatusNoContent},
		{"foreign host", http.MethodPost, "https://evil.example.com", "", http.StatusForbidden},
		{"foreign sibling", http.MethodPut, "https://nas.example.com", "", http.StatusForbidden},
		{"wrong scheme", http.MethodPost, "http://moments.example.com", "", http.StatusForbidden},
		{"non-browser client", http.MethodDelete, "", "", http.StatusNoContent},
		{"read request", http.MethodGet, "https://evil.example.com", "", http.StatusNoContent},
		{"tunnel HTTPS", http.MethodPost, "https://moments.example.com", "https", http.StatusNoContent},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, "https://moments.example.com/api/auth/me", nil)
			if test.origin != "" {
				request.Header.Set("Origin", test.origin)
			}
			if test.proto != "" {
				request.Header.Set("X-Forwarded-Proto", test.proto)
			}
			response := httptest.NewRecorder()
			SameOrigin(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) { writer.WriteHeader(http.StatusNoContent) })).ServeHTTP(response, request)
			if response.Code != test.status {
				t.Fatalf("status = %d, want %d", response.Code, test.status)
			}
		})
	}
}
