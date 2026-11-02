package backuprestore

import (
	"archive/tar"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// Requirements: backup-restore/FR-003.
func TestTailscaleExportProducesReplayableJSON(t *testing.T) {
	plugin := &TailscalePlugin{run: func(arguments ...string) ([]byte, error) {
		if !reflect.DeepEqual(arguments, []string{"get", "--set-flags"}) {
			t.Fatalf("arguments = %v", arguments)
		}
		return []byte("--accept-dns=false --ssh=true\n"), nil
	}}
	data, err := plugin.Export()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := decodeTailscaleState(data); err != nil {
		t.Fatalf("exported JSON is not replayable: %v", err)
	}
}

// Requirements: backup-restore/SR-002.
func TestTailscaleApplyPassesValidatedArgumentsWithoutShell(t *testing.T) {
	var got []string
	plugin := &TailscalePlugin{run: func(arguments ...string) ([]byte, error) {
		if arguments[0] == "get" {
			return []byte("--accept-dns=true --ssh=false"), nil
		}
		got = append([]string(nil), arguments...)
		return nil, nil
	}}
	if err := plugin.Apply([]byte(`{"schemaVersion":1,"flags":["--accept-dns=false","--ssh=true"]}`)); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []string{"set", "--accept-dns=false", "--ssh=true"}) {
		t.Fatalf("arguments = %v", got)
	}
}

// Requirements: backup-restore/FR-003.
func TestTailscaleApplyAcceptsInstalledCLIFlagAddedAfterPluginBuild(t *testing.T) {
	var applied []string
	plugin := &TailscalePlugin{run: func(arguments ...string) ([]byte, error) {
		if arguments[0] == "get" {
			return []byte("--future-setting=true"), nil
		}
		applied = append([]string(nil), arguments...)
		return nil, nil
	}}
	if err := plugin.Apply([]byte(`{"schemaVersion":1,"flags":["--future-setting=false"]}`)); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(applied, []string{"set", "--future-setting=false"}) {
		t.Fatalf("applied arguments = %v", applied)
	}
}

// Requirements: backup-restore/SR-002.
func TestTailscaleApplyRejectsFlagMissingFromInstalledCLI(t *testing.T) {
	plugin := &TailscalePlugin{run: func(arguments ...string) ([]byte, error) {
		if arguments[0] == "get" {
			return []byte("--ssh=true"), nil
		}
		t.Fatal("unsupported set flag was invoked")
		return nil, nil
	}}
	if err := plugin.Apply([]byte(`{"schemaVersion":1,"flags":["--future-setting=false"]}`)); err == nil {
		t.Fatal("flag absent from installed CLI was accepted")
	}
}

// Requirements: backup-restore/SR-002.
func TestTailscaleApplyRejectsMalformedFlags(t *testing.T) {
	cases := []string{
		`{"schemaVersion":1,"flags":["--authkey=secret"]}`,
		`{"schemaVersion":1,"flags":["--ssh=true","--ssh=false"]}`,
		`{"schemaVersion":1,"flags":["--hostname=--help"]}`,
		`{"schemaVersion":1,"flags":["--hostname=foo bar"]}`,
		`{"schemaVersion":1,"flags":["--ssh=true"],"extra":true}`,
	}
	for _, data := range cases {
		t.Run(data, func(t *testing.T) {
			plugin := &TailscalePlugin{run: func(arguments ...string) ([]byte, error) {
				if arguments[0] == "get" {
					return []byte("--ssh=true --hostname=foo"), nil
				}
				t.Fatal("set was invoked")
				return nil, nil
			}}
			if err := plugin.Apply([]byte(data)); err == nil {
				t.Fatal("malformed preferences were accepted")
			}
		})
	}
}

// Requirements: backup-restore/FR-003.
func TestCloudflaredValidatesArchivedCredentialAndIngress(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "cloudflared")
	credential := filepath.Join(directory, "tunnel.json")
	config := []byte("tunnel: 1234\ncredentials-file: " + credential + "\ningress:\n  - service: http_status:404\n")
	var got string
	plugin := &CloudflaredPlugin{run: func(path string) error { got = path; return nil }}
	if err := plugin.Validate(config, map[string]bool{credential: true}, "/tmp/staged-config"); err != nil {
		t.Fatal(err)
	}
	if got != "/tmp/staged-config" {
		t.Fatalf("validated path = %q", got)
	}
}

// Requirements: backup-restore/SR-002.
func TestCloudflaredRejectsCredentialOutsideArchive(t *testing.T) {
	plugin := &CloudflaredPlugin{run: func(string) error { t.Fatal("CLI was invoked"); return nil }}
	config := []byte("tunnel: 1234\ncredentials-file: /tmp/elsewhere.json\n")
	if err := plugin.Validate(config, nil, "/tmp/config"); err == nil {
		t.Fatal("external credential path was accepted")
	}
}

