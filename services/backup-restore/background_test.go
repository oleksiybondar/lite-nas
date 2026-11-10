package backuprestore

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	contract "lite-nas/shared/contracts/backuprestore"
)

// TestBackupTaskRunnerPersistsCompletion verifies task state survives a service restart.
// Requirements: backup-restore/FR-013.
func TestBackupTaskRunnerPersistsCompletion(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "example.conf")
	if err := os.WriteFile(source, []byte("enabled=yes\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runtime := &Runtime{
		specs: map[string]Spec{"example": {Application: "example", Plugin: "file", Selections: []Selection{{Path: source}}}},
		store: Store{Root: filepath.Join(root, "backups")},
	}
	runner, err := newBackupTaskRunner(runtime.store.Root, runtime)
	if err != nil {
		t.Fatal(err)
	}
	go runner.Run(context.Background())
	defer runner.Close()
	started, err := runner.Start("example", false, "nightly", "", "admin")
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		status, err := runner.Status(started.TaskID)
		if err != nil {
			t.Fatal(err)
		}
		if status.State == "succeeded" || status.State == "failed" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("background backup did not finish")
		}
		time.Sleep(10 * time.Millisecond)
	}
	runner.Close()
	restored, err := newBackupTaskRunner(runtime.store.Root, runtime)
	if err != nil {
		t.Fatal(err)
	}
	status, err := restored.Status(started.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	if status.State != "succeeded" {
		t.Fatalf("restored task state = %q, want succeeded", status.State)
	}
}

// TestBackupTaskRunnerMarksQueuedWorkFailedAfterRestart preserves interruption state.
// Requirements: backup-restore/FR-013.
func TestBackupTaskRunnerMarksQueuedWorkFailedAfterRestart(t *testing.T) {
	root := t.TempDir()
	runtime := &Runtime{store: Store{Root: filepath.Join(root, "backups")}}
	runner, err := newBackupTaskRunner(runtime.store.Root, runtime)
	if err != nil {
		t.Fatal(err)
	}
	started, err := runner.Start("example", false, "", "", "admin")
	if err != nil {
		t.Fatal(err)
	}
	restarted, err := newBackupTaskRunner(runtime.store.Root, runtime)
	if err != nil {
		t.Fatal(err)
	}
	status, err := restarted.Status(started.TaskID)
	if err != nil {
		t.Fatal(err)
	}
	if status.State != "failed" {
		t.Fatalf("interrupted task state = %q, want failed", status.State)
	}
}

// TestBackupTaskRunnerReportsAllTargetProgress verifies bounded per-application progress updates.
// Requirements: backup-restore/FR-013.
func TestBackupTaskRunnerReportsAllTargetProgress(t *testing.T) {
	status := mustCompletedAllBackupTask(t)
	if status.Progress == nil || status.Progress.Percent == nil || *status.Progress.Percent != 100 {
		t.Fatalf("final progress = %+v, want 100 percent", status.Progress)
	}
}

// TestBackupTaskRunnerReturnsAllTargetOutcomes verifies each configured application's result.
// Requirements: backup-restore/FR-013.
func TestBackupTaskRunnerReturnsAllTargetOutcomes(t *testing.T) {
	status := mustCompletedAllBackupTask(t)
	got := make([]string, 0, len(status.Results))
	for _, result := range status.Results {
		got = append(got, result.Status)
	}
	if !reflect.DeepEqual(got, []string{"completed", "completed"}) {
		t.Fatalf("target outcomes = %v, want two completed targets", got)
	}
}

// mustCompletedAllBackupTask runs a two-target batch and returns its terminal snapshot.
func mustCompletedAllBackupTask(t *testing.T) contract.BackupTaskStatusResponse {
	t.Helper()
	root := t.TempDir()
	specs := make(map[string]Spec, 2)
	for _, application := range []string{"alpha", "beta"} {
		path := filepath.Join(root, application+".conf")
		if err := os.WriteFile(path, []byte("enabled=yes\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		specs[application] = Spec{Application: application, Plugin: "file", Selections: []Selection{{Path: path}}}
	}
	runtime := &Runtime{specs: specs, store: Store{Root: filepath.Join(root, "backups")}}
	runner, err := newBackupTaskRunner(runtime.store.Root, runtime)
	if err != nil {
		t.Fatal(err)
	}
	go runner.Run(context.Background())
	t.Cleanup(runner.Close)
	started, err := runner.Start("", true, "nightly", "", "admin")
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		status, err := runner.Status(started.TaskID)
		if err != nil {
			t.Fatal(err)
		}
		if status.State == "succeeded" || status.State == "failed" {
			runner.Close()
			return status
		}
		if time.Now().After(deadline) {
			runner.Close()
			t.Fatal("all-application backup did not finish")
		}
		time.Sleep(10 * time.Millisecond)
	}
}
