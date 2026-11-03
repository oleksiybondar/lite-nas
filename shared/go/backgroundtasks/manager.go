// Package backgroundtasks runs trusted, configured command jobs and exposes
// their in-process state for on-demand polling.
package backgroundtasks

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"lite-nas/shared/workers"
)

var (
	// ErrNotFound identifies a task ID not held by this manager.
	ErrNotFound = errors.New("background task not found")
	// ErrNotFinished prevents retrying a command while its previous attempt runs.
	ErrNotFinished = errors.New("background task is still running")
	// ErrInvalidCommand identifies a missing executable in a trusted task spec.
	ErrInvalidCommand = errors.New("background task command is required")
	// ErrQueueFull prevents accepting a task when the bounded input queue is full.
	ErrQueueFull = errors.New("background task queue is full")
)

const (
	// defaultQueueCapacity bounds accepted command jobs waiting for the worker.
	defaultQueueCapacity = 32
	// defaultEventCapacity bounds task state events waiting for the state consumer.
	defaultEventCapacity = 64
	// defaultResultCapacity bounds completed snapshots awaiting service processing.
	defaultResultCapacity = 64
)

// State identifies the lifecycle state of one command attempt.
type State string

const (
	StateQueued    State = "queued"
	StateRunning   State = "running"
	StateSucceeded State = "succeeded"
	StateFailed    State = "failed"
	StateCancelled State = "cancelled"
)

// Progress describes optional command-specific progress. Generic commands
// leave it nil because the runner cannot infer completion percentage.
type Progress struct {
	Percent *int   `json:"percent,omitempty"`
	Message string `json:"message,omitempty"`
}

// CommandSpec contains a command selected and validated by the owning service.
// It is not a transport DTO and must never be populated directly from a user
// request. The command is executed without a shell.
type CommandSpec struct {
	Name string
	Args []string
	Dir  string
	Env  []string
	// Timeout bounds the command attempt when positive; the manager context remains the shutdown bound.
	Timeout        time.Duration
	OutputLimit    int
	ProgressParser func(output string) *Progress
	// BeforeRun revalidates service-owned assumptions immediately before execution.
	BeforeRun func() error
	// AfterRun can reject a result when service-owned assumptions changed during execution.
	AfterRun func() error
	// AfterSuccess runs only after a zero exit status, before the task is marked succeeded.
	AfterSuccess func() error
}

