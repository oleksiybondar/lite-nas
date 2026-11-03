package browserauth

import (
	"strings"
	"testing"
)

// TestRedirectPathAcceptsOnlyLocalRoutes covers safe login return targets.
func TestRedirectPathAcceptsOnlyLocalRoutes(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name string
		path string
		want string
	}{
		{name: "gallery", path: "/galleries/abc?view=wall#photo", want: "/galleries/abc?view=wall#photo"},
		{name: "empty", path: "", want: "/"},
		{name: "remote", path: "https://evil.example/path", want: "/"},
		{name: "network path", path: "//evil.example/path", want: "/"},
		{name: "backslash", path: `/\evil.example/path`, want: "/"},
		{name: "encoded backslash", path: "/%5Cevil.example/path", want: "/"},
		{name: "encoded newline", path: "/system%0A/alerts", want: "/"},
		{name: "login loop", path: "/login?redirect_to=%2Fsystem", want: "/"},
		{name: "login trailing slash", path: "/login/", want: "/"},
		{name: "control", path: "/system\n/alerts", want: "/"},
		{name: "oversized", path: "/" + strings.Repeat("a", maxRedirectPathLength), want: "/"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if got := RedirectPath(testCase.path); got != testCase.want {
				t.Fatalf("RedirectPath() = %q, want %q", got, testCase.want)
			}
		})
	}
}
