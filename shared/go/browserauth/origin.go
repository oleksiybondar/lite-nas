package browserauth

import (
	"net/http"
	"net/url"
)

// SameOrigin rejects unsafe browser requests sent from a different host.
// Requests without an Origin header remain available to non-browser clients.
func SameOrigin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if unsafeMethod(request.Method) && !sameOrigin(request) {
			http.Error(writer, "cross-origin request denied", http.StatusForbidden)
			return
		}
		next.ServeHTTP(writer, request)
	})
}

// unsafeMethod identifies requests that can change server state.
func unsafeMethod(method string) bool {
	return method != http.MethodGet && method != http.MethodHead && method != http.MethodOptions && method != http.MethodTrace
}

// sameOrigin checks a browser Origin against the request's proxied host.
func sameOrigin(request *http.Request) bool {
	origin := request.Header.Get("Origin")
	if origin == "" {
		return true
	}
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Host != request.Host || parsed.User != nil || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return false
	}
	requestScheme := request.Header.Get("X-Forwarded-Proto")
	if requestScheme == "" {
		requestScheme = "http"
		if request.TLS != nil {
			requestScheme = "https"
		}
	}
	return parsed.Scheme == requestScheme
}
