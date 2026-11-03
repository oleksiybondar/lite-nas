package backuprestore

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"gopkg.in/ini.v1"

	"lite-nas/shared/authtoken"
	authcontract "lite-nas/shared/contracts/auth"
	contract "lite-nas/shared/contracts/backuprestore"
	managercontract "lite-nas/shared/contracts/servicesmanager"
	"lite-nas/shared/messaging"
	"lite-nas/shared/testutil/authtokentest"
	"lite-nas/shared/testutil/messagingtest"
)

// failingRestartClient records the requested unit and simulates a failed restart.
type failingRestartClient struct {
	messagingtest.RecordingClient
	unit string
}

// failingRegistrationServer fails one selected subject while preserving other server behavior.
type failingRegistrationServer struct {
	messagingtest.RecordingServer
	subject string
}

// RegisterRPC reports a controlled registration failure for the selected subject.
func (server *failingRegistrationServer) RegisterRPC(subject string, handler messaging.RPCHandler) error {
	if subject == server.subject {
		return errors.New("registration failed")
	}
	return server.RecordingServer.RegisterRPC(subject, handler)
}

// Request records a restart request and returns an observable failure.
func (client *failingRestartClient) Request(_ context.Context, subject string, request any, response any) error {
	if subject == authcontract.ValidateAccessTokenRPCSubject {
		*response.(*authcontract.ValidateAccessTokenResponse) = authcontract.ValidateAccessTokenResponse{Valid: true, Status: authcontract.StatusAuthenticated, Username: "tester"}
		return nil
	}
	if subject != "services-manager.rpc.restart" {
		return errors.New("unexpected subject")
	}
	client.unit = request.(managercontract.ActionRequest).Unit
	return errors.New("restart failed")
}

// Requirements: backup-restore/FR-004.
func TestRestoreReportsRestartFailureAsWarning(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "settings.conf")
	if err := os.WriteFile(path, []byte("before"), 0o600); err != nil {
		t.Fatal(err)
	}
	spec := Spec{Application: "example", Selections: []Selection{{Path: path}}, RestartUnits: []string{"example.service"}}
	store := Store{Root: filepath.Join(root, "backups")}
	version, err := store.Backup(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("after"), 0o600); err != nil {
		t.Fatal(err)
	}
	verifier, token := authtokentest.MustIssueAccessToken(t, []string{"admin"})
	client := &failingRestartClient{}
	runtime := &Runtime{specs: map[string]Spec{"example": spec}, store: store, verifier: verifier, validate: validator.New(validator.WithRequiredStructEnabled()), client: client}
	envelope := messagingtest.MustEnvelope(t, contract.RestoreRequest{AccessToken: token, Application: "example", Version: version, Mode: "replace"})
	result, err := runtime.restore(context.Background(), envelope)
	if err != nil {
		t.Fatal(err)
	}
	response := result.(contract.RestoreResponse)
	if !response.Restored || len(response.Warnings) != 1 || client.unit != "example.service" {
		t.Fatalf("response = %+v, unit = %s", response, client.unit)
	}
}

// Requirements: backup-restore/FR-012.
// TestRestoreByNameSelectsNewestMatchingVersion checks exact names and version order.
func TestRestoreByNameSelectsNewestMatchingVersion(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "settings.conf")
	if err := os.WriteFile(path, []byte("first"), 0o600); err != nil {
		t.Fatal(err)
	}
	spec := Spec{Application: "example", Selections: []Selection{{Path: path}}}
	store := Store{Root: filepath.Join(root, "backups")}
	if _, err := store.BackupWithMetadata(spec, "named", "", "admin"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("second"), 0o600); err != nil {
		t.Fatal(err)
	}
	want, err := store.BackupWithMetadata(spec, "named", "", "admin")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("live"), 0o600); err != nil {
		t.Fatal(err)
	}
	verifier, token := authtokentest.MustIssueAccessToken(t, []string{"admin"})
	runtime := &Runtime{specs: map[string]Spec{"example": spec}, store: store, verifier: verifier, validate: validator.New(validator.WithRequiredStructEnabled()), client: &messagingtest.RecordingClient{AuthUsername: "tester"}}
	result, err := runtime.restoreByName(context.Background(), messagingtest.MustEnvelope(t, contract.RestoreByNameRequest{AccessToken: token, Application: "example", Name: "named", Mode: "replace"}))
	if err != nil {
		t.Fatal(err)
	}
	if got := result.(contract.RestoreByNameResponse).Version; got != want {
		t.Fatalf("restored version = %q, want %q", got, want)
	}
}

