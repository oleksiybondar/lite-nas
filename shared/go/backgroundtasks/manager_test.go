package backgroundtasks

import (
	"context"
	"errors"
	"testing"
	"time"

	"lite-nas/shared/workers"
)

const printfCommand = "/usr/bin/printf"

// newRunningManager prepares a manager with its command and result workers active.
func newRunningManager(t *testing.T) (*Manager, context.CancelFunc, <-chan struct{}) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	manager := NewManager(ctx)
	done := make(chan struct{})
	workers.NewBaseWorker(func(workerCtx context.Context) {
		defer close(done)
		manager.Run(workerCtx)
	}).Start(ctx)
	<-manager.Ready()
	t.Cleanup(func() {
		manager.Close()
		cancel()
		<-done
		manager.Wait()
	})
	return manager, cancel, done
}

// mustStartTask submits a trusted test command and fails setup if admission fails.
func mustStartTask(t *testing.T, manager *Manager, spec CommandSpec) Snapshot {
	t.Helper()
	snapshot, err := manager.Start(spec)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	return snapshot
}

// mustRetryTask starts an expected-success retry attempt.
func mustRetryTask(t *testing.T, manager *Manager, id string) Snapshot {
	t.Helper()
	snapshot, err := manager.Retry(id)
	if err != nil {
		t.Fatalf("Retry() error = %v", err)
	}
	return snapshot
}