// Requirements: backup-restore/FR-003.
func TestPackagedCLIPluginSpecsLoad(t *testing.T) {
	for _, name := range []string{"cloudflared", "tailscale"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join("../../configs/etc/lite-nas/backup.d", name+".conf")
			if _, err := LoadSpec(path, name); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// Requirements: backup-restore/FR-003.
func TestCloudflaredSpecAllowsConfiguredLocalPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cloudflared.conf")
	content := "[backup]\nplugin=cloudflared\nconfig_path=/usr/local/etc/cloudflared/tunnel.yaml\n[directory.config]\npath=/usr/local/etc/cloudflared\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	spec, err := LoadSpec(path, "cloudflared")
	if err != nil {
		t.Fatal(err)
	}
	if spec.ConfigPath != "/usr/local/etc/cloudflared/tunnel.yaml" {
		t.Fatalf("config path = %q", spec.ConfigPath)
	}
}

// Requirements: backup-restore/FR-003.
func TestCloudflaredSpecAllowsSeparateConfigAndCredentialFiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cloudflared.conf")
	content := "[backup]\nplugin=cloudflared\nconfig_path=/etc/cloudflared/tunnel.yaml\n[file.config]\npath=/etc/cloudflared/tunnel.yaml\n[file.credential]\npath=/var/lib/cloudflared/tunnel.json\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadSpec(path, "cloudflared"); err != nil {
		t.Fatal(err)
	}
}

// Requirements: backup-restore/SR-002.
func TestCloudflaredSpecRejectsConfigOutsideSelectedDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cloudflared.conf")
	content := "[backup]\nplugin=cloudflared\nconfig_path=/tmp/other.yaml\n[directory.config]\npath=/etc/cloudflared\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadSpec(path, "cloudflared"); err == nil {
		t.Fatal("external config path was accepted")
	}
}

