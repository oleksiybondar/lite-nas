package main

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

// TestFileServiceStatUsesInitiatingJWT covers moments-backend/FR-003.
func TestFileServiceStatUsesInitiatingJWT(t *testing.T) {
	var authorization string
	client := fileServiceClient{origin: "http://127.0.0.1:8082", client: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		authorization = request.Header.Get("Authorization")
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"path":"/home/alice/photos","kind":"directory"}`))}, nil
	})}}
	if _, err := client.stat(context.Background(), "user-jwt", "/home/alice/photos", "directory"); err != nil {
		t.Fatal(err)
	}
	if authorization != "Bearer user-jwt" {
		t.Fatalf("authorization = %q", authorization)
	}
}