// Requirements: backup-restore/FR-012.
// TestRestoreAllByNameSkipsUnmatchedApplications checks all-target outcome reporting.
func TestRestoreAllByNameSkipsUnmatchedApplications(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "settings.conf")
	if err := os.WriteFile(path, []byte("stored"), 0o600); err != nil {
		t.Fatal(err)
	}
	spec := Spec{Application: "matched", Selections: []Selection{{Path: path}}}
	store := Store{Root: filepath.Join(root, "backups")}
	if _, err := store.BackupWithMetadata(spec, "release", "", "admin"); err != nil {
		t.Fatal(err)
	}
	verifier, token := authtokentest.MustIssueAccessToken(t, []string{"admin"})
	runtime := &Runtime{specs: map[string]Spec{"matched": spec, "unmatched": {Application: "unmatched", Selections: []Selection{{Path: path}}}}, store: store, verifier: verifier, validate: validator.New(validator.WithRequiredStructEnabled()), client: &messagingtest.RecordingClient{AuthUsername: "tester"}}
	result, err := runtime.restoreAllByName(context.Background(), messagingtest.MustEnvelope(t, contract.RestoreAllByNameRequest{AccessToken: token, Name: "release", Mode: "replace"}))
	if err != nil {
		t.Fatal(err)
	}
	results := result.(contract.RestoreAllByNameResponse).Results
	if len(results) != 2 || results[0].Application != "matched" || results[0].Status != "completed" || results[1].Application != "unmatched" || results[1].Status != "skipped" {
		t.Fatalf("restore results = %+v", results)
	}
}

// Requirements: backup-restore/FR-012.
// TestRestoreAllByNameContinuesAfterFailure checks that later targets are still visited.
func TestRestoreAllByNameContinuesAfterFailure(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "settings.conf")
	if err := os.WriteFile(path, []byte("stored"), 0o600); err != nil {
		t.Fatal(err)
	}
	store := Store{Root: filepath.Join(root, "backups")}
	bad := Spec{Application: "bad", Selections: []Selection{{Path: path}}}
	good := Spec{Application: "good", Selections: []Selection{{Path: path}}}
	for _, spec := range []Spec{bad, good} {
		if _, err := store.BackupWithMetadata(spec, "release", "", "admin"); err != nil {
			t.Fatal(err)
		}
	}
	bad.Selections[0].Path = filepath.Join(root, "different.conf")
	verifier, token := authtokentest.MustIssueAccessToken(t, []string{"admin"})
	runtime := &Runtime{specs: map[string]Spec{"bad": bad, "good": good}, store: store, verifier: verifier, validate: validator.New(validator.WithRequiredStructEnabled()), client: &messagingtest.RecordingClient{AuthUsername: "tester"}}
	result, err := runtime.restoreAllByName(context.Background(), messagingtest.MustEnvelope(t, contract.RestoreAllByNameRequest{AccessToken: token, Name: "release", Mode: "replace"}))
	if err != nil {
		t.Fatal(err)
	}
	results := result.(contract.RestoreAllByNameResponse).Results
	if len(results) != 2 || results[0].Status != "failed" || results[1].Status != "completed" {
		t.Fatalf("restore results = %+v", results)
	}
}

// Requirements: backup-restore/FR-009.
func TestApplicationsRPCListsConfiguredTarget(t *testing.T) {
	runtime, token, _, _ := mergeFixture(t)
	result, err := runtime.applications(context.Background(), messagingtest.MustEnvelope(t, contract.ApplicationsRequest{AccessToken: token}))
	if err != nil {
		t.Fatal(err)
	}
	if got := result.(contract.ApplicationsResponse).Applications; len(got) != 1 || got[0] != "example" {
		t.Fatalf("applications = %v", got)
	}
}