// mustWaitForTerminalTask polls until the result consumer records a terminal snapshot.
func mustWaitForTerminalTask(t *testing.T, manager *Manager, id string) Snapshot {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		snapshot, err := manager.Status(id)
		if err != nil {
			t.Fatalf("Status() error = %v", err)
		}
		if snapshot.State != StateQueued && snapshot.State != StateRunning {
			return snapshot
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("task %q did not finish before timeout", id)
	return Snapshot{}
}

// mustWaitForTaskState waits for a task to reach an expected lifecycle precondition.
func mustWaitForTaskState(t *testing.T, manager *Manager, id string, want State) Snapshot {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		snapshot, err := manager.Status(id)
		if err != nil {
			t.Fatalf("Status() error = %v", err)
		}
		if snapshot.State == want {
			return snapshot
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("task %q did not reach state %q", id, want)
	return Snapshot{}
}

// mustGetTaskSnapshot returns a known task or fails the test setup.
func mustGetTaskSnapshot(t *testing.T, manager *Manager, id string) Snapshot {
	t.Helper()
	snapshot, err := manager.Status(id)
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	return snapshot
}

// TestManagerRunsTasksThroughTheWorkerPipeline verifies successful command completion.
// Requirements: shared-background-tasks/FR-002
func TestManagerRunsTasksThroughTheWorkerPipeline(t *testing.T) {
	manager, _, _ := newRunningManager(t)
	task := mustStartTask(t, manager, CommandSpec{Name: printfCommand, Args: []string{"task-output"}})
	finished := mustWaitForTerminalTask(t, manager, task.ID)
	if finished.State != StateSucceeded {
		t.Fatalf("State = %q, want %q", finished.State, StateSucceeded)
	}
}

// TestManagerPublishesTerminalSnapshots verifies completed tasks flow through the public result pipe.
// Requirements: shared-background-tasks/FR-002
func TestManagerPublishesTerminalSnapshots(t *testing.T) {
	manager, _, _ := newRunningManager(t)
	task := mustStartTask(t, manager, CommandSpec{Name: printfCommand, Args: []string{"task-output"}})
	select {
	case result := <-manager.Results():
		if result.ID != task.ID || result.State != StateSucceeded {
			t.Fatalf("terminal result = %+v, want task %q succeeded", result, task.ID)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("manager did not publish a terminal result")
	}
}

// TestManagerReturnsCapturedTaskOutput verifies output reaches the pollable result.
// Requirements: shared-background-tasks/FR-002
func TestManagerReturnsCapturedTaskOutput(t *testing.T) {
	manager, _, _ := newRunningManager(t)
	task := mustStartTask(t, manager, CommandSpec{Name: printfCommand, Args: []string{"task-output"}})
	finished := mustWaitForTerminalTask(t, manager, task.ID)
	if finished.Output != "task-output" {
		t.Fatalf("Output = %q, want %q", finished.Output, "task-output")
	}
}

// TestManagerBoundsCapturedOutput verifies that task output is retained as a bounded tail.
// Requirements: shared-background-tasks/FR-002
func TestManagerBoundsCapturedOutput(t *testing.T) {
	manager, _, _ := newRunningManager(t)
	task := mustStartTask(t, manager, CommandSpec{Name: printfCommand, Args: []string{"abcdefghijk"}, OutputLimit: 8})
	finished := mustWaitForTerminalTask(t, manager, task.ID)
	actual := struct {
		Output    string
		Truncated bool
	}{Output: finished.Output, Truncated: finished.OutputTruncated}
	want := struct {
		Output    string
		Truncated bool
	}{Output: "defghijk", Truncated: true}
	if actual != want {
		t.Fatalf("bounded output = %+v, want %+v", actual, want)
	}
}

// TestManagerRetryCreatesANewAttempt verifies retry preserves attempt identity boundaries.
// Requirements: shared-background-tasks/FR-003
func TestManagerRetryCreatesANewAttempt(t *testing.T) {
	manager, _, _ := newRunningManager(t)
	first := mustStartTask(t, manager, CommandSpec{Name: printfCommand, Args: []string{"retry"}})
	_ = mustWaitForTerminalTask(t, manager, first.ID)
	second := mustRetryTask(t, manager, first.ID)
	if first.ID == second.ID {
		t.Fatalf("retry task ID = %q, want a new task ID", second.ID)
	}
}

// TestManagerRejectsTasksWhenTheBoundedInputQueueIsFull verifies explicit backpressure.
// Requirements: shared-background-tasks/RR-001
func TestManagerRejectsTasksWhenTheBoundedInputQueueIsFull(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	manager := NewManager(ctx)
	for range defaultQueueCapacity {
		mustStartTask(t, manager, CommandSpec{Name: printfCommand, Args: []string{"queued"}})
	}
	_, err := manager.Start(CommandSpec{Name: printfCommand, Args: []string{"overflow"}})
	if !errors.Is(err, ErrQueueFull) {
		t.Fatalf("Start() error = %v, want %v", err, ErrQueueFull)
	}
	manager.Close()
	cancel()
	manager.Run(ctx)
	manager.Wait()
}

// TestManagerMarksQueuedTasksCancelledDuringShutdown verifies queued work is accounted for.
// Requirements: shared-background-tasks/RR-002
func TestManagerMarksQueuedTasksCancelledDuringShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	manager := NewManager(ctx)
	task := mustStartTask(t, manager, CommandSpec{Name: printfCommand, Args: []string{"queued"}})
	cancel()
	manager.Close()
	manager.Run(ctx)
	manager.Wait()
	finished := mustGetTaskSnapshot(t, manager, task.ID)
	if finished.State != StateCancelled {
		t.Fatalf("State = %q, want %q", finished.State, StateCancelled)
	}
}

// TestManagerCancelsTheActiveCommand verifies parent cancellation reaches a running process.
// Requirements: shared-background-tasks/RR-002
func TestManagerCancelsTheActiveCommand(t *testing.T) {
	manager, cancel, _ := newRunningManager(t)
	task := mustStartTask(t, manager, CommandSpec{Name: "/usr/bin/sleep", Args: []string{"30"}})
	_ = mustWaitForTaskState(t, manager, task.ID, StateRunning)
	cancel()
	manager.Close()
	manager.Wait()
	finished := mustGetTaskSnapshot(t, manager, task.ID)
	if finished.State != StateCancelled {
		t.Fatalf("State = %q, want %q", finished.State, StateCancelled)
	}
}
