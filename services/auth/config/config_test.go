package config

import (
	"errors"
	"testing"
	"time"
)

type fakeReader struct {
	data []byte
	err  error
}

func (r fakeReader) Read() ([]byte, error) {
	if r.err != nil {
		return nil, r.err
	}

	return r.data, nil
}

// Requirements: auth-service/OR-002
func TestLoadConfigReturnsReaderError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("read failed")

	if _, err := LoadConfig(fakeReader{err: expectedErr}); !errors.Is(err, expectedErr) {
		t.Fatalf("LoadConfig() error = %v, want %v", err, expectedErr)
	}
}

// Requirements: auth-service/OR-002
func TestLoadConfigMessagingFields(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		got  func(Config) any
		want any
	}{
		{name: "url", got: func(cfg Config) any { return cfg.Messaging.URL }, want: "tls://127.0.0.1:4222"},
		{name: "client name", got: func(cfg Config) any { return cfg.Messaging.ClientName }, want: "auth-service"},
		{name: "ca path", got: func(cfg Config) any { return cfg.Messaging.CA }, want: "/etc/lite-nas/certificates/root-ca.crt"},
		{name: "cert path", got: func(cfg Config) any { return cfg.Messaging.Cert }, want: "/etc/lite-nas/certificates/lite-nas-auth-service/client.crt"},
		{name: "key path", got: func(cfg Config) any { return cfg.Messaging.Key }, want: "/etc/lite-nas/certificates/lite-nas-auth-service/client.key"},
		{name: "timeout", got: func(cfg Config) any { return cfg.Messaging.Timeout }, want: 5 * time.Second},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cfg := loadConfigFixture(t)
			if got := testCase.got(cfg); got != testCase.want {
				t.Fatalf("%s = %#v, want %#v", testCase.name, got, testCase.want)
			}
		})
	}
}

// Requirements: auth-service/OR-002
func TestLoadConfigLoggingFields(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		got  func(Config) any
		want any
	}{
		{name: "level", got: func(cfg Config) any { return cfg.Logging.Level }, want: "info"},
		{name: "format", got: func(cfg Config) any { return cfg.Logging.Format }, want: "rfc5424"},
		{name: "output", got: func(cfg Config) any { return cfg.Logging.Output }, want: "file"},
		{name: "file path", got: func(cfg Config) any { return cfg.Logging.FilePath }, want: "/var/lib/lite-nas/auth-service.log"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cfg := loadConfigFixture(t)
			if got := testCase.got(cfg); got != testCase.want {
				t.Fatalf("%s = %#v, want %#v", testCase.name, got, testCase.want)
			}
		})
	}
}

func loadConfigFixture(t *testing.T) Config {
	t.Helper()

	cfg, err := LoadConfig(fakeReader{
		data: []byte(
			"[messaging]\n" +
				"url=tls://127.0.0.1:4222\n" +
				"client_name=auth-service\n" +
				"ca=/etc/lite-nas/certificates/root-ca.crt\n" +
				"cert=/etc/lite-nas/certificates/lite-nas-auth-service/client.crt\n" +
				"key=/etc/lite-nas/certificates/lite-nas-auth-service/client.key\n" +
				"timeout=5s\n" +
				"[logging]\n" +
				"level=info\n" +
				"format=rfc5424\n" +
				"output=file\n" +
				"file_path=/var/lib/lite-nas/auth-service.log\n",
		),
	})
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	return cfg
}
