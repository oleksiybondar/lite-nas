package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestSwaggerDocsUseStreamSchema covers moments-backend/IR-002.
func TestSwaggerDocsUseStreamSchema(t *testing.T) {
	response := httptest.NewRecorder()
	(&streamService{}).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/docs/init.js", nil))
	if !strings.Contains(response.Body.String(), "/s/openapi.json") {
		t.Fatalf("Swagger page does not reference stream schema: %s", response.Body.String())
	}
}

// TestStreamOpenAPIListsTranscode covers moments-backend/IR-002.
func TestStreamOpenAPIListsTranscode(t *testing.T) {
	response := httptest.NewRecorder()
	(&streamService{}).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/openapi.json", nil))
	if !strings.Contains(response.Body.String(), "/s/transcode/{absolutePath}") {
		t.Fatalf("transcode route missing from schema: %s", response.Body.String())
	}
}
