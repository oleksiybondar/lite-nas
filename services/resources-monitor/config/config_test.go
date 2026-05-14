package config

import (
	"testing"
	"time"

	"lite-nas/shared/testutil/configtest"
	"lite-nas/shared/testutil/fileiotest"
	"lite-nas/shared/testutil/testcasetest"
)

func TestLoadConfigReturnsReaderError(t *testing.T) {
	t.Parallel()

	configtest.RunReaderErrorCase(t, LoadConfig)
}

func TestLoadConfigMessagingFields(t *testing.T) {
	t.Parallel()

	testCases := []testcasetest.FieldCase[Config]{
		{Name: "url", Got: func(cfg Config) any { return cfg.Messaging.URL }, Want: "nats://localhost:4222"},
		{Name: "client name", Got: func(cfg Config) any { return cfg.Messaging.ClientName }, Want: "resources-monitor"},
		{Name: "ca path", Got: func(cfg Config) any { return cfg.Messaging.CA }, Want: "/etc/lite-nas/certificates/transport/root-ca.crt"},
		{Name: "cert path", Got: func(cfg Config) any { return cfg.Messaging.Cert }, Want: "/etc/lite-nas/certificates/transport/lite-nas-resources-monitor/client.crt"},
		{Name: "key path", Got: func(cfg Config) any { return cfg.Messaging.Key }, Want: "/etc/lite-nas/certificates/transport/lite-nas-resources-monitor/client.key"},
		{Name: "timeout", Got: func(cfg Config) any { return cfg.Messaging.Timeout }, Want: 9 * time.Second},
	}

	testcasetest.RunFieldCases(t, loadConfigFixture, testCases)
}

func TestLoadConfigRulesFields(t *testing.T) {
	t.Parallel()

	testCases := []testcasetest.FieldCase[Config]{
		{Name: "first rules file", Got: func(cfg Config) any { return cfg.Rules.Files[0] }, Want: "/etc/lite-nas/rules/system-resources.json"},
		{Name: "second rules file", Got: func(cfg Config) any { return cfg.Rules.Files[1] }, Want: "/etc/lite-nas/rules/network-resources.json"},
	}

	testcasetest.RunFieldCases(t, loadConfigFixture, testCases)
}

func TestLoadConfigLoggingFields(t *testing.T) {
	t.Parallel()

	testCases := []testcasetest.FieldCase[Config]{
		{Name: "level", Got: func(cfg Config) any { return cfg.Logging.Level }, Want: "debug"},
		{Name: "output", Got: func(cfg Config) any { return cfg.Logging.Output }, Want: "file"},
		{Name: "file path", Got: func(cfg Config) any { return cfg.Logging.FilePath }, Want: "/var/log/lite-nas/resources-monitor.log"},
	}

	testcasetest.RunFieldCases(t, loadConfigFixture, testCases)
}

func TestLoadConfigRejectsMissingRulesFiles(t *testing.T) {
	t.Parallel()

	configtest.RunRejectsInvalidConfigCase(
		t,
		LoadConfig,
		"[messaging]\n"+
			"url=nats://localhost:4222\n"+
			"[logging]\n"+
			"level=info\n"+
			"format=rfc5424\n"+
			"output=stdout\n"+
			"[rules]\n",
	)
}

func loadConfigFixture(t *testing.T) Config {
	t.Helper()

	cfg, err := LoadConfig(fileiotest.Reader{
		Data: []byte(
			"[rules]\n" +
				"files=/etc/lite-nas/rules/system-resources.json,/etc/lite-nas/rules/network-resources.json\n" +
				"[messaging]\n" +
				"url=nats://localhost:4222\n" +
				"client_name=resources-monitor\n" +
				"ca=/etc/lite-nas/certificates/transport/root-ca.crt\n" +
				"cert=/etc/lite-nas/certificates/transport/lite-nas-resources-monitor/client.crt\n" +
				"key=/etc/lite-nas/certificates/transport/lite-nas-resources-monitor/client.key\n" +
				"timeout=9s\n" +
				"[logging]\n" +
				"level=debug\n" +
				"format=rfc5424\n" +
				"output=file\n" +
				"file_path=/var/log/lite-nas/resources-monitor.log\n",
		),
	})
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	return cfg
}
