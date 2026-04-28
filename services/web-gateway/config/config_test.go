package config

import (
	"testing"
	"time"

	sharedconfig "lite-nas/shared/config"
	"lite-nas/shared/testutil/configtest"
	"lite-nas/shared/testutil/fileiotest"
	"lite-nas/shared/testutil/testcasetest"
)

// Requirements: web-gateway/OR-001
func TestLoadConfigReturnsReaderError(t *testing.T) {
	t.Parallel()

	configtest.RunReaderErrorCase(t, LoadConfig)
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

	testCases := []testcasetest.FieldCase[Config]{
		{Name: "url", Got: func(cfg Config) any { return cfg.Messaging.URL }, Want: "tls://127.0.0.1:4222"},
		{Name: "client name", Got: func(cfg Config) any { return cfg.Messaging.ClientName }, Want: "web-gateway"},
		{Name: "ca path", Got: func(cfg Config) any { return cfg.Messaging.CA }, Want: "/etc/lite-nas/certificates/transport/root-ca.crt"},
		{Name: "cert path", Got: func(cfg Config) any { return cfg.Messaging.Cert }, Want: "/etc/lite-nas/certificates/transport/lite-nas-web-gateway/client.crt"},
		{Name: "key path", Got: func(cfg Config) any { return cfg.Messaging.Key }, Want: "/etc/lite-nas/certificates/transport/lite-nas-web-gateway/client.key"},
		{Name: "timeout", Got: func(cfg Config) any { return cfg.Messaging.Timeout }, Want: 5 * time.Second},
	}

	testcasetest.RunFieldCases(t, loadConfigFixture, testCases)
}

// Requirements: web-gateway/OR-001
func TestLoadConfigLoggingFields(t *testing.T) {
	t.Parallel()

	testCases := []testcasetest.FieldCase[Config]{
		{Name: "level", Got: func(cfg Config) any { return cfg.Logging.Level }, Want: "info"},
		{Name: "format", Got: func(cfg Config) any { return cfg.Logging.Format }, Want: "rfc5424"},
		{Name: "output", Got: func(cfg Config) any { return cfg.Logging.Output }, Want: "file"},
		{Name: "file path", Got: func(cfg Config) any { return cfg.Logging.FilePath }, Want: "/var/lib/lite-nas/web-gateway.log"},
	}

	testcasetest.RunFieldCases(t, loadConfigFixture, testCases)
}

// Requirements: web-gateway/OR-001
func TestLoadConfigAuthTokenFields(t *testing.T) {
	t.Parallel()

	testCases := append(configtest.AuthTokenFieldCases(func(cfg Config) sharedconfig.AuthTokenConfig {
		return cfg.AuthTokens
	}), []testcasetest.FieldCase[Config]{
		{Name: "signing key", Got: func(cfg Config) any { return cfg.AuthTokens.SigningKey }, Want: ""},
		{Name: "signing cert", Got: func(cfg Config) any { return cfg.AuthTokens.SigningCert }, Want: ""},
		{Name: "verification cert", Got: func(cfg Config) any { return cfg.AuthTokens.VerificationCert }, Want: "/etc/lite-nas/certificates/auth/token-signing.crt"},
	}...)

	testcasetest.RunFieldCases(t, loadConfigFixture, testCases)
}

// Requirements: web-gateway/OR-001
func TestLoadConfigRejectsInvalidHTTPValues(t *testing.T) {
	t.Parallel()

	configtest.RunRejectsInvalidConfigCase(
		t,
		LoadConfig,
		"[http]\naddress=   \n",
	)
}

// Requirements: web-gateway/OR-001
func TestLoadConfigRejectsInvalidLoggingValues(t *testing.T) {
	t.Parallel()

	configtest.RunRejectsInvalidConfigCase(
		t,
		LoadConfig,
		"[http]\n"+
			"address=127.0.0.1:9191\n"+
			"[logging]\n"+
			"output=file\n",
	)
}

func loadConfigFixture(t *testing.T) Config {
	t.Helper()

	cfg, err := LoadConfig(fileiotest.Reader{
		Data: []byte(
			"[http]\n" +
				"address=127.0.0.1:9191\n" +
				"[messaging]\n" +
				"url=tls://127.0.0.1:4222\n" +
				"client_name=web-gateway\n" +
				"ca=/etc/lite-nas/certificates/transport/root-ca.crt\n" +
				"cert=/etc/lite-nas/certificates/transport/lite-nas-web-gateway/client.crt\n" +
				"key=/etc/lite-nas/certificates/transport/lite-nas-web-gateway/client.key\n" +
				"timeout=5s\n" +
				"[auth_tokens]\n" +
				"issuer=lite-nas-auth\n" +
				"audience=lite-nas-services\n" +
				"access_lifetime=15m\n" +
				"clock_skew=30s\n" +
				"verification_cert=/etc/lite-nas/certificates/auth/token-signing.crt\n" +
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
