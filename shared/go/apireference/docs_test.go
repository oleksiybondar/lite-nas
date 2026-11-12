package apireference

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestHandlerServesHTML verifies the shared docs response boundary.
func TestHandlerServesHTML(t *testing.T) {
	response := httptest.NewRecorder()
	Handler("API", "/openapi.json")(response, httptest.NewRequest(http.MethodGet, "/docs", nil))
	if response.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Fatalf("content type = %q", response.Header().Get("Content-Type"))
	}
}

// TestPageUsesScalarMoonTheme verifies the shared reference engine and theme.
func TestPageUsesScalarMoonTheme(t *testing.T) {
	page := Page("API", "/openapi.json")
	if !strings.Contains(page, "Scalar.createApiReference('#api-reference', {") || !strings.Contains(page, "theme: 'moon'") {
		t.Fatal("page does not initialize Scalar with the moon theme")
	}
}

// TestPageStartsInDarkMode verifies Scalar's documented initial color mode.
func TestPageStartsInDarkMode(t *testing.T) {
	if !strings.Contains(Page("API", "/openapi.json"), "darkMode: true") {
		t.Fatal("page does not start in dark mode")
	}
}

// TestPageEscapesDocumentURL keeps service configuration out of script markup.
func TestPageEscapesDocumentURL(t *testing.T) {
	if strings.Contains(Page("API", `</script><script>alert(1)</script>`), `</script><script>alert(1)</script>`) {
		t.Fatal("document URL broke out of its JavaScript string")
	}
}
