package config_test

import (
	"reflect"
	"testing"

	"gopkg.in/ini.v1"

	"lite-nas/shared/config"
	"lite-nas/shared/testutil/configtest"
)

func TestLoadEmailConfigParsedFields(t *testing.T) {
	t.Parallel()

	cfg := mustLoadEmailConfig(
		t,
		"[email]\n"+
			"to=a@example.com,b@example.com\n"+
			"cc=c@example.com, d@example.com\n"+
			"from=system-alert-notifier@lite-nas.com\n"+
			"subject_prefix=[LiteNAS]\n",
	)

	assertStringSliceEqual(t, cfg.To, []string{"a@example.com", "b@example.com"}, "cfg.To")
	assertStringSliceEqual(t, cfg.CC, []string{"c@example.com", "d@example.com"}, "cfg.CC")
	assertStringEqual(t, cfg.From, "system-alert-notifier@lite-nas.com", "cfg.From")
	assertStringEqual(t, cfg.SubjectPrefix, "[LiteNAS]", "cfg.SubjectPrefix")
}

func TestLoadEmailConfigAllowsEmptyRecipientLists(t *testing.T) {
	t.Parallel()

	cfg := mustLoadEmailConfig(t, "[email]\nfrom=security-alert-notifier@lite-nas.com\n")

	if len(cfg.To) != 0 {
		t.Fatalf("len(cfg.To) = %d, want 0", len(cfg.To))
	}

	if len(cfg.CC) != 0 {
		t.Fatalf("len(cfg.CC) = %d, want 0", len(cfg.CC))
	}
}

func TestLoadEmailConfigRejectsBlankFrom(t *testing.T) {
	t.Parallel()

	configtest.RunINILoadRejectsCase(t, config.LoadEmailConfig, "[email]\nfrom=   \n", "email from is required")
}

func mustLoadEmailConfig(t *testing.T, raw string) config.EmailConfig {
	t.Helper()

	cfgFile, err := ini.Load([]byte(raw))
	if err != nil {
		t.Fatalf("ini.Load() error = %v", err)
	}

	cfg, err := config.LoadEmailConfig(cfgFile)
	if err != nil {
		t.Fatalf("LoadEmailConfig() error = %v", err)
	}

	return cfg
}

func assertStringSliceEqual(t *testing.T, got []string, want []string, fieldName string) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s = %#v, want %#v", fieldName, got, want)
	}
}

func assertStringEqual(t *testing.T, got string, want string, fieldName string) {
	t.Helper()

	if got != want {
		t.Fatalf("%s = %q, want %q", fieldName, got, want)
	}
}
