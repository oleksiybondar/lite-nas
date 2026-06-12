package main

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	serviceconfig "lite-nas/services/resources-monitor/config"
	servicemodules "lite-nas/services/resources-monitor/modules"
	"lite-nas/services/resources-monitor/processor"
	servicerules "lite-nas/services/resources-monitor/rules"
	sharedconfig "lite-nas/shared/config"
	authcontract "lite-nas/shared/contracts/auth"
	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/messaging"
	sharedmodules "lite-nas/shared/modules"
	"lite-nas/shared/servicetoken"
	sharedworkers "lite-nas/shared/workers"
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

func TestRegisterSubscriptionsSubscribesToNetworkSystemAndZFSSnapshots(t *testing.T) {
	t.Parallel()

	server := &stubServer{}

	if err := registerSubscriptions(server, &processor.Processor{}); err != nil {
		t.Fatalf("registerSubscriptions() error = %v", err)
	}

	want := []string{
		"network.metrics.events.snapshot",
		"system.metrics.events.stats",
		"zfs.metrics.events.snapshot",
	}
	if !slices.Equal(server.subscribedSubjects, want) {
		t.Fatalf("subscribedSubjects = %v, want %v", server.subscribedSubjects, want)
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

func TestRunWithDependenciesLoadsConfiguredRulesFiles(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	wantFiles := []string{
		"/etc/lite-nas/resources-monitor/rules/system-metrics.json",
		"/etc/lite-nas/resources-monitor/rules/network-metrics.json",
		"/etc/lite-nas/resources-monitor/rules/zfs-metrics.json",
	}

	var gotFiles []string
	err := runWithDependencies(
		ctx,
		"/etc/lite-nas/resources-monitor.conf",
		"resources-monitor",
		func(string, string) (servicemodules.Infra, error) {
			infra := buildTestInfra()
			infra.Config.Rules.Files = append([]string(nil), wantFiles...)
			return infra, nil
		},
		func(files []string) ([]servicerules.Rule, error) {
			gotFiles = append([]string(nil), files...)
			return []servicerules.Rule{}, nil
		},
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("runWithDependencies() error = %v, want %v", err, context.Canceled)
	}

	if !slices.Equal(gotFiles, wantFiles) {
		t.Fatalf("rules loader files = %v, want %v", gotFiles, wantFiles)
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

func TestHandleAuthRefreshTickRefreshesTokenWhenAvailable(t *testing.T) {
	t.Parallel()

	infra := buildAuthTickTestInfra(t, authTickClientStub{
		refreshResponse: authcontract.ServiceTokenRefreshResponse{
			AccessToken:  "AT-refreshed",
			RefreshToken: "RT-refreshed",
			ExpiresAt:    time.Unix(200, 0),
		},
	})

	if err := infra.AuthTokenManager.Login(t.Context()); err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	handleAuthRefreshTick(t.Context(), infra)

	accessToken, _, err := infra.AuthTokenManager.Token()
	if err != nil {
		t.Fatalf("Token() error = %v", err)
	}
	if accessToken != "AT-refreshed" {
		t.Fatalf("access token = %q, want %q", accessToken, "AT-refreshed")
	}
}

func TestHandleAuthRefreshTickFallsBackToLoginOnRefreshError(t *testing.T) {
	t.Parallel()

	infra := buildAuthTickTestInfra(t, authTickClientStub{
		refreshErr: errors.New("refresh failed"),
		loginResponse: authcontract.ServiceTokenLoginResponse{
			AccessToken:  "AT-login",
			RefreshToken: "RT-login",
			ExpiresAt:    time.Unix(300, 0),
		},
	})

	handleAuthRefreshTick(t.Context(), infra)

	accessToken, _, err := infra.AuthTokenManager.Token()
	if err != nil {
		t.Fatalf("Token() error = %v", err)
	}
	if accessToken != "AT-login" {
		t.Fatalf("access token = %q, want %q", accessToken, "AT-login")
	}
}

type stubServer struct {
	subscribeErr       error
	subscribedSubjects []string
}

func (server *stubServer) Subscribe(subject string, _ messaging.MessageHandler) error {
	server.subscribedSubjects = append(server.subscribedSubjects, subject)
	return server.subscribeErr
}

func (server *stubServer) RegisterRPC(string, messaging.RPCHandler) error {
	return nil
}

func (server *stubServer) UseSubscriptionMiddleware(...messaging.SubscriptionMiddleware) {}

func (server *stubServer) UseRPCMiddleware(...messaging.RPCMiddleware) {}

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
	authTokenManager, authRefreshTimer, authRefreshTicks, err := buildAuthInfraComponents(stubClient{})
	if err != nil {
		panic(err)
	}

	return servicemodules.Infra{
		CoreClientAuthInfra: sharedmodules.CoreClientAuthInfra{
			CoreInfra: sharedmodules.CoreInfra{
				Logger: sharedlogger.NewNop(),
				Client: stubClient{},
				Server: &stubServer{},
			},
			AuthTokenManager: authTokenManager,
			AuthRefreshTimer: authRefreshTimer,
			AuthRefreshTicks: authRefreshTicks,
		},
		Config: serviceconfig.Config{
			Rules: sharedconfig.RulesConfig{Files: []string{"/tmp/rules.json"}},
		},
	}
}

type authTickClientStub struct {
	loginResponse   authcontract.ServiceTokenLoginResponse
	refreshResponse authcontract.ServiceTokenRefreshResponse
	loginErr        error
	refreshErr      error
}

func (stub authTickClientStub) Publish(context.Context, string, any) error { return nil }

func (stub authTickClientStub) Request(_ context.Context, subject string, _ any, response any) error {
	switch subject {
	case authcontract.ServiceTokenLoginRPCSubject:
		return stub.handleLoginRequest(response)
	case authcontract.ServiceTokenRefreshRPCSubject:
		return stub.handleRefreshRequest(response)
	default:
		return errors.New("unexpected subject")
	}
}

func (stub authTickClientStub) Drain() error { return nil }

func (stub authTickClientStub) Close() {}

func (stub authTickClientStub) handleLoginRequest(response any) error {
	if stub.loginErr != nil {
		return stub.loginErr
	}
	out, ok := response.(*authcontract.ServiceTokenLoginResponse)
	if !ok {
		return errors.New("unexpected login response type")
	}
	loginResponse := stub.loginResponse
	if loginResponse.AccessToken == "" {
		loginResponse = defaultLoginResponse()
	}
	*out = loginResponse
	return nil
}

func (stub authTickClientStub) handleRefreshRequest(response any) error {
	if stub.refreshErr != nil {
		return stub.refreshErr
	}
	out, ok := response.(*authcontract.ServiceTokenRefreshResponse)
	if !ok {
		return errors.New("unexpected refresh response type")
	}
	refreshResponse := stub.refreshResponse
	if refreshResponse.AccessToken == "" {
		refreshResponse = defaultRefreshResponse()
	}
	*out = refreshResponse
	return nil
}

func defaultLoginResponse() authcontract.ServiceTokenLoginResponse {
	return authcontract.ServiceTokenLoginResponse{
		AccessToken:  "login-placeholder-value",
		RefreshToken: "refresh-placeholder-value",
		ExpiresAt:    time.Unix(100, 0),
	}
}

func defaultRefreshResponse() authcontract.ServiceTokenRefreshResponse {
	return authcontract.ServiceTokenRefreshResponse{
		AccessToken:  "login-placeholder-value-refreshed",
		RefreshToken: "refresh-placeholder-value-refreshed",
		ExpiresAt:    time.Unix(150, 0),
	}
}

func buildAuthTickTestInfra(t *testing.T, client authTickClientStub) servicemodules.Infra {
	t.Helper()

	authTokenManager, authRefreshTimer, authRefreshTicks, err := buildAuthInfraComponents(client)
	if err != nil {
		t.Fatalf("buildAuthInfraComponents() error = %v", err)
	}

	return servicemodules.Infra{
		CoreClientAuthInfra: sharedmodules.CoreClientAuthInfra{
			CoreInfra: sharedmodules.CoreInfra{
				Logger: sharedlogger.NewNop(),
				Client: stubClient{},
				Server: &stubServer{},
			},
			AuthTokenManager: authTokenManager,
			AuthRefreshTimer: authRefreshTimer,
			AuthRefreshTicks: authRefreshTicks,
		},
	}
}

func buildAuthInfraComponents(
	client messaging.Client,
) (*servicetoken.Manager, sharedworkers.TimerWorker, <-chan struct{}, error) {
	authTokenManager, err := servicetoken.NewManager(client, servicetoken.Options{Service: "resources-monitor"})
	if err != nil {
		return nil, sharedworkers.TimerWorker{}, nil, err
	}

	authRefreshTimer, authRefreshTicks, err := sharedworkers.NewPollingTimerWorker(24*time.Hour, 1)
	if err != nil {
		return nil, sharedworkers.TimerWorker{}, nil, err
	}

	return authTokenManager, authRefreshTimer, authRefreshTicks, nil
}