// Snapshot is a point-in-time view of one task attempt.
type Snapshot struct {
	ID              string     `json:"task_id"`
	State           State      `json:"state"`
	StartedAt       time.Time  `json:"started_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	FinishedAt      *time.Time `json:"finished_at,omitempty"`
	ExitCode        *int       `json:"exit_code,omitempty"`
	Error           string     `json:"error,omitempty"`
	Output          string     `json:"output"`
	OutputTruncated bool       `json:"output_truncated"`
	Progress        *Progress  `json:"progress,omitempty"`
}

// Manager owns command task state for the lifetime of its parent context.
// State is in-memory and is lost when the owning service process exits.
type Manager struct {
	ctx   context.Context
	mu    sync.RWMutex
	tasks map[string]*task
	// jobs: Manager.Start produces under mu; runCommands consumes; capacity is
	// defaultQueueCapacity; Manager.Close owns closure after admission stops.
	jobs chan *task
	// events: runCommands and command-output writers produce; runResults consumes;
	// capacity is defaultEventCapacity; runCommands closes after command writers end.
	events chan taskEvent
	// results: runResults produces terminal snapshots; the owning service consumes;
	// capacity is defaultResultCapacity; runResults owns closure after event drain.
	results chan Snapshot
	ready   chan struct{}
	closed  bool
	runOnce sync.Once
	wg      sync.WaitGroup
}

// task keeps one trusted command specification and its consumer-owned status.
type task struct {
	spec       CommandSpec
	mu         sync.Mutex
	progressMu sync.Mutex
	snapshot   Snapshot
	output     *boundedOutput
}

// taskEvent carries worker-owned state changes to the manager's result consumer.
type taskEvent struct {
	item         *task
	snapshot     Snapshot
	progress     *Progress
	updatedAt    time.Time
	progressOnly bool
}

// NewManager creates a process-local task manager with bounded input and event
// channels, bound to ctx. Run must be started before task RPCs are exposed.
func NewManager(ctx context.Context) *Manager {
	return &Manager{
		ctx: ctx, tasks: make(map[string]*task),
		jobs:    make(chan *task, defaultQueueCapacity),
		events:  make(chan taskEvent, defaultEventCapacity),
		results: make(chan Snapshot, defaultResultCapacity),
		ready:   make(chan struct{}),
	}
}

// Start validates a trusted command spec, records a new attempt, and enqueues it
// asynchronously. Queue saturation returns ErrQueueFull. Each task has its own
// output limit, defaulting to 1 MiB.
func (manager *Manager) Start(spec CommandSpec) (Snapshot, error) {
	if spec.Name == "" {
		return Snapshot{}, ErrInvalidCommand
	}
	if spec.OutputLimit <= 0 {
		spec.OutputLimit = 1 << 20
	}
	id, err := newID()
	if err != nil {
		return Snapshot{}, fmt.Errorf("create background task ID: %w", err)
	}
	now := time.Now().UTC()
	item := &task{
		spec:     spec,
		output:   &boundedOutput{limit: spec.OutputLimit},
		snapshot: Snapshot{ID: id, State: StateQueued, StartedAt: now, UpdatedAt: now, Output: ""},
	}
	item.output.onWrite = func() {
		manager.emitProgress(item)
	}
	manager.mu.Lock()
	if manager.closed || manager.ctx.Err() != nil {
		manager.mu.Unlock()
		return Snapshot{}, context.Canceled
	}
	manager.tasks[id] = item
	select {
	case manager.jobs <- item:
		manager.mu.Unlock()
	case <-manager.ctx.Done():
		delete(manager.tasks, id)
		manager.mu.Unlock()
		return Snapshot{}, context.Canceled
	default:
		delete(manager.tasks, id)
		manager.mu.Unlock()
		return Snapshot{}, ErrQueueFull
	}
	return manager.Status(id)
}

// Retry starts a fresh attempt with the same trusted command configuration and
// returns a new task ID. The original task record remains pollable.
func (manager *Manager) Retry(id string) (Snapshot, error) {
	manager.mu.RLock()
	item, ok := manager.tasks[id]
	manager.mu.RUnlock()
	if !ok {
		return Snapshot{}, ErrNotFound
	}
	item.mu.Lock()
	state := item.snapshot.State
	spec := cloneSpec(item.spec)
	item.mu.Unlock()
	if state == StateQueued || state == StateRunning {
		return Snapshot{}, ErrNotFinished
	}
	return manager.Start(spec)
}

// Status returns a consistent copy of one task's latest state and bounded output.
func (manager *Manager) Status(id string) (Snapshot, error) {
	manager.mu.RLock()
	item, ok := manager.tasks[id]
	manager.mu.RUnlock()
	if !ok {
		return Snapshot{}, ErrNotFound
	}
	item.mu.Lock()
	defer item.mu.Unlock()
	copy := item.snapshot
	copy.Output, copy.OutputTruncated = item.output.snapshot()
	if item.snapshot.FinishedAt != nil {
		finished := *item.snapshot.FinishedAt
		copy.FinishedAt = &finished
	}
	if item.snapshot.ExitCode != nil {
		code := *item.snapshot.ExitCode
		copy.ExitCode = &code
	}
	if item.snapshot.Progress != nil {
		progress := *item.snapshot.Progress
		if item.snapshot.Progress.Percent != nil {
			percent := *item.snapshot.Progress.Percent
			progress.Percent = &percent
		}
		copy.Progress = &progress
	}
	return copy, nil
}

// Results exposes terminal snapshots after the manager records them. The
// owning service must keep one consumer active until this channel closes.
func (manager *Manager) Results() <-chan Snapshot {
	return manager.results
}

// Close prevents new tasks from starting and closes the input channel. The run
// loop drains accepted jobs; parent-context cancellation marks queued jobs
// cancelled and terminates the active command.
func (manager *Manager) Close() {
	manager.mu.Lock()
	if !manager.closed {
		manager.closed = true
		close(manager.jobs)
	}
	manager.mu.Unlock()
}

// Run starts the command worker and consumes its result events. Start the run
// loop before exposing task RPC handlers so accepted jobs always have consumers.
// The command worker drains queued jobs as cancelled when ctx ends.
func (manager *Manager) Run(ctx context.Context) {
	manager.runOnce.Do(func() {
		manager.wg.Add(2)
		workers.NewBaseWorker(func(workerCtx context.Context) {
			defer manager.wg.Done()
			manager.runCommands(workerCtx)
		}).Start(ctx)
		close(manager.ready)
		defer manager.wg.Done()
		manager.runResults()
	})
}

// Ready closes once Run has started its command worker and result consumer.
// Service runtimes should wait for readiness before accepting task requests.
func (manager *Manager) Ready() <-chan struct{} {
	return manager.ready
}

// Wait blocks until the command worker has drained accepted jobs and the result
// consumer has published every terminal snapshot. Close admission first and
// keep a Results consumer active until it closes.
func (manager *Manager) Wait() {
	manager.wg.Wait()
}

// runCommands consumes task input, executes each command, and owns event-channel closure.
func (manager *Manager) runCommands(ctx context.Context) {
	defer close(manager.events)
	for {
		if ctx.Err() != nil {
			manager.Close()
			for item := range manager.jobs {
				manager.emitTerminal(item, StateCancelled, ctx.Err(), nil)
			}
			return
		}
		select {
		case <-ctx.Done():
			continue
		case item, ok := <-manager.jobs:
			if !ok {
				return
			}
			manager.runCommand(ctx, item)
		}
	}
}

// runCommand emits lifecycle events around one trusted command execution.
func (manager *Manager) runCommand(ctx context.Context, item *task) {
	started := item.snapshot
	started.State = StateRunning
	started.UpdatedAt = time.Now().UTC()
	// Keep terminal-state accounting reliable by applying bounded backpressure
	// until the continuously running result consumer accepts this transition.
	manager.events <- taskEvent{item: item, snapshot: started}
	if err := ctx.Err(); err != nil {
		manager.emitTerminal(item, StateCancelled, err, nil)
		return
	}
	if item.spec.BeforeRun != nil {
		if err := item.spec.BeforeRun(); err != nil {
			manager.emitTerminal(item, StateFailed, err, nil)
			return
		}
	}
	commandCtx := ctx
	cancel := func() {}
	if item.spec.Timeout > 0 {
		commandCtx, cancel = context.WithTimeout(ctx, item.spec.Timeout)
	}
	defer cancel()
	command := exec.CommandContext(commandCtx, item.spec.Name, item.spec.Args...)
	command.Dir = item.spec.Dir
	if item.spec.Env != nil {
		command.Env = append(os.Environ(), item.spec.Env...)
	}
	command.Stdout = item.output
	command.Stderr = item.output
	err := command.Run()
	var afterRunErr error
	if item.spec.AfterRun != nil {
		afterRunErr = item.spec.AfterRun()
		if afterRunErr != nil {
			err = afterRunErr
		}
	}
	if err == nil && afterRunErr == nil && item.spec.AfterSuccess != nil {
		afterRunErr = item.spec.AfterSuccess()
		if afterRunErr != nil {
			err = afterRunErr
		}
	}
	finished := time.Now().UTC()
	finalProgress := manager.emitProgress(item)
	terminal := started
	terminal.UpdatedAt = finished
	terminal.FinishedAt = &finished
	terminal.Output, terminal.OutputTruncated = item.output.snapshot()
	terminal.Progress = finalProgress
	if command.ProcessState != nil && afterRunErr == nil {
		code := command.ProcessState.ExitCode()
		terminal.ExitCode = &code
	}
	switch {
	case ctx.Err() != nil:
		terminal.State = StateCancelled
		terminal.Error = ctx.Err().Error()
	case commandCtx.Err() != nil:
		terminal.State = StateFailed
		terminal.Error = commandCtx.Err().Error()
	case err == nil:
		terminal.State = StateSucceeded
	case errors.Is(err, context.Canceled):
		terminal.State = StateCancelled
		terminal.Error = err.Error()
	default:
		terminal.State = StateFailed
		terminal.Error = err.Error()
	}
	// Completion events are never dropped on cancellation; RunResults stays
	// active until this producer closes the channel after draining commands.
	manager.events <- taskEvent{item: item, snapshot: terminal}
}

// emitProgress serializes parser calls and forwards progress through the result channel.
func (manager *Manager) emitProgress(item *task) *Progress {
	if item.spec.ProgressParser == nil {
		return nil
	}
	item.progressMu.Lock()
	defer item.progressMu.Unlock()
	progress := item.spec.ProgressParser(item.output.string())
	if progress == nil {
		return nil
	}
	copy := *progress
	if progress.Percent != nil {
		percent := *progress.Percent
		copy.Percent = &percent
	}
	select {
	case manager.events <- taskEvent{item: item, progress: &copy, updatedAt: time.Now().UTC(), progressOnly: true}:
	case <-manager.ctx.Done():
	}
	return &copy
}

// emitTerminal publishes a terminal event for work cancelled before execution.
func (manager *Manager) emitTerminal(item *task, state State, err error, exitCode *int) {
	finished := time.Now().UTC()
	item.mu.Lock()
	terminal := item.snapshot
	item.mu.Unlock()
	terminal.State = state
	terminal.UpdatedAt = finished
	terminal.FinishedAt = &finished
	terminal.ExitCode = exitCode
	if err != nil {
		terminal.Error = err.Error()
	}
	terminal.Output, terminal.OutputTruncated = item.output.snapshot()
	// The result consumer remains active during shutdown so this event is accounted for.
	manager.events <- taskEvent{item: item, snapshot: terminal}
}

// runResults consumes task events, updates snapshots, and publishes terminal results.
func (manager *Manager) runResults() {
	defer close(manager.results)
	for event := range manager.events {
		item := event.item
		item.mu.Lock()
		if event.progressOnly {
			item.snapshot.Progress = event.progress
			item.snapshot.UpdatedAt = event.updatedAt
		} else {
			item.snapshot = event.snapshot
		}
		snapshot := snapshotLocked(item)
		completed := !event.progressOnly && snapshot.FinishedAt != nil
		item.mu.Unlock()
		if completed {
			manager.results <- snapshot
		}
	}
}

// snapshotLocked copies the task's state while its mutex is held.
func snapshotLocked(item *task) Snapshot {
	copy := item.snapshot
	copy.Output, copy.OutputTruncated = item.output.snapshot()
	return copy
}

// newID allocates a random 128-bit hexadecimal task identifier.
func newID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(value[:]), nil
}

// cloneSpec copies mutable argument slices before storing a retry attempt.
func cloneSpec(spec CommandSpec) CommandSpec {
	spec.Args = append([]string(nil), spec.Args...)
	spec.Env = append([]string(nil), spec.Env...)
	return spec
}

// boundedOutput keeps the newest bytes written by a command within its configured limit.
type boundedOutput struct {
	mu        sync.Mutex
	data      []byte
	limit     int
	truncated bool
	onWrite   func()
}

// Write retains a bounded tail and emits progress callbacks after releasing its lock.
func (output *boundedOutput) Write(data []byte) (int, error) {
	output.mu.Lock()
	length := len(data)
	if length >= output.limit {
		output.data = append(output.data[:0], data[length-output.limit:]...)
		output.truncated = true
		callback := output.onWrite
		output.mu.Unlock()
		if callback != nil {
			callback()
		}
		return length, nil
	}
	if excess := len(output.data) + length - output.limit; excess > 0 {
		output.data = append(output.data[:0], output.data[excess:]...)
		output.truncated = true
	}
	output.data = append(output.data, data...)
	callback := output.onWrite
	output.mu.Unlock()
	if callback != nil {
		callback()
	}
	return length, nil
}

// snapshot returns a copy of captured output and its truncation indicator.
func (output *boundedOutput) snapshot() (string, bool) {
	output.mu.Lock()
	defer output.mu.Unlock()
	return string(append([]byte(nil), output.data...)), output.truncated
}

// string returns the current captured output as a string.
func (output *boundedOutput) string() string {
	value, _ := output.snapshot()
	return value
}

var _ io.Writer = (*boundedOutput)(nil)
