package main

import (
	"context"
	"net/http"
	"testing"
	"time"
)

type recordingRuntimeLogger struct {
	infoCount int
}

func (l *recordingRuntimeLogger) Info(string, ...any) {
	l.infoCount++
}

func (l *recordingRuntimeLogger) Error(string, ...any) {}

func TestServeHTTPReturnsServerListenError(t *testing.T) {
	t.Parallel()

	server := &http.Server{
		Addr:              "127.0.0.1:-1",
		Handler:           http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
		ReadHeaderTimeout: time.Second,
	}

	err := serveHTTP(context.Background(), server, &recordingRuntimeLogger{})
	if err == nil {
		t.Fatal("serveHTTP() error = nil, want error")
	}
}

func TestShutdownHTTPServerReturnsNilForIdleServer(t *testing.T) {
	t.Parallel()

	server := &http.Server{
		Addr:              "127.0.0.1:0",
		Handler:           http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
		ReadHeaderTimeout: time.Second,
	}
	log := &recordingRuntimeLogger{}

	if err := shutdownHTTPServer(server, log); err != nil {
		t.Fatalf("shutdownHTTPServer() error = %v, want nil", err)
	}

	if log.infoCount != 1 {
		t.Fatalf("infoCount = %d, want 1", log.infoCount)
	}
}
