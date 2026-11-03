package configtest

import (
	"testing"

	"lite-nas/shared/fileio"
	"lite-nas/shared/testutil/fileiotest"
)

// MustLoadConfig loads configuration for a success-path fixture and fails the
// test immediately when the loader returns an error.
func MustLoadConfig[T any](
	t *testing.T,
	load func(fileio.Reader) (T, error),
	iniData string,
) T {
	t.Helper()

	cfg, err := load(fileiotest.Reader{Data: []byte(iniData)})
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	return cfg
}

// MetricsServiceSharedConfigFixture returns the common messaging, logging, and
// auth sections used by metrics service config tests.
func MetricsServiceSharedConfigFixture(clientName string) string {
	return "[messaging]\n" +
		"url = nats://127.0.0.1:4222\n" +
		"client_name = " + clientName + "\n" +
		"timeout = 3s\n\n" +
		"[logging]\n" +
		"level = info\n" +
		"format = rfc5424\n" +
		"output = stdout\n\n" +
		"[auth]\n"
}
