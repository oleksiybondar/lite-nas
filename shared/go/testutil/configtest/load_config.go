package configtest

import (
	"errors"
	"testing"

	"lite-nas/shared/fileio"
	"lite-nas/shared/testutil/fileiotest"
)

// RunReaderErrorCase verifies that a config loader surfaces the underlying
// reader error unchanged.
func RunReaderErrorCase[T any](
	t *testing.T,
	load func(fileio.Reader) (T, error),
) {
	t.Helper()

	expectedErr := errors.New("read failed")

	if _, err := load(fileiotest.Reader{Err: expectedErr}); !errors.Is(err, expectedErr) {
		t.Fatalf("load(...) error = %v, want %v", err, expectedErr)
	}
}

// RunRejectsInvalidConfigCase verifies that a config loader rejects invalid
// configuration content.
func RunRejectsInvalidConfigCase[T any](
	t *testing.T,
	load func(fileio.Reader) (T, error),
	iniData string,
) {
	t.Helper()

	if _, err := load(fileiotest.Reader{Data: []byte(iniData)}); err == nil {
		t.Fatal("expected invalid config error")
	}
}
