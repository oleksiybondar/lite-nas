package testutil

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"lite-nas/services/web-gateway/services"
)

func NewRequest(method string, target string, body []byte) *http.Request {
	request := httptest.NewRequest(method, target, bytes.NewReader(body))
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	return request
}

func NewAuthenticatedRequest(method string, target string, body []byte) *http.Request {
	request := NewRequest(method, target, body)
	request.AddCookie(&http.Cookie{
		Name:  services.AccessTokenCookieName,
		Value: "AT-cookie",
	})

	return request
}

func ServeRequest(handler http.Handler, request *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func AssertStatus(t *testing.T, recorder *httptest.ResponseRecorder, want int) {
	t.Helper()

	if recorder.Code != want {
		t.Fatalf("status = %d, want %d", recorder.Code, want)
	}
}

func AssertContentType(t *testing.T, recorder *httptest.ResponseRecorder, want string) {
	t.Helper()

	if got := recorder.Header().Get("Content-Type"); got != want {
		t.Fatalf("Content-Type = %q, want %q", got, want)
	}
}

func AssertCookieCount(t *testing.T, recorder *httptest.ResponseRecorder, want int) {
	t.Helper()

	if got := len(recorder.Result().Cookies()); got != want {
		t.Fatalf("cookie count = %d, want %d", got, want)
	}
}
