package config_test

import (
	"testing"
	"time"

	"lite-nas/shared/config"

	"gopkg.in/ini.v1"
)

func TestLoadSharedAuthTokenConfigParsedFields(t *testing.T) {
	t.Parallel()

	cfgFile, err := ini.Load([]byte(sharedAuthTokenConfigFixture()))
	if err != nil {
		t.Fatalf("ini.Load() error = %v", err)
	}

	cfg, err := config.LoadSharedAuthTokenConfig(cfgFile)
	if err != nil {
		t.Fatalf("LoadSharedAuthTokenConfig() error = %v", err)
	}

	for _, testCase := range sharedAuthTokenFieldCases() {
		if got := testCase.got(cfg); got != testCase.want {
			t.Fatalf("%s = %#v, want %#v", testCase.name, got, testCase.want)
		}
	}
}

func sharedAuthTokenConfigFixture() string {
	return "[messaging]\n" +
		"url=tls://127.0.0.1:4222\n" +
		"client_name=web-gateway\n" +
		"ca=/etc/lite-nas/certificates/transport/root-ca.crt\n" +
		"cert=/etc/lite-nas/certificates/transport/lite-nas-web-gateway/client.crt\n" +
		"key=/etc/lite-nas/certificates/transport/lite-nas-web-gateway/client.key\n" +
		"timeout=5s\n" +
		"[auth_tokens]\n" +
		"issuer=lite-nas-auth\n" +
		"audience=lite-nas-management-api\n" +
		"clock_skew=30s\n" +
		"verification_cert=/etc/lite-nas/certificates/auth/token-signing.crt\n" +
		"[logging]\n" +
		"level=warn\n" +
		"format=rfc5424\n" +
		"output=file\n" +
		"file_path=/var/log/lite-nas/web-gateway.log\n"
}

type sharedAuthTokenFieldCase struct {
	name string
	got  func(config.SharedAuthTokenConfig) any
	want any
}

func sharedAuthTokenFieldCases() []sharedAuthTokenFieldCase {
	return []sharedAuthTokenFieldCase{
		{
			name: "messaging client name",
			got:  func(cfg config.SharedAuthTokenConfig) any { return cfg.Messaging.ClientName },
			want: "web-gateway",
		},
		{
			name: "auth tokens verification cert",
			got:  func(cfg config.SharedAuthTokenConfig) any { return cfg.AuthTokens.VerificationCert },
			want: "/etc/lite-nas/certificates/auth/token-signing.crt",
		},
		{
			name: "auth tokens clock skew",
			got:  func(cfg config.SharedAuthTokenConfig) any { return cfg.AuthTokens.ClockSkew },
			want: 30 * time.Second,
		},
		{
			name: "logging file path",
			got:  func(cfg config.SharedAuthTokenConfig) any { return cfg.Logging.FilePath },
			want: "/var/log/lite-nas/web-gateway.log",
		},
	}
}
