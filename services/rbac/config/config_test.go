package config

import (
	"testing"
	"time"

	"lite-nas/shared/testutil/configtest"
	"lite-nas/shared/testutil/fileiotest"
)

func TestLoadConfigReturnsReaderError(t *testing.T) {
	t.Parallel()

	configtest.RunReaderErrorCase(t, LoadConfig)
}

func TestLoadConfigDoesNotRequireAuthSection(t *testing.T) {
	t.Parallel()

	cfg, err := LoadConfig(fileiotest.Reader{Data: []byte(validConfigWithoutAuthSection())})
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.Messaging.ClientName != "rbac-service" {
		t.Fatalf("cfg.Messaging.ClientName = %q", cfg.Messaging.ClientName)
	}
	assertDefaultCacheConfig(t, cfg)
}

func TestLoadConfigAppliesCacheDefaultsWhenSectionMissing(t *testing.T) {
	t.Parallel()

	cfg, err := LoadConfig(fileiotest.Reader{Data: []byte(configWithoutCacheSection())})
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	assertDefaultCacheConfig(t, cfg)
}

func assertDefaultCacheConfig(t *testing.T, cfg Config) {
	t.Helper()

	if cfg.Cache.InvalidateInterval != time.Hour {
		t.Fatalf("cfg.Cache.InvalidateInterval = %s", cfg.Cache.InvalidateInterval)
	}
	if cfg.Cache.RealUserTTL != 24*time.Hour {
		t.Fatalf("cfg.Cache.RealUserTTL = %s", cfg.Cache.RealUserTTL)
	}
	if cfg.Cache.NonInteractiveUserTTL != 7*24*time.Hour {
		t.Fatalf("cfg.Cache.NonInteractiveUserTTL = %s", cfg.Cache.NonInteractiveUserTTL)
	}
}

func validConfigWithoutAuthSection() string {
	return "[messaging]\n" +
		"url=tls://127.0.0.1:4222\n" +
		"client_name=rbac-service\n" +
		"ca=/etc/lite-nas/certificates/transport/root-ca.crt\n" +
		"cert=/etc/lite-nas/certificates/transport/lite-nas-rbac-service/client.crt\n" +
		"key=/etc/lite-nas/certificates/transport/lite-nas-rbac-service/client.key\n" +
		"timeout=5s\n" +
		"[logging]\n" +
		"level=warn\n" +
		"format=rfc5424\n" +
		"output=file\n" +
		"file_path=/var/log/lite-nas/rbac-service.log\n" +
		"[cache]\n" +
		"invalidate_interval=1h\n" +
		"real_user_ttl=24h\n" +
		"non_interactive_user_ttl=168h\n"
}

func configWithoutCacheSection() string {
	return "[messaging]\n" +
		"url=tls://127.0.0.1:4222\n" +
		"client_name=rbac-service\n" +
		"ca=/etc/lite-nas/certificates/transport/root-ca.crt\n" +
		"cert=/etc/lite-nas/certificates/transport/lite-nas-rbac-service/client.crt\n" +
		"key=/etc/lite-nas/certificates/transport/lite-nas-rbac-service/client.key\n" +
		"timeout=5s\n" +
		"[logging]\n" +
		"level=warn\n" +
		"format=rfc5424\n" +
		"output=file\n" +
		"file_path=/var/log/lite-nas/rbac-service.log\n"
}