// Requirements: backup-restore/FR-009.
func TestApplicationsRPCListsTargetWithoutBackupHistory(t *testing.T) {
	verifier, token := authtokentest.MustIssueAccessToken(t, []string{"admin"})
	runtime := &Runtime{specs: map[string]Spec{"example": {Application: "example"}}, store: Store{Root: t.TempDir()}, verifier: verifier, validate: validator.New(validator.WithRequiredStructEnabled()), client: &messagingtest.RecordingClient{AuthUsername: "tester"}}
	result, err := runtime.applications(context.Background(), messagingtest.MustEnvelope(t, contract.ApplicationsRequest{AccessToken: token}))
	if err != nil {
		t.Fatal(err)
	}
	if got := result.(contract.ApplicationsResponse).Applications; len(got) != 1 || got[0] != "example" {
		t.Fatalf("applications = %v", got)
	}
}

// Requirements: backup-restore/FR-001.
func TestVersionsRPCListsStoredSnapshot(t *testing.T) {
	runtime, token, version, _ := mergeFixture(t)
	result, err := runtime.versions(context.Background(), messagingtest.MustEnvelope(t, contract.ApplicationRequest{AccessToken: token, Application: "example"}))
	if err != nil {
		t.Fatal(err)
	}
	if got := result.(contract.VersionsResponse).Versions; len(got) != 1 || got[0] != version {
		t.Fatalf("versions = %v", got)
	}
}

// Requirements: backup-restore/FR-001.
func TestMetadataRPCReturnsStoredSnapshotDetails(t *testing.T) {
	runtime, token, version, _ := mergeFixture(t)
	result, err := runtime.metadata(context.Background(), messagingtest.MustEnvelope(t, contract.VersionRequest{AccessToken: token, Application: "example", Version: version}))
	if err != nil {
		t.Fatal(err)
	}
	if got := result.(contract.BackupMetadataResponse).Version; got != version {
		t.Fatalf("metadata version = %q, want %q", got, version)
	}
}

// Requirements: backup-restore/FR-009.
func TestDeleteBackupRPCRemovesStoredSnapshot(t *testing.T) {
	runtime, token, version, _ := mergeFixture(t)
	result, err := runtime.deleteBackup(context.Background(), messagingtest.MustEnvelope(t, contract.VersionRequest{AccessToken: token, Application: "example", Version: version}))
	if err != nil {
		t.Fatal(err)
	}
	if !result.(contract.DeleteResponse).Deleted {
		t.Fatal("delete response did not confirm deletion")
	}
}

// Requirements: backup-restore/SR-003.
func TestMetadataRPCRejectsUnknownApplication(t *testing.T) {
	runtime, token, version, _ := mergeFixture(t)
	_, err := runtime.metadata(context.Background(), messagingtest.MustEnvelope(t, contract.VersionRequest{AccessToken: token, Application: "unknown", Version: version}))
	if err == nil {
		t.Fatal("metadata for an unknown application was returned")
	}
}

// Requirements: backup-restore/SR-003.
func TestDeleteBackupRPCRejectsUnknownApplication(t *testing.T) {
	runtime, token, version, _ := mergeFixture(t)
	_, err := runtime.deleteBackup(context.Background(), messagingtest.MustEnvelope(t, contract.VersionRequest{AccessToken: token, Application: "unknown", Version: version}))
	if err == nil {
		t.Fatal("backup for an unknown application was deleted")
	}
}

// Requirements: backup-restore/FR-010.
func TestDiffRPCReportsChangedFile(t *testing.T) {
	runtime, token, version, path := mergeFixture(t)
	result, err := runtime.diff(context.Background(), messagingtest.MustEnvelope(t, contract.VersionRequest{AccessToken: token, Application: "example", Version: version}))
	if err != nil {
		t.Fatal(err)
	}
	if got := result.(contract.DiffResponse).Differences; len(got) != 1 || got[0].Path != path {
		t.Fatalf("differences = %v", got)
	}
}

