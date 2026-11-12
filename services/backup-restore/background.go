package backuprestore

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	contract "lite-nas/shared/contracts/backuprestore"
)

const backupTaskQueueCapacity = 2

// backupTaskJob is an authenticated, service-owned request accepted by the background worker.
type backupTaskJob struct {
	status      contract.BackupTaskStatusResponse
	all         bool
	name        string
	description string
	createdBy   string
}

// backupTaskRunner owns bounded backup jobs and their durable status records.
type backupTaskRunner struct {
	root    string
	runtime *Runtime
	// jobs: authenticated RPC handlers produce; Run consumes; capacity is two;
	// Close owns closure after RPC drain and Run finishes all accepted jobs.
	jobs               chan backupTaskJob
	mu                 sync.RWMutex
	states             map[string]contract.BackupTaskStatusResponse
	closed             bool
	closeOnce          sync.Once
	done               chan struct{}
	onPersistenceError func(error)
}

// newBackupTaskRunner restores task history and marks interrupted work as failed.
func newBackupTaskRunner(root string, runtime *Runtime) (*backupTaskRunner, error) {
	directory := filepath.Join(root, ".tasks")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return nil, err
	}
	info, err := os.Lstat(directory)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("backup task state path must be a directory")
	}
	if err := os.Chmod(directory, 0o700); err != nil {
		return nil, err
	}
	runner := &backupTaskRunner{root: directory, runtime: runtime, jobs: make(chan backupTaskJob, backupTaskQueueCapacity), states: make(map[string]contract.BackupTaskStatusResponse), done: make(chan struct{})}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(directory, entry.Name())
		info, err := os.Lstat(path)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var status contract.BackupTaskStatusResponse
		if err := json.Unmarshal(data, &status); err != nil || !validTaskID(status.TaskID) || entry.Name() != status.TaskID+".json" {
			continue
		}
		if status.State == "queued" || status.State == "running" {
			status.State = "failed"
			status.Error = "backup service restarted before the task completed"
			finished := time.Now().UTC()
			status.FinishedAt, status.UpdatedAt = &finished, finished
			if err := runner.persist(status); err != nil {
				return nil, err
			}
		}
		runner.states[status.TaskID] = status
	}
	return runner, nil
}

// Start persists and queues a backup task, rejecting work when the bounded queue is full.
func (runner *backupTaskRunner) Start(application string, all bool, name, description, createdBy string) (contract.BackupTaskStartResponse, error) {
	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		return contract.BackupTaskStartResponse{}, fmt.Errorf("create backup task ID: %w", err)
	}
	now := time.Now().UTC()
	statusApplication := application
	if all {
		statusApplication = "all"
	}
	status := contract.BackupTaskStatusResponse{TaskID: hex.EncodeToString(idBytes), Application: statusApplication, State: "queued", StartedAt: now, UpdatedAt: now}
	job := backupTaskJob{status: status, all: all, name: name, description: description, createdBy: createdBy}
	runner.mu.Lock()
	defer runner.mu.Unlock()
	if runner.closed {
		return contract.BackupTaskStartResponse{}, fmt.Errorf("backup task runner is shutting down")
	}
	if err := runner.persist(status); err != nil {
		return contract.BackupTaskStartResponse{}, err
	}
	runner.states[status.TaskID] = status
	select {
	case runner.jobs <- job:
		return contract.BackupTaskStartResponse{TaskID: status.TaskID, State: status.State}, nil
	default:
		delete(runner.states, status.TaskID)
		_ = os.Remove(runner.path(status.TaskID))
		return contract.BackupTaskStartResponse{}, fmt.Errorf("backup task queue is full")
	}
}

// Status returns a copy of one durable task snapshot.
func (runner *backupTaskRunner) Status(taskID string) (contract.BackupTaskStatusResponse, error) {
	runner.mu.RLock()
	defer runner.mu.RUnlock()
	status, ok := runner.states[taskID]
	if !ok {
		return contract.BackupTaskStatusResponse{}, fmt.Errorf("backup task not found")
	}
	return status, nil
}

// Run consumes queued jobs until the runtime closes the queue after RPC drain.
func (runner *backupTaskRunner) Run(_ context.Context) {
	defer close(runner.done)
	for job := range runner.jobs {
		runner.execute(job)
	}
}

// Close stops admissions and waits for accepted jobs to finish before shutdown.
func (runner *backupTaskRunner) Close() {
	runner.closeOnce.Do(func() {
		runner.mu.Lock()
		runner.closed = true
		close(runner.jobs)
		runner.mu.Unlock()
		<-runner.done
	})
}

