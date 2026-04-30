package config

import (
	"testing"
	"time"

	sharedconfig "lite-nas/shared/config"
	"lite-nas/shared/testutil/configtest"
	"lite-nas/shared/testutil/fileiotest"
	"lite-nas/shared/testutil/testcasetest"
)

// Requirements: auth-service/OR-002
func TestLoadConfigReturnsReaderError(t *testing.T) {
	t.Parallel()

	configtest.RunReaderErrorCase(t, LoadConfig)
}

// Requirements: auth-service/OR-002
func TestLoadConfigMessagingFields(t *testing.T) {
	t.Parallel()

	testCases := []testcasetest.FieldCase[Config]{
		{Name: "url", Got: func(cfg Config) any { return cfg.Messaging.URL }, Want: "tls://127.0.0.1:4222"},
		{Name: "client name", Got: func(cfg Config) any { return cfg.Messaging.ClientName }, Want: "auth-service"},
		{Name: "ca path", Got: func(cfg Config) any { return cfg.Messaging.CA }, Want: "/etc/lite-nas/certificates/transport/root-ca.crt"},
		{Name: "cert path", Got: func(cfg Config) any { return cfg.Messaging.Cert }, Want: "/etc/lite-nas/certificates/transport/lite-nas-auth-service/client.crt"},
		{Name: "key path", Got: func(cfg Config) any { return cfg.Messaging.Key }, Want: "/etc/lite-nas/certificates/transport/lite-nas-auth-service/client.key"},
		{Name: "timeout", Got: func(cfg Config) any { return cfg.Messaging.Timeout }, Want: 5 * time.Second},
	}

	testcasetest.RunFieldCases(t, loadConfigFixture, testCases)
}

// Requirements: auth-service/OR-002
func TestLoadConfigLoggingFields(t *testing.T) {
	t.Parallel()

	testCases := []testcasetest.FieldCase[Config]{
		{Name: "level", Got: func(cfg Config) any { return cfg.Logging.Level }, Want: "info"},
		{Name: "format", Got: func(cfg Config) any { return cfg.Logging.Format }, Want: "rfc5424"},
		{Name: "output", Got: func(cfg Config) any { return cfg.Logging.Output }, Want: "file"},
		{Name: "file path", Got: func(cfg Config) any { return cfg.Logging.FilePath }, Want: "/var/lib/lite-nas/auth-service.log"},
	}

	testcasetest.RunFieldCases(t, loadConfigFixture, testCases)
}

// Requirements: auth-service/OR-002
func TestLoadConfigAuthTokenFields(t *testing.T) {
	t.Parallel()

	testCases := configtest.AuthTokenFieldCases(
		func(cfg Config) sharedconfig.AuthTokenConfig { return cfg.AuthTokens },
		configtest.AuthTokenExpectedPaths{
			SigningKey:       "/etc/lite-nas/certificates/auth/token-signing.key",
			SigningCert:      "/etc/lite-nas/certificates/auth/token-signing.crt",
			VerificationCert: "/etc/lite-nas/certificates/auth/token-signing.crt",
		},
	)

	testcasetest.RunFieldCases(t, loadConfigFixture, testCases)
}

func loadConfigFixture(t *testing.T) Config {
	t.Helper()

	cfg, err := LoadConfig(fileiotest.Reader{
		Data: []byte(
			"[messaging]\n" +
				"url=tls://127.0.0.1:4222\n" +
				"client_name=auth-service\n" +
				"ca=/etc/lite-nas/certificates/transport/root-ca.crt\n" +
				"cert=/etc/lite-nas/certificates/transport/lite-nas-auth-service/client.crt\n" +
				"key=/etc/lite-nas/certificates/transport/lite-nas-auth-service/client.key\n" +
				"timeout=5s\n" +
				"[auth_tokens]\n" +
				"issuer=lite-nas-auth\n" +
				"audience=lite-nas-management-api\n" +
				"access_lifetime=15m\n" +
				"clock_skew=30s\n" +
				"signing_key=/etc/lite-nas/certificates/auth/token-signing.key\n" +
				"signing_cert=/etc/lite-nas/certificates/auth/token-signing.crt\n" +
				"verification_cert=/etc/lite-nas/certificates/auth/token-signing.crt\n" +
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
