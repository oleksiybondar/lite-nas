package main

import (
	_ "embed"
	"net/http"

	"lite-nas/shared/swaggerui"
)

// apiDocument is the reviewed public API contract served at /f/openapi.json.
//
//go:embed openapi.json
var apiDocument []byte

// serveDocs exposes the authenticated schema and bundled Swagger UI assets.
func (s *fileService) serveDocs(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/openapi.json" {
		swaggerui.Serve(writer, request, "/f", "LiteNAS File Service API", "/f/openapi.json")
		return
	}
	if request.Method != http.MethodGet {
		writeError(writer, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	writer.Header().Set("Content-Type", "application/json")
	_, _ = writer.Write(apiDocument)
}
