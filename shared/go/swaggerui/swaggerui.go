// Package swaggerui serves the bundled Swagger UI from LiteNAS service binaries.
package swaggerui

import (
	"embed"
	"fmt"
	"html"
	"net/http"
	"strings"
)

// assets is the Apache-2.0 Swagger UI 5.33.1 distribution and its notices.
//
//go:embed assets/*
var assets embed.FS

// Serve handles a service's internal /docs route using its external mount path.
// The document URL must be a fixed service-owned path, never request input.
func Serve(writer http.ResponseWriter, request *http.Request, mount, title, documentURL string) {
	if request.Method != http.MethodGet {
		http.Error(writer, "method_not_allowed", http.StatusMethodNotAllowed)
		return
	}
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	writer.Header().Set("Referrer-Policy", "same-origin")
	path := strings.TrimPrefix(request.URL.Path, mount)
	if !strings.HasPrefix(path, "/docs") {
		path = request.URL.Path
	}
	switch path {
	case "/docs":
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		writer.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'self'; style-src 'self' 'unsafe-inline'; connect-src 'self'; img-src 'self' data:; font-src 'self' data:")
		_, _ = fmt.Fprintf(writer, `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>%s</title><link rel="stylesheet" href="%s/docs/swagger-ui.css"></head><body><div id="swagger-ui"></div><footer><a href="%s/docs/LICENSE">Swagger UI license</a> · <a href="%s/docs/NOTICE">notice</a></footer><script src="%s/docs/swagger-ui-bundle.js" defer></script><script src="%s/docs/init.js" defer></script></body></html>`, html.EscapeString(title), mount, mount, mount, mount, mount)
	case "/docs/init.js":
		writer.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		_, _ = fmt.Fprintf(writer, `window.ui = SwaggerUIBundle({url: %q, dom_id: "#swagger-ui", deepLinking: true, displayRequestDuration: true, persistAuthorization: false, validatorUrl: null, presets: [SwaggerUIBundle.presets.apis], requestInterceptor: function (request) { request.credentials = "same-origin"; return request; }});`, documentURL)
	case "/docs/swagger-ui.css", "/docs/swagger-ui-bundle.js", "/docs/LICENSE", "/docs/NOTICE", "/docs/swagger-ui-bundle.js.LICENSE.txt":
		name := "assets/" + strings.TrimPrefix(path, "/docs/")
		data, err := assets.ReadFile(name)
		if err != nil {
			http.NotFound(writer, request)
			return
		}
		switch {
		case strings.HasSuffix(name, ".css"):
			writer.Header().Set("Content-Type", "text/css; charset=utf-8")
		case strings.HasSuffix(name, ".js"):
			writer.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		default:
			writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		}
		_, _ = writer.Write(data)
	default:
		http.NotFound(writer, request)
	}
}
