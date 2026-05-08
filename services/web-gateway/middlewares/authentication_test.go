package middlewares

import "testing"

func TestExtractBearerTokenReturnsToken(t *testing.T) {
	t.Parallel()

	got := extractBearerToken("Bearer AT-123")
	if got != "AT-123" {
		t.Fatalf("extractBearerToken() = %q, want %q", got, "AT-123")
	}
}

func TestExtractBearerTokenRejectsUnsupportedHeader(t *testing.T) {
	t.Parallel()

	got := extractBearerToken("Basic abc")
	if got != "" {
		t.Fatalf("extractBearerToken() = %q, want empty string", got)
	}
}