// Requirements: backup-restore/FR-001.
func TestBackupRPCPublishesVersion(t *testing.T) {
	runtime, token, _, _ := mergeFixture(t)
	result, err := runtime.backup(context.Background(), messagingtest.MustEnvelope(t, contract.BackupRequest{AccessToken: token, Application: "example"}))
	if err != nil {
		t.Fatal(err)
	}
	if result.(contract.BackupResponse).Version == "" {
		t.Fatal("backup version is empty")
	}
}

// Requirements: backup-restore/FR-009.
func TestBackedUpRPCListsStoredTarget(t *testing.T) {
	runtime, token, _, _ := mergeFixture(t)
	result, err := runtime.backedUp(context.Background(), messagingtest.MustEnvelope(t, contract.ApplicationsRequest{AccessToken: token}))
	if err != nil {
		t.Fatal(err)
	}
	if got := result.(contract.ApplicationsResponse).Applications; len(got) != 1 || got[0] != "example" {
		t.Fatalf("backed-up applications = %v", got)
	}
}

// Requirements: backup-restore/FR-001.
func TestBackupAllRPCReportsCompletedTarget(t *testing.T) {
	runtime, token, _, _ := mergeFixture(t)
	result, err := runtime.backupAll(context.Background(), messagingtest.MustEnvelope(t, contract.BackupAllRequest{AccessToken: token}))
	if err != nil {
		t.Fatal(err)
	}
	if got := result.(contract.BackupAllResponse).Results; len(got) != 1 || got[0].Status != "completed" {
		t.Fatalf("backup-all results = %v", got)
	}
}

// Requirements: backup-restore/FR-001.
func TestBackupAllRPCSkipsMissingTarget(t *testing.T) {
	verifier, token := authtokentest.MustIssueAccessToken(t, []string{"admin"})
	runtime := &Runtime{specs: map[string]Spec{"example": {Application: "example", Selections: []Selection{{Path: filepath.Join(t.TempDir(), "missing.conf")}}}}, store: Store{Root: t.TempDir()}, verifier: verifier, validate: validator.New(validator.WithRequiredStructEnabled()), client: &messagingtest.RecordingClient{AuthUsername: "tester"}}
	result, err := runtime.backupAll(context.Background(), messagingtest.MustEnvelope(t, contract.BackupAllRequest{AccessToken: token}))
	if err != nil {
		t.Fatal(err)
	}
	if got := result.(contract.BackupAllResponse).Results[0].Status; got != "skipped" {
		t.Fatalf("backup-all status = %q, want skipped", got)
	}
}

