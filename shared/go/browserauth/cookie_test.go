package browserauth

import (
	"net/http"
	"testing"
	"time"
)

// Requirements: browserauth/FR-002
func TestCookieDomainValidation(t *testing.T) {
	for _, test := range []struct {
		domain string
		valid  bool
	}{
		{"", true},
		{"example.com", true},
		{"nas.example.com", true},
		{"localhost", false},
		{"127.0.0.1", false},
		{"example.com; Secure", false},
		{".example.com", false},
		{"-example.com", false},
		{"example..com", false},
	} {
		t.Run(test.domain, func(t *testing.T) {
			if got := ValidateCookieDomain(test.domain) == nil; got != test.valid {
				t.Fatalf("ValidateCookieDomain(%q) valid = %v, want %v", test.domain, got, test.valid)
			}
		})
	}
}

// Requirements: browserauth/FR-003
func TestCookieIssueAndExpiryUseSameScope(t *testing.T) {
	now := time.Now()
	session := Session{AccessToken: "token", AccessExpires: now.Add(time.Hour)}
	issued := AccessCookie(session, "example.com")
	expired := ClearAccessCookie(now, "example.com")
	if issued.Name != expired.Name || issued.Domain != expired.Domain || issued.Path != expired.Path || issued.Secure != expired.Secure || issued.HttpOnly != expired.HttpOnly || issued.SameSite != http.SameSiteLaxMode || expired.SameSite != http.SameSiteLaxMode {
		t.Fatal("issued and expired cookies have different scopes")
	}
}
