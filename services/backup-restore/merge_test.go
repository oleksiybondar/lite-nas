package backuprestore

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/go-playground/validator/v10"

	contract "lite-nas/shared/contracts/backuprestore"
	"lite-nas/shared/testutil/authtokentest"
	"lite-nas/shared/testutil/messagingtest"
)

// mergeFixture prepares one selected text file and an administrator session.
func mergeFixture(t *testing.T) (*Runtime, string, string, string) {
	t.Helper()
	root := t.TempDir()
	path := filepath.Join(root, "settings.conf")
	if err := os.WriteFile(path, []byte("snapshot\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	spec := Spec{Application: "example", Selections: []Selection{{Path: path}}}
	store := Store{Root: filepath.Join(root, "backups")}
	version, err := store.Backup(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("current\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	verifier, token := authtokentest.MustIssueAccessToken(t, []string{"admin"})
	runtime := &Runtime{specs: map[string]Spec{"example": spec}, store: store, verifier: verifier, validate: validator.New(validator.WithRequiredStructEnabled())}
	return runtime, token, version, path
}

// mustStartMerge starts a validated session for fixture flows.
func mustStartMerge(t *testing.T, runtime *Runtime, token, version string) string {
	t.Helper()
	result, err := runtime.mergeStart(context.Background(), messagingtest.MustEnvelope(t, contract.MergeStartRequest{AccessToken: token, Application: "example", Version: version}))
	if err != nil {
		t.Fatal(err)
	}
	return result.(contract.MergeStartResponse).SessionID
}

// mustStageMerge uploads one complete resolved result.
func mustStageMerge(t *testing.T, runtime *Runtime, token, id, path, content string) {
	t.Helper()
	checksum := sha256.Sum256([]byte(content))
	request := contract.MergeWriteRequest{AccessToken: token, SessionID: id, Path: path, Data: base64.StdEncoding.EncodeToString([]byte(content)), Final: true, SHA256: hex.EncodeToString(checksum[:])}
	if _, err := runtime.mergeWrite(context.Background(), messagingtest.MustEnvelope(t, request)); err != nil {
		t.Fatal(err)
	}
}

// Requirements: backup-restore/FR-005.
func TestMergeStartRejectsSecondGlobalSession(t *testing.T) {
	runtime, token, version, _ := mergeFixture(t)
	mustStartMerge(t, runtime, token, version)
	_, err := runtime.mergeStart(context.Background(), messagingtest.MustEnvelope(t, contract.MergeStartRequest{AccessToken: token, Application: "example", Version: version}))
	if err == nil {
		t.Fatal("second session accepted")
	}
}

// Requirements: backup-restore/FR-005.
func TestMergeStartAcceptsMoreThanThirtyTwoSmallFiles(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "settings.d")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	for index := range 40 {
		path := filepath.Join(directory, fmt.Sprintf("%02d.conf", index))
		if err := os.WriteFile(path, []byte("enabled=yes\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	spec := Spec{Application: "example", Selections: []Selection{{Path: directory, Directory: true}}}
	store := Store{Root: filepath.Join(root, "backups")}
	version, err := store.Backup(spec)
	if err != nil {
		t.Fatal(err)
	}
	verifier, token := authtokentest.MustIssueAccessToken(t, []string{"admin"})
	runtime := &Runtime{specs: map[string]Spec{"example": spec}, store: store, verifier: verifier, validate: validator.New(validator.WithRequiredStructEnabled())}
	result, err := runtime.mergeStart(context.Background(), messagingtest.MustEnvelope(t, contract.MergeStartRequest{AccessToken: token, Application: "example", Version: version}))
	if err != nil {
		t.Fatal(err)
	}
	if got := len(result.(contract.MergeStartResponse).Files); got != 40 {
		t.Fatalf("editable file count = %d, want 40", got)
	}
}

// Requirements: backup-restore/FR-005.
func TestConcurrentMergeStartsAdmitOneSession(t *testing.T) {
	runtime, token, version, _ := mergeFixture(t)
	envelope := messagingtest.MustEnvelope(t, contract.MergeStartRequest{AccessToken: token, Application: "example", Version: version})
	results := make(chan error, 2)
	var workers sync.WaitGroup
	for range 2 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			_, err := runtime.mergeStart(context.Background(), envelope)
			results <- err
		}()
	}
	workers.Wait()
	close(results)
	successes := 0
	for err := range results {
		if err == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("successful starts = %d", successes)
	}
}

// Requirements: backup-restore/IR-002.
func TestMergeSubjectsAreRegistered(t *testing.T) {
	server := &messagingtest.RecordingServer{}
	if err := new(Runtime).register(server); err != nil {
		t.Fatal(err)
	}
	for _, subject := range []string{contract.MergeStartSubject, contract.MergeStatusSubject, contract.MergeReadSubject, contract.MergeWriteSubject, contract.MergeApplySubject, contract.MergeSaveSubject, contract.MergeDiscardSubject} {
		t.Run(subject, func(t *testing.T) {
			if server.Handlers[subject] == nil {
				t.Fatal("merge RPC not registered")
			}
		})
	}
}

// Requirements: backup-restore/FR-005.
func TestMergeDiscardReleasesGlobalSession(t *testing.T) {
	runtime, token, version, _ := mergeFixture(t)
	id := mustStartMerge(t, runtime, token, version)
	if _, err := runtime.mergeDiscard(context.Background(), messagingtest.MustEnvelope(t, contract.MergeActionRequest{AccessToken: token, SessionID: id})); err != nil {
		t.Fatal(err)
	}
	_, err := runtime.mergeStart(context.Background(), messagingtest.MustEnvelope(t, contract.MergeStartRequest{AccessToken: token, Application: "example", Version: version}))
	if err != nil {
		t.Fatalf("new session rejected after discard: %v", err)
	}
}

// Requirements: backup-restore/SR-003.
func TestMergeReadRequiresSessionOwner(t *testing.T) {
	runtime, token, version, path := mergeFixture(t)
	id := mustStartMerge(t, runtime, token, version)
	runtime.session.owner = "another-subject"
	_, err := runtime.mergeRead(context.Background(), messagingtest.MustEnvelope(t, contract.MergeReadRequest{AccessToken: token, SessionID: id, Path: path, Side: "snapshot"}))
	if err == nil {
		t.Fatal("non-owner read the session")
	}
}

// Requirements: backup-restore/FR-005, backup-restore/IR-002.
func TestMergeReadReturnsChecksummedSnapshotChunk(t *testing.T) {
	runtime, token, version, path := mergeFixture(t)
	id := mustStartMerge(t, runtime, token, version)
	result, err := runtime.mergeRead(context.Background(), messagingtest.MustEnvelope(t, contract.MergeReadRequest{AccessToken: token, SessionID: id, Path: path, Side: "snapshot"}))
	if err != nil {
		t.Fatal(err)
	}
	response := result.(contract.MergeReadResponse)
	content, err := base64.StdEncoding.DecodeString(response.Data)
	if err != nil {
		t.Fatal(err)
	}
	checksum := sha256.Sum256(content)
	if string(content) != "snapshot\n" || !response.Done || response.SHA256 != hex.EncodeToString(checksum[:]) {
		t.Fatalf("snapshot chunk = %+v", response)
	}
}

// Requirements: backup-restore/SR-003.
func TestMergeReadRejectsUnselectedPath(t *testing.T) {
	runtime, token, version, _ := mergeFixture(t)
	id := mustStartMerge(t, runtime, token, version)
	_, err := runtime.mergeRead(context.Background(), messagingtest.MustEnvelope(t, contract.MergeReadRequest{AccessToken: token, SessionID: id, Path: "/etc/shadow", Side: "snapshot"}))
	if err == nil {
		t.Fatal("unselected path was read")
	}
}

// Requirements: backup-restore/SR-003.
func TestMergeWriteRejectsWrongChecksum(t *testing.T) {
	runtime, token, version, path := mergeFixture(t)
	id := mustStartMerge(t, runtime, token, version)
	request := contract.MergeWriteRequest{AccessToken: token, SessionID: id, Path: path, Data: base64.StdEncoding.EncodeToString([]byte("resolved")), Final: true, SHA256: strings.Repeat("0", 64)}
	_, err := runtime.mergeWrite(context.Background(), messagingtest.MustEnvelope(t, request))
	if err == nil {
		t.Fatal("wrong checksum accepted")
	}
}

// Requirements: backup-restore/FR-006.
func TestMergeSaveUsesStagedContentWithoutApplyingLive(t *testing.T) {
	runtime, token, version, path := mergeFixture(t)
	id := mustStartMerge(t, runtime, token, version)
	mustStageMerge(t, runtime, token, id, path, "resolved\n")
	result, err := runtime.mergeSave(context.Background(), messagingtest.MustEnvelope(t, contract.MergeSaveRequest{AccessToken: token, SessionID: id}))
	if err != nil {
		t.Fatal(err)
	}
	archive, err := runtime.store.OpenVersion("example", result.(contract.BackupResponse).Version)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	stage := t.TempDir()
	entries, err := stageArchive(archive, stage)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.path == path {
			content, err := os.ReadFile(entry.stage)
			if err != nil {
				t.Fatal(err)
			}
			if string(content) != "resolved\n" {
				t.Fatalf("saved content = %q", content)
			}
			return
		}
	}
	t.Fatal("saved path missing")
}

// Requirements: backup-restore/FR-006.
func TestMergeSaveLeavesCurrentFileUnchanged(t *testing.T) {
	runtime, token, version, path := mergeFixture(t)
	id := mustStartMerge(t, runtime, token, version)
	mustStageMerge(t, runtime, token, id, path, "resolved\n")
	if _, err := runtime.mergeSave(context.Background(), messagingtest.MustEnvelope(t, contract.MergeSaveRequest{AccessToken: token, SessionID: id})); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "current\n" {
		t.Fatalf("live content = %q", content)
	}
}

// Requirements: backup-restore/FR-005.
func TestMergeApplyRejectsChangedLiveInput(t *testing.T) {
	runtime, token, version, path := mergeFixture(t)
	id := mustStartMerge(t, runtime, token, version)
	mustStageMerge(t, runtime, token, id, path, "resolved\n")
	if err := os.WriteFile(path, []byte("changed\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := runtime.mergeApply(context.Background(), messagingtest.MustEnvelope(t, contract.MergeActionRequest{AccessToken: token, SessionID: id}))
	if err == nil {
		t.Fatal("changed live input was applied")
	}
}

// Requirements: backup-restore/FR-005.
func TestMergeConflictKeepsSessionForSave(t *testing.T) {
	runtime, token, version, path := mergeFixture(t)
	id := mustStartMerge(t, runtime, token, version)
	mustStageMerge(t, runtime, token, id, path, "resolved\n")
	if err := os.WriteFile(path, []byte("changed\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, _ = runtime.mergeApply(context.Background(), messagingtest.MustEnvelope(t, contract.MergeActionRequest{AccessToken: token, SessionID: id}))
	result, err := runtime.mergeStatus(context.Background(), messagingtest.MustEnvelope(t, contract.MergeActionRequest{AccessToken: token, SessionID: id}))
	if err != nil || result.(contract.MergeStartResponse).ApplyAvailable {
		t.Fatalf("status after conflict = %#v, %v", result, err)
	}
}

// Requirements: backup-restore/FR-005.
func TestMergeApplyWritesStagedResult(t *testing.T) {
	runtime, token, version, path := mergeFixture(t)
	id := mustStartMerge(t, runtime, token, version)
	mustStageMerge(t, runtime, token, id, path, "resolved\n")
	if _, err := runtime.mergeApply(context.Background(), messagingtest.MustEnvelope(t, contract.MergeActionRequest{AccessToken: token, SessionID: id})); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "resolved\n" {
		t.Fatalf("live content = %q", content)
	}
}
