package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"lite-nas/shared/browserauth"
	authcontract "lite-nas/shared/contracts/auth"
)

// passwordRouteAuthStub supplies a pending PAM prompt to the browser routes.
type passwordRouteAuthStub struct{ authStub }

// BeginPasswordChange returns a pending password prompt without a session.
func (passwordRouteAuthStub) BeginPasswordChange(context.Context, string) (browserauth.PasswordResult, error) {
	return browserauth.PasswordResult{Status: authcontract.StatusPasswordChangeRequired}, nil
}

// ReplyPasswordChange returns the next pending prompt without a session.
func (passwordRouteAuthStub) ReplyPasswordChange(context.Context, time.Time, string, int32, string) (browserauth.PasswordResult, error) {
	return browserauth.PasswordResult{Status: authcontract.StatusPasswordChangeRequired}, nil
}

// TestPasswordChangeStartShowsPendingState covers the authenticated browser action.
func TestPasswordChangeStartShowsPendingState(t *testing.T) {
	response := httptest.NewRecorder()
	newRouter(passwordRouteAuthStub{}, "").ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/auth/password/start", strings.NewReader("{}")))
	if !strings.Contains(response.Body.String(), string(authcontract.StatusPasswordChangeRequired)) {
		t.Fatalf("password start response = %s", response.Body.String())
	}
}

// TestPasswordChangeReplyShowsPendingState covers validated PAM answers.
func TestPasswordChangeReplyShowsPendingState(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/auth/password/reply", strings.NewReader(`{"conversation_id":"pending","sequence":1,"answer":"secret"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	newRouter(passwordRouteAuthStub{}, "").ServeHTTP(response, request)
	if !strings.Contains(response.Body.String(), string(authcontract.StatusPasswordChangeRequired)) {
		t.Fatalf("password reply response = %s", response.Body.String())
	}
}
