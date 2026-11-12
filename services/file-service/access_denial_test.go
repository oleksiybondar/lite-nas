package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/logger/formatters/rfc5424"
)

// Requirements: security-monitoring/FR-010.
func TestAccessDenialLogContainsPrincipal(t *testing.T) {
	t.Parallel()
	line := mustDenialLog(t, http.StatusForbidden, "forbidden")
	if !strings.Contains(line, "access_denied user=alice") {
		t.Fatalf("log = %q", line)
	}
}

// Requirements: security-monitoring/FR-010.
func TestAccessDenialLogExcludesServiceFailure(t *testing.T) {
	t.Parallel()
	line := mustDenialLog(t, http.StatusServiceUnavailable, "authorization_unavailable")
	if line != "" {
		t.Fatalf("unexpected log = %q", line)
	}
}

// mustDenialLog writes one response through the production log adapter.
func mustDenialLog(t *testing.T, status int, code string) string {
	t.Helper()
	var output bytes.Buffer
	log, err := sharedlogger.New(sharedlogger.Config{ServiceName: "file-service", Hostname: "nas", Writer: &output, Formatter: rfc5424.New(rfc5424.Config{})})
	if err != nil {
		t.Fatalf("create logger: %v", err)
	}
	principal := "alice"
	writer := &accessDenialWriter{ResponseWriter: httptest.NewRecorder(), logger: log, principal: &principal}
	writeError(writer, status, code)
	return output.String()
}
