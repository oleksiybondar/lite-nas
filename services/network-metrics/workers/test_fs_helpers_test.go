package workers

import (
	"testing"

	"lite-nas/shared/testutil/fstest"
)

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	fstest.MustMkdirAll(t, path, 0o750)
}

func mustSymlink(t *testing.T, target string, path string) {
	t.Helper()
	fstest.MustSymlink(t, target, path)
}
