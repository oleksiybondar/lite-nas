package configtest

import (
	"strings"
	"testing"

	"gopkg.in/ini.v1"
)

// RunINILoadRejectsCase verifies that a parsed INI payload is rejected by the
// provided section loader and that the returned error includes the expected
// substring.
func RunINILoadRejectsCase[T any](
	t *testing.T,
	load func(*ini.File) (T, error),
	iniData string,
	wantSubstring string,
) {
	t.Helper()

	cfgFile, err := ini.Load([]byte(iniData))
	if err != nil {
		t.Fatalf("ini.Load() error = %v", err)
	}

	_, err = load(cfgFile)
	if err == nil {
		t.Fatal("expected validation error")
	}

	if !strings.Contains(err.Error(), wantSubstring) {
		t.Fatalf("error = %q, want substring %q", err, wantSubstring)
	}
}
