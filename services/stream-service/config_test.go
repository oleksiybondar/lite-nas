package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// streamConfigReader supplies a controlled INI payload to the config boundary.
type streamConfigReader string

// Read returns the fixture's INI bytes.
func (reader streamConfigReader) Read() ([]byte, error) { return []byte(reader), nil }

// mustPackagedStreamConfig reads the installed Stream Service configuration fixture.
func mustPackagedStreamConfig(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("../../configs/etc/lite-nas/stream-service.conf")
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// TestPackagedStreamConfigLoads covers the installed TLS and storage boundary.
func TestPackagedStreamConfigLoads(t *testing.T) {
	config, err := loadServiceConfig(streamConfigReader(mustPackagedStreamConfig(t)))
	if err != nil {
		t.Fatal(err)
	}
	if config.Address != "127.0.0.1:8083" {
		t.Fatalf("HTTP address = %q", config.Address)
	}
}

// TestStreamConfigRejectsUnsafeOverrides covers transport and path validation.
func TestStreamConfigRejectsUnsafeOverrides(t *testing.T) {
	payload := mustPackagedStreamConfig(t)
	for _, test := range []struct {
		name string
		old  string
		new  string
	}{
		{"plain NATS", "url=tls://127.0.0.1:4222", "url=nats://127.0.0.1:4222"},
		{"wrong client", "client_name=stream-service", "client_name=other-service"},
		{"relative CA", "ca=/etc/lite-nas/certificates/transport/root-ca.crt", "ca=relative.crt"},
		{"external HTTP", "address=127.0.0.1:8083", "address=0.0.0.0:8083"},
		{"invalid port", "address=127.0.0.1:8083", "address=127.0.0.1:70000"},
		{"root read scope", "read_roots=/home", "read_roots=/"},
		{"relative read scope", "read_roots=/home", "read_roots=relative"},
		{
			"root denied path", "denied_paths=/bin,", "denied_paths=/,",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			changed := strings.Replace(payload, test.old, test.new, 1)
			if _, err := loadServiceConfig(streamConfigReader(changed)); err == nil {
				t.Fatal("unsafe Stream Service configuration was accepted")
			}
		})
	}
}

// TestRunRejectsInvalidConfigBeforeConnecting covers service startup validation.
func TestRunRejectsInvalidConfigBeforeConnecting(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root is rejected before configuration is loaded")
	}
	path := filepath.Join(t.TempDir(), "stream-service.conf")
	if err := os.WriteFile(path, []byte("not valid INI"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := run(t.Context(), path); err == nil {
		t.Fatal("service started with invalid configuration")
	}
}
