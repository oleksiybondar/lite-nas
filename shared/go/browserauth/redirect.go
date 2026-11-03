package browserauth

import (
	"net/url"
	"strings"
	"unicode"
)

const maxRedirectPathLength = 2048

// RedirectPath returns an app-local login destination or the safe root fallback.
// Browser apps remain responsible for navigation after the JSON login response.
func RedirectPath(raw string) string {
	if raw == "" || len(raw) > maxRedirectPathLength || !strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, "//") || strings.Contains(raw, `\`) {
		return "/"
	}
	for _, character := range raw {
		if unicode.IsControl(character) {
			return "/"
		}
	}
	target, err := url.Parse(raw)
	if err != nil || target.IsAbs() || target.Host != "" || target.User != nil || !strings.HasPrefix(target.Path, "/") || strings.HasPrefix(target.Path, "//") || strings.Contains(target.Path, `\`) || strings.TrimRight(target.Path, "/") == "/login" {
		return "/"
	}
	for _, character := range target.Path {
		if unicode.IsControl(character) {
			return "/"
		}
	}
	return target.String()
}