// Requirements: backup-restore/FR-001.
func TestBackupAllRPCReportsFailedTarget(t *testing.T) {
	verifier, token := authtokentest.MustIssueAccessToken(t, []string{"admin"})
	path := filepath.Join(t.TempDir(), "settings.conf")
	if err := os.WriteFile(path, []byte("enabled=yes\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runtime := &Runtime{specs: map[string]Spec{"example": {Application: "invalid_name", Selections: []Selection{{Path: path}}}}, store: Store{Root: t.TempDir()}, verifier: verifier, validate: validator.New(validator.WithRequiredStructEnabled()), client: &messagingtest.RecordingClient{AuthUsername: "tester"}}
	result, err := runtime.backupAll(context.Background(), messagingtest.MustEnvelope(t, contract.BackupAllRequest{AccessToken: token}))
	if err != nil {
		t.Fatal(err)
	}
	if got := result.(contract.BackupAllResponse).Results[0].Status; got != "failed" {
		t.Fatalf("backup-all status = %q, want failed", got)
	}
}

// Requirements: backup-restore/IR-002.
func TestRegisterPropagatesRPCRegistrationFailures(t *testing.T) {
	subjects := []string{
		contract.MergeStartSubject,
		contract.MergeReadSubject,
		contract.MergeStatusSubject,
		contract.MergeWriteSubject,
		contract.MergeApplySubject,
		contract.MergeSaveSubject,
		contract.MergeDiscardSubject,
		contract.ApplicationsSubject,
		contract.BackedUpSubject,
		contract.VersionsSubject,
		contract.MetadataSubject,
		contract.DeleteSubject,
		contract.BackupSubject,
		contract.BackupAllSubject,
		contract.DiffSubject,
		contract.RestoreSubject,
		contract.RestoreByNameSubject,
		contract.RestoreAllByNameSubject,
	}
	for _, subject := range subjects {
		t.Run(subject, func(t *testing.T) {
			server := &failingRegistrationServer{subject: subject}
			if err := new(Runtime).register(server); err == nil {
				t.Fatalf("registration failure for %s was ignored", subject)
			}
		})
	}
}

// Requirements: backup-restore/SR-003.
func TestDecodeRequestRequiresAdministratorRole(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	issuer, err := authtoken.NewIssuer(authtoken.IssuerOptions{Issuer: "test", Audience: "test", AccessLifetime: time.Minute}, privateKey)
	if err != nil {
		t.Fatal(err)
	}
	verifier, err := authtoken.NewVerifier(authtoken.VerifierOptions{Issuer: "test", Audience: "test"}, publicKey)
	if err != nil {
		t.Fatal(err)
	}
	runtime := &Runtime{verifier: verifier, validate: validator.New(validator.WithRequiredStructEnabled()), client: &messagingtest.RecordingClient{AuthUsername: "test"}}
	cases := []struct {
		name  string
		login string
		roles []string
		allow bool
	}{
		{name: "admin", roles: []string{"admin"}, allow: true},
		{name: "sudo", roles: []string{"sudo"}, allow: true},
		{name: "dedicated CLI identity", login: "lite-nas-backup-restore-cli", allow: false},
		{name: "security officer", roles: []string{"lite-nas-security"}, allow: false},
		{name: "operator", roles: []string{"lite-nas-operator"}, allow: false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			login := testCase.login
			if login == "" {
				login = "test"
			}
			token, _, err := issuer.Issue(authtoken.Principal{Subject: "123", Login: login, Roles: testCase.roles})
			if err != nil {
				t.Fatal(err)
			}
			payload, err := json.Marshal(contract.ApplicationsRequest{AccessToken: token})
			if err != nil {
				t.Fatal(err)
			}
			_, err = decodeRequest[contract.ApplicationsRequest](runtime, messaging.Envelope{Payload: payload})
			if (err == nil) != testCase.allow {
				t.Fatalf("authorization error = %v, allow = %v", err, testCase.allow)
			}
		})
	}
}

// Requirements: backup-restore/SR-003.
func TestRuntimePathsRejectsRelativeDirectory(t *testing.T) {
	configuration, err := ini.Load([]byte("[backup]\nspec_dir=relative\nstore_root=/var/lib/lite-nas/backup\n"))
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := runtimePaths(configuration); err == nil {
		t.Fatal("relative spec directory was accepted")
	}
}

// Requirements: backup-restore/SR-003.
func TestRuntimePathsReturnsConfiguredDirectories(t *testing.T) {
	configuration, err := ini.Load([]byte("[backup]\nspec_dir=/etc/lite-nas/backup.d\nstore_root=/var/lib/lite-nas/backup\n"))
	if err != nil {
		t.Fatal(err)
	}
	specDirectory, storeRoot, err := runtimePaths(configuration)
	if err != nil {
		t.Fatal(err)
	}
	if specDirectory != "/etc/lite-nas/backup.d" || storeRoot != "/var/lib/lite-nas/backup" {
		t.Fatalf("runtime paths = %q, %q", specDirectory, storeRoot)
	}
}

// Requirements: backup-restore/SR-003.
func TestServeRejectsNonRootCaller(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root callers pass the service privilege check")
	}
	if err := Serve(context.Background(), filepath.Join(t.TempDir(), "missing.conf")); err == nil {
		t.Fatal("service started for a non-root caller")
	}
}
