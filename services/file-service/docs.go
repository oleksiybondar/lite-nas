package main

import (
	"embed"
	"net/http"
)

// apiDocument is the reviewed public API contract served at /f/openapi.json.
//
//go:embed openapi.json
var apiDocument []byte

// swaggerAssets contains the Apache-2.0 Swagger UI 5.33.1 distribution.
// The matching LICENSE, NOTICE, and bundle license notices are kept with it.
//
//go:embed swagger-ui/*
var swaggerAssets embed.FS

// swaggerPage loads only local assets and this service's authenticated schema.
const swaggerPage = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>LiteNAS File Service API</title>
  <link rel="stylesheet" href="/f/docs/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <footer><a href="/f/docs/LICENSE">Swagger UI license</a> · <a href="/f/docs/NOTICE">notice</a></footer>
  <script src="/f/docs/swagger-ui-bundle.js" defer></script>
  <script src="/f/docs/init.js" defer></script>
</body>
</html>`

// swaggerInit configures same-origin requests and never persists authorization.
const swaggerInit = `window.ui = SwaggerUIBundle({
  url: "/f/openapi.json",
  dom_id: "#swagger-ui",
  deepLinking: true,
  displayRequestDuration: true,
  persistAuthorization: false,
  validatorUrl: null,
  presets: [SwaggerUIBundle.presets.apis],
  requestInterceptor: function (request) {
    request.credentials = "same-origin";
    return request;
  }
});`

// serveDocs exposes the authenticated schema and bundled Swagger UI assets.
func (s *fileService) serveDocs(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writeError(writer, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	writer.Header().Set("Referrer-Policy", "same-origin")
	switch request.URL.Path {
	case "/openapi.json":
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write(apiDocument)
	case "/docs":
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		writer.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'self'; style-src 'self' 'unsafe-inline'; connect-src 'self'; img-src 'self' data:; font-src 'self' data:")
		_, _ = writer.Write([]byte(swaggerPage))
	case "/docs/init.js":
		writer.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		_, _ = writer.Write([]byte(swaggerInit))
	case "/docs/swagger-ui.css", "/docs/swagger-ui-bundle.js",
		"/docs/LICENSE", "/docs/NOTICE", "/docs/swagger-ui-bundle.js.LICENSE.txt":
		s.serveSwaggerAsset(writer, request.URL.Path)
	default:
		writeError(writer, http.StatusNotFound, "not_found")
	}
}

// serveSwaggerAsset sends one bundled file from a fixed authenticated route.
func (s *fileService) serveSwaggerAsset(writer http.ResponseWriter, path string) {
	name := "swagger-ui/" + path[len("/docs/"):]
	data, err := swaggerAssets.ReadFile(name)
	if err != nil {
		writeError(writer, http.StatusNotFound, "not_found")
		return
	}
	if path == "/docs/swagger-ui.css" {
		writer.Header().Set("Content-Type", "text/css; charset=utf-8")
	} else if path == "/docs/swagger-ui-bundle.js" {
		writer.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	} else {
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}
	_, _ = writer.Write(data)
}
