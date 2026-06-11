package main

import (
	"errors"
	"testing"
	"time"

	rbacmodules "lite-nas/services/rbac/modules"
	sharedworkers "lite-nas/shared/workers"
)

func TestRunReturnsInfraError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("infra failed")
	originalInfraFactory := newInfraModule
	originalTimerFactory := newPollingTimerFunc
	t.Cleanup(func() {
		newInfraModule = originalInfraFactory
		newPollingTimerFunc = originalTimerFactory
	})

	newInfraModule = func(string, string) (rbacmodules.Infra, error) {
		return rbacmodules.Infra{}, wantErr
	}
	newPollingTimerFunc = func(interval time.Duration, bufferSize int) (sharedworkers.TimerWorker, <-chan struct{}, error) {
		t.Fatal("timer factory should not be called when infra bootstrap fails")
		return sharedworkers.TimerWorker{}, nil, nil
	}

	err := run(t.Context())
	if !errors.Is(err, wantErr) {
		t.Fatalf("run() error = %v, want %v", err, wantErr)
	}
}
