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

// TestMissingRequiresReadableParent distinguishes a deleted path from denied access.
func TestMissingRequiresReadableParent(t *testing.T) {
	for _, test := range []struct {
		name       string
		parentCode int
		fileCode   int
		want       bool
	}{
		{name: "deleted file", parentCode: http.StatusOK, fileCode: http.StatusNotFound, want: true},
		{name: "denied file", parentCode: http.StatusOK, fileCode: http.StatusForbidden, want: false},
		{name: "file became directory", parentCode: http.StatusOK, fileCode: http.StatusOK, want: true},
		{name: "unreadable parent", parentCode: http.StatusForbidden, fileCode: http.StatusNotFound, want: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			client := fileServiceClient{origin: "http://127.0.0.1:8082", client: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
				status := test.fileCode
				body := ""
				if request.URL.Path == "/stat/home/alice" {
					status = test.parentCode
					body = `{"path":"/home/alice","kind":"directory"}`
				} else if status == http.StatusOK {
					body = `{"path":"/home/alice/photo.jpg","kind":"directory"}`
				}
				return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body))}, nil
			})}}
			missing, err := client.missing(context.Background(), "user-jwt", "/home/alice/photo.jpg")
			if test.name == "unreadable parent" {
				if err == nil {
					t.Fatal("unreadable parent was treated as confirmed absence")
				}
				return
			}
			if err != nil || missing != test.want {
				t.Fatalf("missing = %v, error = %v, want %v", missing, err, test.want)
			}
		})
	}
}
