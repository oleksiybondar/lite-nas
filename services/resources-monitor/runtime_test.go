package main

import (
	"context"
	"errors"
	"testing"
	"time"

	serviceconfig "lite-nas/services/resources-monitor/config"
	servicemodules "lite-nas/services/resources-monitor/modules"
	"lite-nas/services/resources-monitor/processor"
	servicerules "lite-nas/services/resources-monitor/rules"
	sharedconfig "lite-nas/shared/config"
	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/messaging"
	sharedmodules "lite-nas/shared/modules"
)

func TestInitialEventCounterUsesNonNegativeSeed(t *testing.T) {
	t.Parallel()

	if got := initialEventCounter(time.Unix(-10, 0)); got != 0 {
		t.Fatalf("initialEventCounter(negative) = %d, want 0", got)
	}

	got := initialEventCounter(time.Unix(123456789, 0))
	if got != uint64(123456789%99_999_999) {
		t.Fatalf("initialEventCounter(positive) = %d", got)
	}
}

func TestRegisterSubscriptionsReturnsServerError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("subscribe failed")
	server := &stubServer{subscribeErr: wantErr}

	err := registerSubscriptions(server, &processor.Processor{})
	if !errors.Is(err, wantErr) {
		t.Fatalf("registerSubscriptions() error = %v, want %v", err, wantErr)
	}
}

func TestRunWithDependenciesReturnsInfraError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("infra failed")
	err := runWithDependencies(
		t.Context(),
		"/etc/lite-nas/resources-monitor.conf",
		"resources-monitor",
		func(string, string) (servicemodules.Infra, error) { return servicemodules.Infra{}, wantErr },
		func([]string) ([]servicerules.Rule, error) { return nil, nil },
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("runWithDependencies() error = %v, want %v", err, wantErr)
	}
}

func TestRunWithDependenciesReturnsRulesError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("rules failed")
	err := runWithDependencies(
		t.Context(),
		"/etc/lite-nas/resources-monitor.conf",
		"resources-monitor",
		func(string, string) (servicemodules.Infra, error) { return buildTestInfra(), nil },
		func([]string) ([]servicerules.Rule, error) { return nil, wantErr },
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("runWithDependencies() error = %v, want %v", err, wantErr)
	}
}

func TestRunWithDependenciesReturnsSubscribeError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("subscribe failed")
	err := runWithDependencies(
		t.Context(),
		"/etc/lite-nas/resources-monitor.conf",
		"resources-monitor",
		func(string, string) (servicemodules.Infra, error) {
			infra := buildTestInfra()
			infra.Server = &stubServer{subscribeErr: wantErr}
			return infra, nil
		},
		func([]string) ([]servicerules.Rule, error) { return []servicerules.Rule{}, nil },
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("runWithDependencies() error = %v, want %v", err, wantErr)
	}
}

func TestRunWithDependenciesReturnsContextCanceledOnGracefulStop(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	err := runWithDependencies(
		ctx,
		"/etc/lite-nas/resources-monitor.conf",
		"resources-monitor",
		func(string, string) (servicemodules.Infra, error) { return buildTestInfra(), nil },
		func([]string) ([]servicerules.Rule, error) { return []servicerules.Rule{}, nil },
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("runWithDependencies() error = %v, want %v", err, context.Canceled)
	}
}

type stubServer struct {
	subscribeErr error
}

func (server *stubServer) Subscribe(string, messaging.MessageHandler) error {
	return server.subscribeErr
}

func (server *stubServer) RegisterRPC(string, messaging.RPCHandler) error {
	return nil
}

func (server *stubServer) Drain() error {
	return nil
}

func (server *stubServer) Close() {}

type stubClient struct{}

func (stubClient) Publish(context.Context, string, any) error { return nil }
func (stubClient) Request(context.Context, string, any, any) error {
	return nil
}
func (stubClient) Drain() error { return nil }
func (stubClient) Close()       {}

func buildTestInfra() servicemodules.Infra {
	return servicemodules.Infra{
		CoreInfra: sharedmodules.CoreInfra{
			Logger: sharedlogger.NewNop(),
			Client: stubClient{},
			Server: &stubServer{},
		},
		Config: serviceconfig.Config{
			Rules: sharedconfig.RulesConfig{Files: []string{"/tmp/rules.json"}},
		},
	}
}
