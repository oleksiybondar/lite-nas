package config_test

import (
	"testing"
	"time"

	"gopkg.in/ini.v1"

	"lite-nas/shared/config"
	"lite-nas/shared/testutil/testcasetest"
)

func TestLoadAuthTokenConfigParsedFields(t *testing.T) {
	t.Parallel()

	testCases := []testcasetest.FieldCase[config.AuthTokenConfig]{
		{Name: "issuer", Got: func(cfg config.AuthTokenConfig) any { return cfg.Issuer }, Want: "lite-nas-auth"},
		{Name: "audience", Got: func(cfg config.AuthTokenConfig) any { return cfg.Audience }, Want: "lite-nas-services"},
		{Name: "access lifetime", Got: func(cfg config.AuthTokenConfig) any { return cfg.AccessLifetime }, Want: 10 * time.Minute},
		{Name: "clock skew", Got: func(cfg config.AuthTokenConfig) any { return cfg.ClockSkew }, Want: 5 * time.Second},
		{Name: "signing key", Got: func(cfg config.AuthTokenConfig) any { return cfg.SigningKey }, Want: "/tmp/token-signing.key"},
		{Name: "signing cert", Got: func(cfg config.AuthTokenConfig) any { return cfg.SigningCert }, Want: "/tmp/token-signing.crt"},
		{Name: "verification cert", Got: func(cfg config.AuthTokenConfig) any { return cfg.VerificationCert }, Want: "/tmp/token-signing.crt"},
	}

	testcasetest.RunFieldCases(t, func(t *testing.T) config.AuthTokenConfig {
		t.Helper()
		return loadAuthTokenConfigFixture(t, validAuthTokenConfigINI())
	}, testCases)
}

func TestLoadAuthTokenConfigUsesDurationDefaults(t *testing.T) {
	t.Parallel()

	cfg := loadAuthTokenConfigFixture(t,
		"[auth_tokens]\n"+
			"issuer=lite-nas-auth\n"+
			"audience=lite-nas-services\n",
	)

	if cfg.AccessLifetime != 15*time.Minute {
		t.Fatalf("AccessLifetime = %v, want 15m", cfg.AccessLifetime)
	}

	if cfg.ClockSkew != 30*time.Second {
		t.Fatalf("ClockSkew = %v, want 30s", cfg.ClockSkew)
	}
}

func TestLoadAuthTokenConfigAllowsEmptyCertificatePaths(t *testing.T) {
	t.Parallel()

	cfg := loadAuthTokenConfigFixture(t,
		"[auth_tokens]\n"+
			"issuer=lite-nas-auth\n"+
			"audience=lite-nas-services\n",
	)

	if cfg.SigningKey != "" || cfg.SigningCert != "" || cfg.VerificationCert != "" {
		t.Fatalf("certificate paths = (%q, %q, %q), want empty", cfg.SigningKey, cfg.SigningCert, cfg.VerificationCert)
	}
}

func TestLoadAuthTokenConfigRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	testCases := []string{
		"[auth_tokens]\nissuer=\naudience=lite-nas-services\n",
		"[auth_tokens]\nissuer=lite-nas-auth\naudience=\n",
		"[auth_tokens]\nissuer=lite-nas-auth\naudience=lite-nas-services\naccess_lifetime=0s\n",
		"[auth_tokens]\nissuer=lite-nas-auth\naudience=lite-nas-services\nclock_skew=-1s\n",
		"[auth_tokens]\nissuer=lite-nas-auth\naudience=lite-nas-services\naccess_lifetime=soon\n",
		"[auth_tokens]\nissuer=lite-nas-auth\naudience=lite-nas-services\nclock_skew=soon\n",
	}

	for _, testCase := range testCases {
		cfgFile, err := ini.Load([]byte(testCase))
		if err != nil {
			t.Fatalf("ini.Load() error = %v", err)
		}

		if _, err = config.LoadAuthTokenConfig(cfgFile); err == nil {
			t.Fatal("expected invalid config error")
		}
	}
}

func validAuthTokenConfigINI() string {
	return "[auth_tokens]\n" +
		"issuer= lite-nas-auth \n" +
		"audience= lite-nas-services \n" +
		"access_lifetime=10m\n" +
		"clock_skew=5s\n" +
		"signing_key= /tmp/token-signing.key \n" +
		"signing_cert= /tmp/token-signing.crt \n" +
		"verification_cert= /tmp/token-signing.crt \n"
}

func loadAuthTokenConfigFixture(t *testing.T, iniContent string) config.AuthTokenConfig {
	t.Helper()

	cfgFile, err := ini.Load([]byte(iniContent))
	if err != nil {
		t.Fatalf("ini.Load() error = %v", err)
	}

	cfg, err := config.LoadAuthTokenConfig(cfgFile)
	if err != nil {
		t.Fatalf("LoadAuthTokenConfig() error = %v", err)
	}

	return cfg
}
