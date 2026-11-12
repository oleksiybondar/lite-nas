package modules

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	sharedconfig "lite-nas/shared/config"
	"lite-nas/shared/fileio"
)

type testCoreConfig struct {
	Messaging sharedconfig.MessagingConfig
	Logging   sharedconfig.LoggingConfig
}

func TestLoadCoreInfraReturnsReaderError(t *testing.T) {
	t.Parallel()

	_, _, err := LoadCoreInfra(
		"",
		"test-service",
		func(reader fileio.Reader) (testCoreConfig, error) {
			t.Fatal("loadConfig should not be called when reader construction fails")
			return testCoreConfig{}, nil
		},
		func(serviceName string, logging sharedconfig.LoggingConfig, messaging sharedconfig.MessagingConfig) (CoreInfra, error) {
			t.Fatal("constructor should not be called when reader construction fails")
			return CoreInfra{}, nil
		},
		func(cfg testCoreConfig) sharedconfig.MessagingConfig { return cfg.Messaging },
		func(cfg testCoreConfig) sharedconfig.LoggingConfig { return cfg.Logging },
	)
	if err == nil {
		t.Fatal("LoadCoreInfra() error = nil, want reader error")
	}
}

func TestLoadCoreInfraReturnsConfigError(t *testing.T) {
	t.Parallel()

	configPath := writeCoreInfraFixture(t)
	loadErr := errors.New("load failed")

	_, _, err := LoadCoreInfra(
		configPath,
		"test-service",
		func(reader fileio.Reader) (testCoreConfig, error) {
			if _, readErr := reader.Read(); readErr != nil {
				t.Fatalf("reader.Read() error = %v", readErr)
			}
			return testCoreConfig{}, loadErr
		},
		func(serviceName string, logging sharedconfig.LoggingConfig, messaging sharedconfig.MessagingConfig) (CoreInfra, error) {
			t.Fatal("constructor should not be called when config loading fails")
			return CoreInfra{}, nil
		},
		func(cfg testCoreConfig) sharedconfig.MessagingConfig { return cfg.Messaging },
		func(cfg testCoreConfig) sharedconfig.LoggingConfig { return cfg.Logging },
	)
	if !errors.Is(err, loadErr) {
		t.Fatalf("LoadCoreInfra() error = %v, want %v", err, loadErr)
	}
}

func TestLoadCoreInfraReturnsConstructorError(t *testing.T) {
	t.Parallel()

	configPath := writeCoreInfraFixture(t)
	constructorErr := errors.New("construct failed")

	_, _, err := LoadCoreInfra(
		configPath,
		"test-service",
		func(reader fileio.Reader) (testCoreConfig, error) {
			if _, readErr := reader.Read(); readErr != nil {
				t.Fatalf("reader.Read() error = %v", readErr)
			}
			return testCoreConfig{
				Messaging: sharedconfig.MessagingConfig{URL: "nats://127.0.0.1:4222", ClientName: "test-client"},
				Logging:   sharedconfig.LoggingConfig{Level: "info", Format: "rfc5424", Output: "stdout"},
			}, nil
		},
		func(serviceName string, logging sharedconfig.LoggingConfig, messaging sharedconfig.MessagingConfig) (CoreInfra, error) {
			if serviceName != "test-service" {
				t.Fatalf("serviceName = %q, want %q", serviceName, "test-service")
			}
			return CoreInfra{}, constructorErr
		},
		func(cfg testCoreConfig) sharedconfig.MessagingConfig { return cfg.Messaging },
		func(cfg testCoreConfig) sharedconfig.LoggingConfig { return cfg.Logging },
	)
	if !errors.Is(err, constructorErr) {
		t.Fatalf("LoadCoreInfra() error = %v, want %v", err, constructorErr)
	}
}

func TestLoadCoreInfraReturnsCoreAndConfig(t *testing.T) {
	t.Parallel()

	configPath := writeCoreInfraFixture(t)
	wantCfg := testCoreConfig{
		Messaging: sharedconfig.MessagingConfig{URL: "nats://127.0.0.1:4222", ClientName: "test-client"},
		Logging:   sharedconfig.LoggingConfig{Level: "info", Format: "rfc5424", Output: "stdout"},
	}
	wantCore := CoreInfra{}

	core, cfg, err := LoadCoreInfra(
		configPath,
		"test-service",
		loadStaticCoreConfig(t, wantCfg),
		assertingCoreConstructor(t, wantCfg, wantCore),
		func(cfg testCoreConfig) sharedconfig.MessagingConfig { return cfg.Messaging },
		func(cfg testCoreConfig) sharedconfig.LoggingConfig { return cfg.Logging },
	)
	if err != nil {
		t.Fatalf("LoadCoreInfra() error = %v", err)
	}

	assertLoadedCoreConfig(t, cfg, wantCfg)
	assertZeroCoreInfra(t, core)
}

func loadStaticCoreConfig(t *testing.T, wantCfg testCoreConfig) func(fileio.Reader) (testCoreConfig, error) {
	t.Helper()

	return func(reader fileio.Reader) (testCoreConfig, error) {
		if _, readErr := reader.Read(); readErr != nil {
			t.Fatalf("reader.Read() error = %v", readErr)
		}
		return wantCfg, nil
	}
}

func assertingCoreConstructor(
	t *testing.T,
	wantCfg testCoreConfig,
	wantCore CoreInfra,
) CoreInfraConstructor {
	t.Helper()

	return func(serviceName string, logging sharedconfig.LoggingConfig, messaging sharedconfig.MessagingConfig) (CoreInfra, error) {
		if serviceName != "test-service" {
			t.Fatalf("serviceName = %q, want %q", serviceName, "test-service")
		}
		if messaging != wantCfg.Messaging {
			t.Fatalf("messaging = %#v, want %#v", messaging, wantCfg.Messaging)
		}
		if logging != wantCfg.Logging {
			t.Fatalf("logging = %#v, want %#v", logging, wantCfg.Logging)
		}
		return wantCore, nil
	}
}

func assertLoadedCoreConfig(t *testing.T, got testCoreConfig, want testCoreConfig) {
	t.Helper()

	if got != want {
		t.Fatalf("cfg = %#v, want %#v", got, want)
	}
}

func assertZeroCoreInfra(t *testing.T, got CoreInfra) {
	t.Helper()

	if got.Logger != nil || got.Client != nil || got.Server != nil || got.logCleanup != nil {
		t.Fatalf("core = %#v, want zero-value CoreInfra", got)
	}
}

func writeCoreInfraFixture(t *testing.T) string {
	t.Helper()

	configPath := filepath.Join(t.TempDir(), "config.ini")
	if err := os.WriteFile(configPath, []byte("fixture"), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", configPath, err)
	}

	return configPath
}
