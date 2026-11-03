package backuprestore

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"

	contract "lite-nas/shared/contracts/backuprestore"
	"lite-nas/shared/testutil/authtokentest"
	"lite-nas/shared/testutil/messagingtest"
)

// mustUFWPlugin builds a fake CLI-backed plugin with isolated configuration files.
func mustUFWPlugin(t *testing.T, report string, run func(...string) ([]byte, error)) *UFWPlugin {
	t.Helper()
	root := t.TempDir()
	defaults := filepath.Join(root, "defaults")
	configuration := filepath.Join(root, "ufw.conf")
	if err := os.WriteFile(defaults, []byte("DEFAULT_INPUT_POLICY=\"DROP\"\nDEFAULT_OUTPUT_POLICY=\"ACCEPT\"\nDEFAULT_FORWARD_POLICY=\"DROP\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configuration, []byte("ENABLED=yes\nLOGLEVEL=low\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if run == nil {
		run = func(args ...string) ([]byte, error) {
			if len(args) == 2 && args[0] == "show" && args[1] == "added" {
				return []byte(report), nil
			}
			return nil, nil
		}
	}
	return &UFWPlugin{run: run, defaults: defaults, configuration: configuration}
}

// mustUFWStore creates one configured UFW target and its first snapshot.
func mustUFWStore(t *testing.T, plugin *UFWPlugin) (Spec, Store, string) {
	t.Helper()
	root := t.TempDir()
	source := filepath.Join(root, "ufw-rules.conf")
	if err := os.WriteFile(source, []byte("[backup]\nplugin=ufw\nrestart_units=\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	spec, err := LoadSpec(source, "ufw-rules")
	if err != nil {
		t.Fatal(err)
	}
	store := Store{Root: filepath.Join(root, "backups"), UFW: plugin}
	version, err := store.Backup(spec)
	if err != nil {
		t.Fatal(err)
	}
	return spec, store, version
}

// Requirements: backup-restore/FR-003.
func TestParseUFWAddedPreservesRouteAndQuotedApplication(t *testing.T) {
	report := "Added user rules (see 'ufw status' for running firewall):\nufw route allow in on eth0 out on eth1 from 192.0.2.0/24 to any port 443 proto tcp\nufw allow 'Web Server' comment 'public site'\n"
	rules, err := parseUFWAdded([]byte(report))
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 2 || rules[0].Kind != "route" || rules[1].Arguments[0] != "Web Server" || rules[1].Arguments[2] != "public site" {
		t.Fatalf("parsed rules = %+v", rules)
	}
}

// Requirements: backup-restore/FR-003.
func TestParseUFWAddedRejectsAmbiguousCommentQuote(t *testing.T) {
	_, err := parseUFWAdded([]byte("Added user rules (see 'ufw status' for running firewall):\nufw allow 443/tcp comment 'don't can't'\n"))
	if err == nil {
		t.Fatal("ambiguous comment accepted")
	}
}

// Requirements: backup-restore/SR-003.
func TestUFWRuleValidationRejectsOptionShapedOperand(t *testing.T) {
	err := validateUFWRule(UFWRule{Kind: "rule", Action: "allow", Arguments: []string{"--force"}})
	if err == nil {
		t.Fatal("option-shaped rule operand accepted")
	}
}

// Requirements: backup-restore/SR-003.
func TestDecodeUFWStateRejectsMissingEnabledField(t *testing.T) {
	_, err := decodeUFWState([]byte(`{"schemaVersion":1,"logging":"low","defaults":{"incoming":"deny","outgoing":"allow","routed":"deny"},"rules":[]}`))
	if err == nil {
		t.Fatal("missing enabled field accepted")
	}
}

// Requirements: backup-restore/FR-003.
func TestUFWExportProducesVersionedJSON(t *testing.T) {
	plugin := mustUFWPlugin(t, "Added user rules (see 'ufw status' for running firewall):\nufw allow 443/tcp\n", nil)
	data, err := plugin.Export()
	if err != nil {
		t.Fatal(err)
	}
	var state UFWState
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatal(err)
	}
	if state.SchemaVersion != 1 || !state.Enabled || state.Defaults.Incoming != "deny" || len(state.Rules) != 1 {
		t.Fatalf("exported UFW state = %+v", state)
	}
}

// Requirements: backup-restore/SR-003.
func TestUFWApplyRejectsInvalidJSONBeforeCommands(t *testing.T) {
	commands := 0
	plugin := mustUFWPlugin(t, "", func(...string) ([]byte, error) {
		commands++
		return nil, nil
	})
	err := plugin.Apply([]byte(`{"schemaVersion":1,"enabled":true,"logging":"low","defaults":{"incoming":"deny","outgoing":"allow","routed":"deny"},"rules":[{"kind":"rule","action":"reset","arguments":["all"]}]}`))
	if err == nil || commands != 0 {
		t.Fatalf("invalid UFW JSON reached CLI: error=%v commands=%d", err, commands)
	}
}

// Requirements: backup-restore/FR-003.
func TestUFWArchiveContainsGeneratedJSON(t *testing.T) {
	plugin := mustUFWPlugin(t, "Added user rules (see 'ufw status' for running firewall):\n(None)\n", nil)
	spec, store, version := mustUFWStore(t, plugin)
	archive, err := store.OpenVersion(spec.Application, version)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	paths, err := ArchivePaths(archive)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(paths, ","), "ufw.json") {
		t.Fatalf("archive paths = %v", paths)
	}
}

// TestUFWDiffFindsChangedRules covers backup-restore/FR-003.
func TestUFWDiffFindsChangedRules(t *testing.T) {
	const header = "Added user rules (see 'ufw status' for running firewall):\n"
	plugin := mustUFWPlugin(t, header+"(None)\n", nil)
	spec, store, version := mustUFWStore(t, plugin)
	plugin.run = func(args ...string) ([]byte, error) { return []byte(header + "ufw allow 443/tcp\n"), nil }
	differences, err := store.Diff(spec, version)
	if err != nil {
		t.Fatal(err)
	}
	if len(differences) != 1 || differences[0].Path != ufwArchivePath || differences[0].Status != "changed" {
		t.Fatalf("differences = %#v", differences)
	}
}

// TestUFWDiffOmitsUnchangedRules covers backup-restore/FR-010.
func TestUFWDiffOmitsUnchangedRules(t *testing.T) {
	plugin := mustUFWPlugin(t, "Added user rules (see 'ufw status' for running firewall):\n(None)\n", nil)
	spec, store, version := mustUFWStore(t, plugin)
	differences, err := store.Diff(spec, version)
	if err != nil || len(differences) != 0 {
		t.Fatalf("differences = %#v, error = %v", differences, err)
	}
}

// TestUFWRunnerRejectsMissingCommand covers backup-restore/SR-003.
func TestUFWRunnerRejectsMissingCommand(t *testing.T) {
	if _, err := runUFWCommand(); err == nil {
		t.Fatal("missing command accepted")
	}
}

// TestUFWExportRejectsMissingConfiguration covers backup-restore/SR-003.
func TestUFWExportRejectsMissingConfiguration(t *testing.T) {
	plugin := &UFWPlugin{run: func(...string) ([]byte, error) { t.Fatal("CLI reached"); return nil, nil }, defaults: filepath.Join(t.TempDir(), "missing-defaults")}
	if _, err := plugin.Export(); err == nil {
		t.Fatal("missing UFW defaults accepted")
	}
}

// TestUFWExportRejectsMissingRuntimeConfiguration covers backup-restore/SR-003.
func TestUFWExportRejectsMissingRuntimeConfiguration(t *testing.T) {
	plugin := mustUFWPlugin(t, "", nil)
	plugin.configuration = filepath.Join(t.TempDir(), "missing-ufw.conf")
	if _, err := plugin.Export(); err == nil {
		t.Fatal("missing runtime configuration accepted")
	}
}

// TestUFWExportRejectsMalformedDefaults covers backup-restore/SR-003.
func TestUFWExportRejectsMalformedDefaults(t *testing.T) {
	plugin := mustUFWPlugin(t, "", nil)
	if err := os.WriteFile(plugin.defaults, []byte("INVALID=\"value\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := plugin.Export(); err == nil {
		t.Fatal("malformed UFW defaults accepted")
	}
}

// TestLimitedUFWOutputBoundsCommandText covers backup-restore/SR-003.
func TestLimitedUFWOutputBoundsCommandText(t *testing.T) {
	var output limitedUFWOutput
	_, _ = output.Write(make([]byte, 1<<20+1))
	if !output.exceeded || output.Len() != 0 {
		t.Fatal("oversized command output was retained")
	}
}

// Requirements: backup-restore/FR-003.
func TestUFWApplyReplaysRulesThroughCLI(t *testing.T) {
	var commands []string
	plugin := mustUFWPlugin(t, "", func(args ...string) ([]byte, error) {
		commands = append(commands, strings.Join(args, " "))
		if len(args) == 2 && args[0] == "show" && args[1] == "added" {
			return []byte("Added user rules (see 'ufw status' for running firewall):\nufw deny 80/tcp\n"), nil
		}
		return nil, nil
	})
	state := UFWState{SchemaVersion: 1, Enabled: true, Logging: "low", Defaults: UFWDefaults{Incoming: "deny", Outgoing: "allow", Routed: "deny"}, Rules: []UFWRule{{Kind: "rule", Action: "allow", Arguments: []string{"443/tcp"}}}}
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := plugin.Apply(data); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(commands, "|"), "--force delete deny 80/tcp|default deny incoming|default allow outgoing|default deny routed|allow 443/tcp|logging low|--force enable") {
		t.Fatalf("UFW commands = %v", commands)
	}
}

// Requirements: backup-restore/FR-003.
func TestUFWApplySkipsEquivalentState(t *testing.T) {
	mutations := 0
	plugin := mustUFWPlugin(t, "", func(args ...string) ([]byte, error) {
		if len(args) == 2 && args[0] == "show" && args[1] == "added" {
			return []byte("Added user rules (see 'ufw status' for running firewall):\n(None)\n"), nil
		}
		if len(args) > 0 && args[0] != "--dry-run" {
			mutations++
		}
		return nil, nil
	})
	state := UFWState{SchemaVersion: 1, Enabled: true, Logging: "low", Defaults: UFWDefaults{Incoming: "deny", Outgoing: "allow", Routed: "deny"}, Rules: []UFWRule{}}
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := plugin.Apply(data); err != nil {
		t.Fatal(err)
	}
	if mutations != 0 {
		t.Fatalf("equivalent state caused %d UFW mutations", mutations)
	}
}

// Requirements: backup-restore/FR-011.
func TestUFWDirectMergePreservesLiveOnlyRule(t *testing.T) {
	var commands []string
	plugin := mustUFWPlugin(t, "", func(args ...string) ([]byte, error) {
		commands = append(commands, strings.Join(args, " "))
		if len(args) == 2 && args[0] == "show" && args[1] == "added" {
			return []byte("Added user rules (see 'ufw status' for running firewall):\nufw deny 80/tcp\n"), nil
		}
		return nil, nil
	})
	state := UFWState{SchemaVersion: 1, Enabled: true, Logging: "low", Defaults: UFWDefaults{Incoming: "deny", Outgoing: "allow", Routed: "deny"}, Rules: []UFWRule{{Kind: "rule", Action: "allow", Arguments: []string{"443/tcp"}}}}
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := plugin.Merge(data); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(commands, "|"), "allow 443/tcp|deny 80/tcp|logging low") {
		t.Fatalf("merged UFW commands = %v", commands)
	}
}

// Requirements: backup-restore/FR-003.
func TestUFWApplyAttemptsRollbackAfterRuleFailure(t *testing.T) {
	rules := "ufw deny 80/tcp\n"
	plugin := mustUFWPlugin(t, "", func(args ...string) ([]byte, error) {
		if len(args) == 2 && args[0] == "show" && args[1] == "added" {
			return []byte("Added user rules (see 'ufw status' for running firewall):\n" + rules), nil
		}
		if len(args) >= 2 && args[0] == "--force" && args[1] == "delete" {
			rules = ""
		}
		if len(args) == 2 && args[0] == "allow" && args[1] == "443/tcp" {
			return nil, errors.New("simulated add failure")
		}
		if len(args) == 2 && args[0] == "deny" && args[1] == "80/tcp" {
			rules = "ufw deny 80/tcp\n"
		}
		return nil, nil
	})
	state := UFWState{SchemaVersion: 1, Enabled: true, Logging: "low", Defaults: UFWDefaults{Incoming: "deny", Outgoing: "allow", Routed: "deny"}, Rules: []UFWRule{{Kind: "rule", Action: "allow", Arguments: []string{"443/tcp"}}}}
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	err = plugin.Apply(data)
	if err == nil || rules != "ufw deny 80/tcp\n" {
		t.Fatalf("rollback result: error=%v rules=%q", err, rules)
	}
}

// TestUFWApplyReportsRollbackInspectionFailure covers backup-restore/SR-003.
func TestUFWApplyReportsRollbackInspectionFailure(t *testing.T) {
	reads := 0
	plugin := mustUFWPlugin(t, "", func(args ...string) ([]byte, error) {
		if len(args) == 2 && args[0] == "show" && args[1] == "added" {
			reads++
			if reads > 1 {
				return nil, errors.New("inspection unavailable")
			}
			return []byte("Added user rules (see 'ufw status' for running firewall):\nufw deny 80/tcp\n"), nil
		}
		if len(args) == 2 && args[0] == "allow" {
			return nil, errors.New("rule rejected")
		}
		return nil, nil
	})
	state := UFWState{SchemaVersion: 1, Enabled: true, Logging: "low", Defaults: UFWDefaults{Incoming: "deny", Outgoing: "allow", Routed: "deny"}, Rules: []UFWRule{{Kind: "rule", Action: "allow", Arguments: []string{"443/tcp"}}}}
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := plugin.Apply(data); err == nil || !strings.Contains(err.Error(), "rollback inspection failed") {
		t.Fatalf("rollback error = %v", err)
	}
}

// TestUFWApplyReportsRollbackFailure covers backup-restore/SR-003.
func TestUFWApplyReportsRollbackFailure(t *testing.T) {
	plugin := mustUFWPlugin(t, "", func(args ...string) ([]byte, error) {
		if len(args) == 2 && args[0] == "show" && args[1] == "added" {
			return []byte("Added user rules (see 'ufw status' for running firewall):\nufw deny 80/tcp\n"), nil
		}
		if len(args) == 2 && (args[0] == "allow" || args[0] == "deny") {
			return nil, errors.New("rule rejected")
		}
		return nil, nil
	})
	state := UFWState{SchemaVersion: 1, Enabled: true, Logging: "low", Defaults: UFWDefaults{Incoming: "deny", Outgoing: "allow", Routed: "deny"}, Rules: []UFWRule{{Kind: "rule", Action: "allow", Arguments: []string{"443/tcp"}}}}
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	if err := plugin.Apply(data); err == nil || !strings.Contains(err.Error(), "rollback failed") {
		t.Fatalf("rollback error = %v", err)
	}
}

// Requirements: backup-restore/FR-005.
func TestUFWBackupStartsUIMergeSessionWithJSON(t *testing.T) {
	plugin := mustUFWPlugin(t, "Added user rules (see 'ufw status' for running firewall):\n(None)\n", nil)
	spec, store, version := mustUFWStore(t, plugin)
	verifier, token := authtokentest.MustIssueAccessToken(t, []string{"admin"})
	runtime := &Runtime{specs: map[string]Spec{"ufw-rules": spec}, store: store, verifier: verifier, validate: validator.New(validator.WithRequiredStructEnabled()), client: &messagingtest.RecordingClient{AuthUsername: "tester"}}
	result, err := runtime.mergeStart(context.Background(), messagingtest.MustEnvelope(t, contract.MergeStartRequest{AccessToken: token, Application: "ufw-rules", Version: version}))
	if err != nil {
		t.Fatal(err)
	}
	if files := result.(contract.MergeStartResponse).Files; len(files) != 1 || files[0].Path != ufwArchivePath {
		t.Fatalf("merge files = %v", files)
	}
}

// Requirements: backup-restore/FR-005.
func TestUFWUIMergeApplyReplaysResolvedJSON(t *testing.T) {
	var commands []string
	plugin := mustUFWPlugin(t, "", func(args ...string) ([]byte, error) {
		commands = append(commands, strings.Join(args, " "))
		if len(args) == 2 && args[0] == "show" && args[1] == "added" {
			return []byte("Added user rules (see 'ufw status' for running firewall):\n(None)\n"), nil
		}
		return nil, nil
	})
	spec, store, version := mustUFWStore(t, plugin)
	verifier, token := authtokentest.MustIssueAccessToken(t, []string{"admin"})
	runtime := &Runtime{specs: map[string]Spec{"ufw-rules": spec}, store: store, verifier: verifier, validate: validator.New(validator.WithRequiredStructEnabled()), client: &messagingtest.RecordingClient{AuthUsername: "tester"}}
	started, err := runtime.mergeStart(context.Background(), messagingtest.MustEnvelope(t, contract.MergeStartRequest{AccessToken: token, Application: "ufw-rules", Version: version}))
	if err != nil {
		t.Fatal(err)
	}
	state := UFWState{SchemaVersion: 1, Enabled: true, Logging: "low", Defaults: UFWDefaults{Incoming: "deny", Outgoing: "allow", Routed: "deny"}, Rules: []UFWRule{{Kind: "rule", Action: "allow", Arguments: []string{"443/tcp"}}}}
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	id := started.(contract.MergeStartResponse).SessionID
	mustStageMerge(t, runtime, token, id, ufwArchivePath, string(data))
	if _, err := runtime.mergeApply(context.Background(), messagingtest.MustEnvelope(t, contract.MergeActionRequest{AccessToken: token, SessionID: id})); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(commands, "|"), "|allow 443/tcp|logging low") {
		t.Fatalf("UFW merge commands = %v", commands)
	}
}

// Requirements: backup-restore/FR-004.
func TestUFWRestoreReplaysArchivedJSON(t *testing.T) {
	rules := ""
	plugin := mustUFWPlugin(t, "", func(args ...string) ([]byte, error) {
		if len(args) == 2 && args[0] == "show" && args[1] == "added" {
			if rules == "" {
				return []byte("Added user rules (see 'ufw status' for running firewall):\n(None)\n"), nil
			}
			return []byte("Added user rules (see 'ufw status' for running firewall):\n" + rules), nil
		}
		if len(args) >= 2 && args[0] == "--force" && args[1] == "delete" {
			rules = ""
		}
		return nil, nil
	})
	spec, store, version := mustUFWStore(t, plugin)
	rules = "ufw deny 80/tcp\n"
	if err := store.Restore(spec, version, ReplaceMode); err != nil {
		t.Fatal(err)
	}
	if rules != "" {
		t.Fatalf("restored UFW rules = %q", rules)
	}
}
