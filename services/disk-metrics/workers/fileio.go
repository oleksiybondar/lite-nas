package workers

import "os"

// readTrustedFile reads files from worker-internal runtime paths and test
// fixtures that are not influenced by untrusted user input.
func readTrustedFile(path string) ([]byte, error) {
	// #nosec G304 -- worker source paths come from trusted runtime wiring or test fixtures.
	return os.ReadFile(path)
}
