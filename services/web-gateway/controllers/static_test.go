package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubReader struct {
	data []byte
	err  error
}

func (r stubReader) Read() ([]byte, error) {
	return r.data, r.err
}

// Requirements: web-gateway/FR-001, web-gateway/OR-002
func TestStaticControllerServeIndexReturnsHTML(t *testing.T) {
	t.Parallel()

	controller := staticControllerFixture()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()

	controller.ServeIndex(recorder, request)

	assertStaticResponse(t, recorder, "text/html; charset=utf-8")
}

// Requirements: web-gateway/FR-001, web-gateway/OR-002
func TestStaticControllerServeIndexCSSReturnsCSS(t *testing.T) {
	t.Parallel()

	controller := staticControllerFixture()
	request := httptest.NewRequest(http.MethodGet, "/assets/index.css", nil)
	recorder := httptest.NewRecorder()

	controller.ServeIndexCSS(recorder, request)

	assertStaticResponse(t, recorder, "text/css; charset=utf-8")
}

// Requirements: web-gateway/FR-001, web-gateway/OR-002
func TestStaticControllerServeIndexJSReturnsJavaScript(t *testing.T) {
	t.Parallel()

	controller := staticControllerFixture()
	request := httptest.NewRequest(http.MethodGet, "/assets/index.js", nil)
	recorder := httptest.NewRecorder()

	controller.ServeIndexJS(recorder, request)

	assertStaticResponse(t, recorder, "application/javascript; charset=utf-8")
}

// Requirements: web-gateway/FR-001, web-gateway/OR-002
func TestStaticControllerServeFaviconReturnsIcon(t *testing.T) {
	t.Parallel()

	controller := staticControllerFixture()
	request := httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)
	recorder := httptest.NewRecorder()

	controller.ServeFavicon(recorder, request)

	assertStaticResponse(t, recorder, "image/x-icon")
}