// execute serializes each background backup against restore, delete, and merge mutations.
func (runner *backupTaskRunner) execute(job backupTaskJob) {
	runner.update(job.status.TaskID, func(status *contract.BackupTaskStatusResponse) {
		status.State = "running"
		status.Progress = &contract.BackupTaskProgress{Message: "Capturing application configuration"}
	})
	runner.runtime.mutex.Lock()
	defer runner.runtime.mutex.Unlock()
	if job.all {
		runner.executeAll(job)
		return
	}
	spec, ok := runner.runtime.specs[job.status.Application]
	if !ok {
		runner.finish(job.status.TaskID, "failed", "unknown application", "", nil)
		return
	}
	version, err := runner.runtime.store.BackupWithMetadata(spec, job.name, job.description, job.createdBy)
	if err != nil {
		runner.finish(job.status.TaskID, "failed", err.Error(), "", nil)
		return
	}
	runner.finish(job.status.TaskID, "succeeded", "", version, nil)
}

// executeAll backs up configured applications in stable order and reports per-target outcomes.
func (runner *backupTaskRunner) executeAll(job backupTaskJob) {
	names := make([]string, 0, len(runner.runtime.specs))
	for name := range runner.runtime.specs {
		names = append(names, name)
	}
	sort.Strings(names)
	results := make([]contract.BackupAllResult, 0, len(names))
	for index, name := range names {
		spec := runner.runtime.specs[name]
		result := runner.runtime.backupOne(spec, job.name, job.description, job.createdBy)
		results = append(results, result)
		percent := (index + 1) * 100 / max(len(names), 1)
		runner.update(job.status.TaskID, func(status *contract.BackupTaskStatusResponse) {
			status.Progress = &contract.BackupTaskProgress{Percent: intPointer(percent), Message: "Processed " + name}
			status.Results = append([]contract.BackupAllResult(nil), results...)
		})
	}
	runner.finish(job.status.TaskID, "succeeded", "", "", results)
}

// backupOne applies the same plugin availability checks as the synchronous backup-all operation.
func (runtime *Runtime) backupOne(spec Spec, name, description, createdBy string) contract.BackupAllResult {
	result := contract.BackupAllResult{Application: spec.Application}
	present := true
	switch spec.Plugin {
	case "ufw":
		for _, path := range []string{"/usr/sbin/ufw", "/etc/ufw/ufw.conf", "/etc/default/ufw"} {
			if _, err := os.Stat(path); os.IsNotExist(err) {
				present = false
				break
			}
		}
	default:
		present = false
		for _, selection := range spec.Selections {
			if _, err := os.Lstat(selection.Path); err == nil || !os.IsNotExist(err) {
				present = true
				break
			}
		}
	}
	if !present {
		result.Status = "skipped"
		return result
	}
	version, err := runtime.store.BackupWithMetadata(spec, name, description, createdBy)
	if err != nil {
		result.Status, result.Error = "failed", err.Error()
	} else {
		result.Status, result.Version = "completed", version
	}
	return result
}

// update persists one task mutation while holding the runner state lock.
func (runner *backupTaskRunner) update(taskID string, mutate func(*contract.BackupTaskStatusResponse)) {
	runner.mu.Lock()
	defer runner.mu.Unlock()
	status, ok := runner.states[taskID]
	if !ok {
		return
	}
	mutate(&status)
	status.UpdatedAt = time.Now().UTC()
	if err := runner.persist(status); err != nil {
		status.Error = "backup status persistence failed: " + err.Error()
		if runner.onPersistenceError != nil {
			runner.onPersistenceError(err)
		}
	}
	runner.states[taskID] = status
}

// finish persists a terminal result and finish timestamp for one task.
func (runner *backupTaskRunner) finish(taskID, state, taskError, version string, results []contract.BackupAllResult) {
	runner.update(taskID, func(status *contract.BackupTaskStatusResponse) {
		finished := time.Now().UTC()
		status.State = state
		status.Error = taskError
		status.Version = version
		status.FinishedAt = &finished
		status.Progress = &contract.BackupTaskProgress{Percent: intPointer(100), Message: "Backup finished"}
		if results != nil {
			status.Results = append([]contract.BackupAllResult(nil), results...)
		}
	})
}

// persist atomically writes a private task snapshot beneath the backup store.
func (runner *backupTaskRunner) persist(status contract.BackupTaskStatusResponse) error {
	if err := os.MkdirAll(runner.root, 0o700); err != nil {
		return err
	}
	content, err := json.Marshal(status)
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(runner.root, ".task-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if err := file.Chmod(0o600); err != nil {
		_ = file.Close()
		return err
	}
	if _, err := file.Write(content); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), runner.path(status.TaskID))
}

// path maps a validated task ID to its private status record.
func (runner *backupTaskRunner) path(taskID string) string {
	return filepath.Join(runner.root, taskID+".json")
}

// validTaskID accepts only the random 128-bit lowercase hexadecimal identifiers used by the runner.
func validTaskID(taskID string) bool {
	if len(taskID) != 32 || strings.ToLower(taskID) != taskID {
		return false
	}
	_, err := hex.DecodeString(taskID)
	return err == nil
}

// intPointer creates a pointer for optional progress percentages.
func intPointer(value int) *int { return &value }
