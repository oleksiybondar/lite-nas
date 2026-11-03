package main

import (
	"context"
	"errors"
	"os/exec"
	"testing"
	"time"

	serviceconfig "lite-nas/services/litenas-integrity/config"
	servicemodules "lite-nas/services/litenas-integrity/modules"
	backgroundtasks "lite-nas/shared/backgroundtasks"
	sharedlogger "lite-nas/shared/logger"
	sharedmodules "lite-nas/shared/modules"
	"lite-nas/shared/testutil/messagingtest"
	sharedworkers "lite-nas/shared/workers"
)

// Requirements: litenas-integrity/FR-001, litenas-integrity/FR-002.
// TestPollIntegrityChecksForwardsResult verifies one scheduled command result
// enters the bounded worker channel.
func TestPollIntegrityChecksForwardsResult(t *testing.T) {
	ticks := make(chan struct{}, 1)
	ticks <- struct{}{}
	close(ticks)
	results := make(chan checkResult, 1)
	runner := &integrityRunner{results: []integrityRunnerResult{aideResult("--check", "drift", &exec.ExitError{})}}
	service := newTestIntegrityService(runner, integrityClient{})
	pollIntegrityChecks(t.Context(), ticks, results, service, sharedlogger.NewNop())
	if got := (<-results).Status; got != checkStatusFailure {
		t.Fatalf("poll result status = %v, want failure", got)
	}
}

// Requirements: litenas-integrity/FR-001.
// TestPollIntegrityChecksClosesOnCancellation verifies the producer releases
// ownership of its result channel when stopped.
func TestPollIntegrityChecksClosesOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	results := make(chan checkResult, 1)
	pollIntegrityChecks(ctx, make(chan struct{}), results,
		newTestIntegrityService(&integrityRunner{}, integrityClient{}), sharedlogger.NewNop())
	if _, ok := <-results; ok {
		t.Fatal("poll result channel remained open after cancellation")
	}
}

// Requirements: litenas-integrity/FR-002, litenas-integrity/FR-003.
// TestReportIntegrityChecksAppliesResult verifies the consumer preserves alert
// lifecycle behavior for a queued failed check.
func TestReportIntegrityChecksAppliesResult(t *testing.T) {
	results := make(chan checkResult, 1)
	results <- checkResult{Status: checkStatusFailure, Output: "drift"}
	close(results)
	service := newTestIntegrityService(&integrityRunner{}, integrityClient{
		publish: func(string, any) error { return nil },
	})
	reportIntegrityChecks(t.Context(), results, nil, service)
	if service.active == nil {
		t.Fatal("failed check did not open an incident")
	}
}

// Requirements: litenas-integrity/FR-008.
// TestOnDemandIntegrityTaskWritesToTheResultChannel verifies manager snapshots enter the reporter result pipe.
func TestOnDemandIntegrityTaskWritesToTheResultChannel(t *testing.T) {
	results := make(chan checkResult, 1)
	snapshots := make(chan backgroundtasks.Snapshot, 1)
	exitCode := 0
	snapshots <- backgroundtasks.Snapshot{State: backgroundtasks.StateSucceeded, ExitCode: &exitCode, Output: "clean"}
	close(snapshots)
	forwardIntegrityTaskResults(snapshots, results)
	if got := (<-results).Status; got != checkStatusNormal {
		t.Fatalf("on-demand result status = %v, want normal", got)
	}
}

// Requirements: litenas-integrity/FR-008.
// TestOnDemandIntegrityFindingWritesToTheResultChannel verifies AIDE findings enter the reporter pipeline.
func TestOnDemandIntegrityFindingWritesToTheResultChannel(t *testing.T) {
	results := make(chan checkResult, 1)
	snapshots := make(chan backgroundtasks.Snapshot, 1)
	exitCode := 1
	snapshots <- backgroundtasks.Snapshot{State: backgroundtasks.StateFailed, ExitCode: &exitCode, Output: "drift"}
	close(snapshots)
	forwardIntegrityTaskResults(snapshots, results)
	if got := (<-results).Status; got != checkStatusFailure {
		t.Fatalf("on-demand result status = %v, want failure", got)
	}
}

// Requirements: litenas-integrity/FR-008.
// TestCancelledOnDemandIntegrityTaskWritesToTheResultChannel verifies cancellation remains observable to the reporter.
func TestCancelledOnDemandIntegrityTaskWritesToTheResultChannel(t *testing.T) {
	results := make(chan checkResult, 1)
	snapshots := make(chan backgroundtasks.Snapshot, 1)
	snapshots <- backgroundtasks.Snapshot{State: backgroundtasks.StateCancelled}
	close(snapshots)
	forwardIntegrityTaskResults(snapshots, results)
	if got := (<-results).Status; got != checkStatusCancelled {
		t.Fatalf("on-demand result status = %v, want cancelled", got)
	}
}

// TestOnDemandAIDEExecutionFailureRaisesAnIncident verifies invocation failures use the alert reporter.
// Requirements: litenas-integrity/FR-006, litenas-integrity/FR-008
func TestOnDemandAIDEExecutionFailureRaisesAnIncident(t *testing.T) {
	service := newTestIntegrityService(&integrityRunner{}, integrityClient{
		publish: func(string, any) error { return nil },
	})
	result := integrityTaskResult(backgroundtasks.Snapshot{State: backgroundtasks.StateFailed, Error: "permission denied"})
	service.applyCheckResult(t.Context(), result)
	if service.active == nil {
		t.Fatal("failed on-demand AIDE invocation did not open an incident")
	}
}

// Requirements: litenas-integrity/FR-001.
// TestRunWithDependenciesClosesWorkersOnCancellation verifies runtime wiring
// returns after a canceled service context and its reporter exits.
func TestRunWithDependenciesClosesWorkersOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	err := runWithDependencies(ctx, "/tmp/config", "integrity",
		func(string, string) (servicemodules.Infra, error) { return newPipelineTestInfra(t), nil },
		func(interval time.Duration, capacity int) (sharedworkers.TimerWorker, <-chan struct{}, error) {
			return sharedworkers.NewPollingTimerWorker(interval, capacity)
		})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("runWithDependencies() error = %v, want context canceled", err)
	}
}

// newPipelineTestInfra assembles runtime-owned dependencies without a NATS connection.
func newPipelineTestInfra(t *testing.T) servicemodules.Infra {
	t.Helper()
	_, tokenManager := mustAuthRefreshInfra(t, integrityClient{})
	authTimer, authTicks, err := sharedworkers.NewPollingTimerWorker(time.Hour, 1)
	if err != nil {
		t.Fatalf("NewPollingTimerWorker() error = %v", err)
	}
	return servicemodules.Infra{
		CoreClientAuthInfra: sharedmodules.CoreClientAuthInfra{
			CoreInfra: sharedmodules.CoreInfra{
				Logger: sharedlogger.NewNop(), Client: integrityClient{}, Server: &messagingtest.RecordingServer{},
			},
			AuthTokenManager: tokenManager,
			AuthRefreshTimer: authTimer,
			AuthRefreshTicks: authTicks,
		},
		Config: serviceconfig.Config{AIDE: serviceconfig.AIDEConfig{Schedule: time.Hour}},
	}
}