// Requirements: backup-restore/FR-003.
func TestTailscaleArchiveContainsOnlyJSONAndSpec(t *testing.T) {
	plugin := &TailscalePlugin{run: func(...string) ([]byte, error) { return []byte("--ssh=false"), nil }}
	var archive bytes.Buffer
	spec := Spec{Application: "tailscale", Plugin: "tailscale", Selections: []Selection{{Path: tailscaleArchivePath}}}
	if err := writeArchive(&archive, spec, nil, nil, plugin, nil); err != nil {
		t.Fatal(err)
	}
	paths, err := ArchivePaths(bytes.NewReader(archive.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(paths, []string{"tailscale.json"}) {
		t.Fatalf("archive paths = %v", paths)
	}
}

// Requirements: backup-restore/SR-002.
func TestCloudflaredRejectsMissingCredential(t *testing.T) {
	plugin := &CloudflaredPlugin{run: func(string) error { t.Fatal("CLI was invoked"); return nil }}
	config := []byte("tunnel: 1234\ncredentials-file: /etc/cloudflared/tunnel.json\n")
	if err := plugin.Validate(config, nil, "/tmp/config"); err == nil {
		t.Fatal("missing credential was accepted")
	}
}

// Requirements: backup-restore/SR-002.
func TestCloudflaredRejectsMalformedYAML(t *testing.T) {
	plugin := &CloudflaredPlugin{run: func(string) error { t.Fatal("CLI was invoked"); return nil }}
	if err := plugin.Validate([]byte("tunnel: ["), nil, "/tmp/config"); err == nil {
		t.Fatal("malformed YAML was accepted")
	}
}

// Requirements: backup-restore/FR-003.
func TestTailscaleExportRejectsUnrepresentablePreference(t *testing.T) {
	plugin := &TailscalePlugin{run: func(...string) ([]byte, error) { return []byte("--nickname=two words"), nil }}
	if _, err := plugin.Export(); err == nil {
		t.Fatalf("unrepresentable preference error = %v", err)
	}
}

// Requirements: backup-restore/FR-003.
func TestCLIPluginSpecsRejectUnexpectedSelection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tailscale.conf")
	if err := os.WriteFile(path, []byte("[backup]\nplugin=tailscale\n[file.state]\npath=/var/lib/tailscale/tailscaled.state\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadSpec(path, "tailscale"); err == nil {
		t.Fatal("Tailscale identity state selection was accepted")
	}
}

// Requirements: backup-restore/FR-003.
func TestTailscaleStoreRestoreReplaysPreferences(t *testing.T) {
	var applied []string
	plugin := &TailscalePlugin{run: func(arguments ...string) ([]byte, error) {
		if arguments[0] == "get" {
			return []byte("--ssh=false --accept-dns=true"), nil
		}
		applied = append([]string(nil), arguments...)
		return nil, nil
	}}
	store := Store{Root: t.TempDir(), Tailscale: plugin}
	spec := Spec{Application: "tailscale", Plugin: "tailscale", Selections: []Selection{{Path: tailscaleArchivePath}}}
	version, err := store.Backup(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Restore(spec, version, ReplaceMode); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(applied, []string{"set", "--ssh=false", "--accept-dns=true"}) {
		t.Fatalf("applied arguments = %v", applied)
	}
}

// Requirements: backup-restore/SR-002.
func TestCloudflaredRestoreValidationUsesStagedConfig(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "cloudflared")
	credential := filepath.Join(directory, "tunnel.json")
	configPath := filepath.Join(directory, "config.yml")
	staged := filepath.Join(t.TempDir(), "staged-config")
	if err := os.WriteFile(staged, []byte("tunnel: 1234\ncredentials-file: "+credential+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	entries := []stagedEntry{
		{path: configPath, stage: staged, header: &tar.Header{Typeflag: tar.TypeReg}},
		{path: credential, header: &tar.Header{Typeflag: tar.TypeReg}},
	}
	var got string
	plugin := &CloudflaredPlugin{run: func(path string) error { got = path; return nil }}
	if err := validateStagedCloudflared(Spec{ConfigPath: configPath, Selections: []Selection{{Path: directory, Directory: true}}}, entries, plugin); err != nil {
		t.Fatal(err)
	}
	if got != staged {
		t.Fatalf("validated path = %q", got)
	}
}

// TestCloudflaredLiveValidationChecksSelectedCredential covers backup-restore/SR-002.
func TestCloudflaredLiveValidationChecksSelectedCredential(t *testing.T) {
	root := t.TempDir()
	credential := filepath.Join(root, "tunnel.json")
	config := filepath.Join(root, "config.yml")
	if err := os.WriteFile(credential, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, []byte("tunnel: 1234\ncredentials-file: "+credential+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var validated string
	plugin := &CloudflaredPlugin{run: func(path string) error { validated = path; return nil }}
	spec := Spec{ConfigPath: config, Selections: []Selection{{Path: root, Directory: true}}}
	if err := validateLiveCloudflared(spec, plugin); err != nil {
		t.Fatal(err)
	}
	if validated != config {
		t.Fatalf("validated path = %q", validated)
	}
}

// TestCloudflaredBackupValidatesLiveConfig covers backup-restore/FR-003.
func TestCloudflaredBackupValidatesLiveConfig(t *testing.T) {
	root := t.TempDir()
	credential := filepath.Join(root, "tunnel.json")
	config := filepath.Join(root, "config.yml")
	if err := os.WriteFile(credential, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, []byte("tunnel: 1234\ncredentials-file: "+credential+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var validated bool
	plugin := &CloudflaredPlugin{run: func(path string) error { validated = path == config; return nil }}
	store := Store{Root: filepath.Join(t.TempDir(), "backups"), Cloudflared: plugin}
	spec := Spec{Application: "cloudflared", Plugin: "cloudflared", ConfigPath: config, Selections: []Selection{{Path: root, Directory: true}}}
	_, err := store.Backup(spec)
	if err != nil || !validated {
		t.Fatalf("backup = %v, validated = %t", err, validated)
	}
}

// TestCloudflaredValidationRejectsRelativeConfig covers backup-restore/SR-003.
func TestCloudflaredValidationRejectsRelativeConfig(t *testing.T) {
	if err := runCloudflaredValidation("relative.yml"); err == nil {
		t.Fatal("relative config path accepted")
	}
}

// TestCloudflaredLiveValidationRejectsUnselectedCredential covers backup-restore/SR-002.
func TestCloudflaredLiveValidationRejectsUnselectedCredential(t *testing.T) {
	root := t.TempDir()
	config := filepath.Join(root, "config.yml")
	if err := os.WriteFile(config, []byte("tunnel: 1234\ncredentials-file: /tmp/unselected.json\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	plugin := &CloudflaredPlugin{run: func(string) error { t.Fatal("CLI reached"); return nil }}
	spec := Spec{ConfigPath: config, Selections: []Selection{{Path: config}}}
	if err := validateLiveCloudflared(spec, plugin); err == nil {
		t.Fatal("unselected credential accepted")
	}
}

// TestCloudflaredLiveValidationRejectsMissingCredential covers backup-restore/SR-002.
func TestCloudflaredLiveValidationRejectsMissingCredential(t *testing.T) {
	root := t.TempDir()
	config := filepath.Join(root, "config.yml")
	credential := filepath.Join(root, "missing.json")
	if err := os.WriteFile(config, []byte("tunnel: 1234\ncredentials-file: "+credential+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	plugin := &CloudflaredPlugin{run: func(string) error { t.Fatal("CLI reached"); return nil }}
	spec := Spec{ConfigPath: config, Selections: []Selection{{Path: root, Directory: true}}}
	if err := validateLiveCloudflared(spec, plugin); err == nil {
		t.Fatal("missing credential accepted")
	}
}

// TestTailscaleExportRejectsUnavailablePlugin covers backup-restore/SR-003.
func TestTailscaleExportRejectsUnavailablePlugin(t *testing.T) {
	var plugin *TailscalePlugin
	if _, err := plugin.Export(); err == nil {
		t.Fatal("unavailable plugin exported preferences")
	}
}

// TestTailscaleApplyRejectsMalformedInstalledPreferences covers backup-restore/SR-002.
func TestTailscaleApplyRejectsMalformedInstalledPreferences(t *testing.T) {
	plugin := &TailscalePlugin{run: func(arguments ...string) ([]byte, error) {
		if arguments[0] == "get" {
			return []byte("--ssh=invalid"), nil
		}
		t.Fatal("set reached after malformed installed preferences")
		return nil, nil
	}}
	if err := plugin.Apply([]byte(`{"schemaVersion":1,"flags":["--ssh=false"]}`)); err == nil {
		t.Fatal("malformed installed preferences accepted")
	}
}

// TestTailscaleApplySurfacesInstalledCLIError covers backup-restore/SR-002.
func TestTailscaleApplySurfacesInstalledCLIError(t *testing.T) {
	plugin := &TailscalePlugin{run: func(arguments ...string) ([]byte, error) {
		if arguments[0] != "get" {
			t.Fatal("set reached after CLI error")
		}
		return nil, errors.New("CLI unavailable")
	}}
	if err := plugin.Apply([]byte(`{"schemaVersion":1,"flags":["--ssh=false"]}`)); err == nil {
		t.Fatal("CLI error ignored")
	}
}

// TestTailscaleRunnerRejectsMissingCommand covers backup-restore/SR-003.
func TestTailscaleRunnerRejectsMissingCommand(t *testing.T) {
	if _, err := runTailscaleCommand(); err == nil {
		t.Fatal("missing command accepted")
	}
}

// TestTailscaleFlagValuesRejectMalformedOperands covers backup-restore/SR-002.
func TestTailscaleFlagValuesRejectMalformedOperands(t *testing.T) {
	cases := []struct{ name, flag, value string }{
		{"route", "advertise-routes", "not-a-prefix"},
		{"netfilter", "netfilter-mode", "invalid"},
		{"port", "relay-server-port", "65536"},
		{"endpoint", "relay-server-static-endpoints", "invalid"},
		{"hostname", "hostname", "space here"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if err := validateTailscaleFlagValue(testCase.flag, testCase.value, true); err == nil {
				t.Fatal("malformed Tailscale flag value accepted")
			}
		})
	}
}

// TestTailscaleFlagValuesAcceptValidNetworkOperands covers backup-restore/FR-003.
func TestTailscaleFlagValuesAcceptValidNetworkOperands(t *testing.T) {
	cases := []struct{ name, flag, value string }{
		{"route", "advertise-routes", "192.0.2.0/24"},
		{"netfilter", "netfilter-mode", "nodivert"},
		{"port", "relay-server-port", "41641"},
		{"endpoint", "relay-server-static-endpoints", "192.0.2.1:41641"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if err := validateTailscaleFlagValue(testCase.flag, testCase.value, true); err != nil {
				t.Fatalf("valid Tailscale flag rejected: %v", err)
			}
		})
	}
}

// TestTailscaleDiffOmitsUnchangedPreferences covers backup-restore/FR-010.
func TestTailscaleDiffOmitsUnchangedPreferences(t *testing.T) {
	root := t.TempDir()
	plugin := &TailscalePlugin{run: func(...string) ([]byte, error) { return []byte("--ssh=false"), nil }}
	store := Store{Root: filepath.Join(root, "backups"), Tailscale: plugin}
	spec := Spec{Application: "tailscale", Plugin: "tailscale", Selections: []Selection{{Path: tailscaleArchivePath}}}
	version, err := store.Backup(spec)
	if err != nil {
		t.Fatal(err)
	}
	differences, err := store.Diff(spec, version)
	if err != nil {
		t.Fatal(err)
	}
	if len(differences) != 0 {
		t.Fatalf("differences = %#v", differences)
	}
}
