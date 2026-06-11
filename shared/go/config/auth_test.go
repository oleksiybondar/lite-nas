package config_test

import (
	"testing"

	"gopkg.in/ini.v1"

	"lite-nas/shared/config"
	"lite-nas/shared/testutil/testcasetest"
)

func TestLoadAuthConfigParsedFields(t *testing.T) {
	t.Parallel()

	testCases := []testcasetest.FieldCase[config.AuthConfig]{
		{Name: "ca", Got: func(cfg config.AuthConfig) any { return cfg.CA }, Want: "/etc/lite-nas/certificates/identities/root-ca.crt"},
		{Name: "cert", Got: func(cfg config.AuthConfig) any { return cfg.Cert }, Want: "/etc/lite-nas/certificates/identities/lite-nas-resources-monitor/client.crt"},
		{Name: "key", Got: func(cfg config.AuthConfig) any { return cfg.Key }, Want: "/etc/lite-nas/certificates/identities/lite-nas-resources-monitor/client.key"},
		{Name: "service name", Got: func(cfg config.AuthConfig) any { return cfg.ServiceName }, Want: "resources-monitor"},
		{Name: "service login", Got: func(cfg config.AuthConfig) any { return cfg.ServiceLogin }, Want: "lite-nas-resources-monitor"},
	}

	testcasetest.RunFieldCases(t, func(t *testing.T) config.AuthConfig {
		t.Helper()
		return loadAuthConfigFixture(t, validAuthConfigINI())
	}, testCases)
}

func loadAuthConfigFixture(t *testing.T, iniContent string) config.AuthConfig {
	t.Helper()

	cfgFile, err := ini.Load([]byte(iniContent))
	if err != nil {
		t.Fatalf("ini.Load() error = %v", err)
	}

	cfg, err := config.LoadAuthConfig(cfgFile)
	if err != nil {
		t.Fatalf("LoadAuthConfig() error = %v", err)
	}

	return cfg
}

func validAuthConfigINI() string {
	return "[auth]\n" +
		"ca=/etc/lite-nas/certificates/identities/root-ca.crt\n" +
		"cert=/etc/lite-nas/certificates/identities/lite-nas-resources-monitor/client.crt\n" +
		"key=/etc/lite-nas/certificates/identities/lite-nas-resources-monitor/client.key\n" +
		"service_name=resources-monitor\n" +
		"service_login=lite-nas-resources-monitor\n"
}

func TestLoadAuthConfigFallsBackToRootCA(t *testing.T) {
	t.Parallel()

	cfg := loadAuthConfigFixture(t,
		"[auth]\n"+
			"root_ca=/etc/lite-nas/certificates/identities/root-ca.crt\n"+
			"cert=/etc/lite-nas/certificates/identities/lite-nas-resources-monitor/client.crt\n"+
			"key=/etc/lite-nas/certificates/identities/lite-nas-resources-monitor/client.key\n"+
			"service_name=resources-monitor\n"+
			"service_login=lite-nas-resources-monitor\n",
	)
	if cfg.CA != "/etc/lite-nas/certificates/identities/root-ca.crt" {
		t.Fatalf("cfg.CA = %q", cfg.CA)
	}
}
