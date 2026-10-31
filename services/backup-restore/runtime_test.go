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

	"lite-nas/shared/authtoken"
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

// Request records a restart request and returns an observable failure.
func (client *failingRestartClient) Request(_ context.Context, subject string, request any, _ any) error {
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
	runtime := &Runtime{specs: map[string]Spec{"example": {Application: "example"}}, store: Store{Root: t.TempDir()}, verifier: verifier, validate: validator.New(validator.WithRequiredStructEnabled())}
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
	runtime := &Runtime{verifier: verifier, validate: validator.New(validator.WithRequiredStructEnabled())}
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
