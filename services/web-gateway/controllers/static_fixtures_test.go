package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	sharedlogger "lite-nas/shared/logger"
)

func staticControllerFixture() StaticController {
	return NewStaticController(
		StaticFiles{
			IndexHTML: stubReader{data: []byte("<html>ok</html>")},
			IndexCSS:  stubReader{data: []byte("body {}")},
			IndexJS:   stubReader{data: []byte("console.log('ok')")},
			Favicon:   stubReader{data: []byte{0x00, 0x00, 0x01, 0x00}},
		},
		sharedlogger.NewNop(),
	)
}

func assertStaticResponse(
	t *testing.T,
	recorder *httptest.ResponseRecorder,
	wantContentType string,
) {
	t.Helper()

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}

	if got := recorder.Header().Get("Content-Type"); got != wantContentType {
		t.Fatalf("Content-Type = %q, want %q", got, wantContentType)
	}
}
