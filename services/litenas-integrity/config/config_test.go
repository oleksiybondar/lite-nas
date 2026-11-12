package config_test

import (
	"testing"
	"time"

	serviceconfig "lite-nas/services/litenas-integrity/config"
	"lite-nas/shared/testutil/configtest"
	"lite-nas/shared/testutil/fileiotest"
)

// TestLoadConfig parses the shared runtime sections plus the required `[AIDE]`
// schedule for the integrity service.
func TestLoadConfig(t *testing.T) {
	t.Parallel()

	cfg, err := serviceconfig.LoadConfig(fileiotest.Reader{Data: []byte(validConfigFixture())})
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.AIDE.Schedule != 24*time.Hour {
		t.Fatalf("AIDE.Schedule = %v, want %v", cfg.AIDE.Schedule, 24*time.Hour)
	}
	if cfg.Messaging.ClientName != "litenas-integrity" {
		t.Fatalf("Messaging.ClientName = %q, want litenas-integrity", cfg.Messaging.ClientName)
	}
	if cfg.Auth.ServiceLogin != "lite-nas-integrity" {
		t.Fatalf("Auth.ServiceLogin = %q, want lite-nas-integrity", cfg.Auth.ServiceLogin)
	}
	if cfg.Logging.Output != "file" {
		t.Fatalf("Logging.Output = %q, want %q", cfg.Logging.Output, "file")
	}
}

// TestLoadConfigRejectsMissingSchedule ensures the dedicated AIDE schedule is
// required for service startup.
func TestLoadConfigRejectsMissingSchedule(t *testing.T) {
	t.Parallel()

	configtest.RunRejectsInvalidConfigCase(t, serviceconfig.LoadConfig, validSharedConfigFixture())
}

// TestLoadConfigRejectsNonPositiveSchedule ensures the service rejects invalid
// hourly schedule values.
func TestLoadConfigRejectsNonPositiveSchedule(t *testing.T) {
	t.Parallel()

	configtest.RunRejectsInvalidConfigCase(
		t,
		serviceconfig.LoadConfig,
		validSharedConfigFixture()+"[AIDE]\n"+"schedule=0\n",
	)
}

// TestLoadConfigReaderError preserves reader errors from the shared loader.
func TestLoadConfigReaderError(t *testing.T) {
	t.Parallel()

	configtest.RunReaderErrorCase(t, serviceconfig.LoadConfig)
}

// TestLoadConfigINIParseError rejects malformed INI content.
func TestLoadConfigINIParseError(t *testing.T) {
	t.Parallel()

	configtest.RunINIParseErrorCase(t, serviceconfig.LoadConfig)
}

// validConfigFixture returns one complete valid integrity-service config file.
func validConfigFixture() string {
	return validSharedConfigFixture() +
		"[AIDE]\n" +
		"schedule=24\n"
}

// validSharedConfigFixture returns the minimal valid shared sections needed by
// the integrity-service config loader.
func validSharedConfigFixture() string {
	return "[messaging]\n" +
		"url=tls://127.0.0.1:4222\n" +
		"client_name=litenas-integrity\n" +
		"ca=/etc/lite-nas/certificates/transport/root-ca.crt\n" +
		"cert=/etc/lite-nas/certificates/transport/lite-nas-integrity/client.crt\n" +
		"key=/etc/lite-nas/certificates/transport/lite-nas-integrity/client.key\n" +
		"timeout=5s\n" +
		"[auth]\n" +
		"ca=/etc/lite-nas/certificates/identities/root-ca.crt\n" +
		"cert=/etc/lite-nas/certificates/identities/lite-nas-integrity/client.crt\n" +
		"key=/etc/lite-nas/certificates/identities/lite-nas-integrity/client.key\n" +
		"service_name=litenas-integrity\n" +
		"service_login=lite-nas-integrity\n" +
		"[logging]\n" +
		"level=warn\n" +
		"format=rfc5424\n" +
		"output=file\n" +
		"file_path=/var/log/lite-nas/lite-nas-integrity.log\n"
}
