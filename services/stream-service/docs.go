package main

import (
	_ "embed"
	"net/http"
)

// openAPIDocument describes the existing authenticated media routes.
//
//go:embed openapi.json
var openAPIDocument []byte

// serveOpenAPI exposes the stream contract for the bundled Swagger UI.
func serveOpenAPI(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		http.Error(writer, "method_not_allowed", http.StatusMethodNotAllowed)
		return
	}
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	writer.Header().Set("Content-Type", "application/json")
	_, _ = writer.Write(openAPIDocument)
}
