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

// Requirements: web-gateway/OR-001
func TestLoadConfigReturnsReaderError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("read failed")

	if _, err := LoadConfig(fakeReader{err: expectedErr}); !errors.Is(err, expectedErr) {
		t.Fatalf("LoadConfig() error = %v, want %v", err, expectedErr)
	}
}

// Requirements: web-gateway/OR-001
func TestLoadConfigHTTPFields(t *testing.T) {
	t.Parallel()

	cfg := loadConfigFixture(t)
	if cfg.HTTP.Address != "127.0.0.1:9191" {
		t.Fatalf("cfg.HTTP.Address = %q, want %q", cfg.HTTP.Address, "127.0.0.1:9191")
	}
}

// Requirements: web-gateway/IR-002
func TestLoadConfigMessagingFields(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		got  func(Config) any
		want any
	}{
		{name: "url", got: func(cfg Config) any { return cfg.Messaging.URL }, want: "tls://127.0.0.1:4222"},
		{name: "client name", got: func(cfg Config) any { return cfg.Messaging.ClientName }, want: "web-gateway"},
		{name: "ca path", got: func(cfg Config) any { return cfg.Messaging.CA }, want: "/etc/lite-nas/certificates/root-ca.crt"},
		{name: "cert path", got: func(cfg Config) any { return cfg.Messaging.Cert }, want: "/etc/lite-nas/certificates/lite-nas-web-gateway/client.crt"},
		{name: "key path", got: func(cfg Config) any { return cfg.Messaging.Key }, want: "/etc/lite-nas/certificates/lite-nas-web-gateway/client.key"},
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

// Requirements: web-gateway/OR-001
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
		{name: "file path", got: func(cfg Config) any { return cfg.Logging.FilePath }, want: "/var/lib/lite-nas/web-gateway.log"},
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

// Requirements: web-gateway/OR-001
func TestLoadConfigRejectsInvalidHTTPValues(t *testing.T) {
	t.Parallel()

	reader := fakeReader{
		data: []byte("[http]\naddress=   \n"),
	}

	if _, err := LoadConfig(reader); err == nil {
		t.Fatal("expected invalid http error")
	}
}

// Requirements: web-gateway/OR-001
func TestLoadConfigRejectsInvalidLoggingValues(t *testing.T) {
	t.Parallel()

	reader := fakeReader{
		data: []byte(
			"[http]\n" +
				"address=127.0.0.1:9191\n" +
				"[logging]\n" +
				"output=file\n",
		),
	}

	if _, err := LoadConfig(reader); err == nil {
		t.Fatal("expected invalid logging error")
	}
}

func loadConfigFixture(t *testing.T) Config {
	t.Helper()

	cfg, err := LoadConfig(fakeReader{
		data: []byte(
			"[http]\n" +
				"address=127.0.0.1:9191\n" +
				"[messaging]\n" +
				"url=tls://127.0.0.1:4222\n" +
				"client_name=web-gateway\n" +
				"ca=/etc/lite-nas/certificates/root-ca.crt\n" +
				"cert=/etc/lite-nas/certificates/lite-nas-web-gateway/client.crt\n" +
				"key=/etc/lite-nas/certificates/lite-nas-web-gateway/client.key\n" +
				"timeout=5s\n" +
				"[logging]\n" +
				"level=info\n" +
				"format=rfc5424\n" +
				"output=file\n" +
				"file_path=/var/lib/lite-nas/web-gateway.log\n",
		),
	})
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	return cfg
}
