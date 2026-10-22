package main

import (
	"net/http"
)

// apiDocument describes the first service routes using their public /f prefix.
var apiDocument = []byte(`{"openapi":"3.1.0","info":{"title":"LiteNAS File Service","version":"0.1.0"},"servers":[{"url":"/f"}],"paths":{"/upload":{"post":{"summary":"Upload a browser form file"}},"/proprietary/upload":{"post":{"summary":"Create a resumable upload"}},"/proprietary/upload-window":{"get":{"summary":"Inspect chunk capacity"}},"/proprietary/upload/{upload-id}":{"get":{"summary":"Get upload progress"},"head":{"summary":"Get upload headers"},"patch":{"summary":"Upload one chunk"},"delete":{"summary":"Cancel an upload"}},"/proprietary/upload/{upload-id}/completion":{"put":{"summary":"Publish a verified upload"}},"/{absolute-path}":{"get":{"summary":"Download an authorized file"},"head":{"summary":"Inspect an authorized file"},"delete":{"summary":"Delete an authorized regular file"}}}}`)

// serveDocs exposes local HTML and the public OpenAPI document when enabled.
func (s *fileService) serveDocs(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writeError(writer, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}
	if request.URL.Path == "/openapi.json" {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write(apiDocument)
		return
	}
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = writer.Write([]byte(`<!doctype html><html lang="en"><meta charset="utf-8"><title>LiteNAS File Service API</title><h1>LiteNAS File Service API</h1><p>OpenAPI: <a href="/f/openapi.json">/f/openapi.json</a></p></html>`))
}
