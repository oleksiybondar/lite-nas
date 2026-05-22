package modules

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
)

func TestNewInfraModuleReturnsErrorForMissingConfig(t *testing.T) {
	t.Parallel()

	_, err := NewInfraModule("/non-existent/resources-monitor.conf", "resources-monitor")
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}

func TestLoadConfigWithDefaultAuthServiceAppliesFallback(t *testing.T) {
	t.Parallel()

	configPath := writeConfigFixture(t, "")

	cfg, err := loadConfigWithDefaultAuthService(configPath, "resources-monitor")
	if err != nil {
		t.Fatalf("loadConfigWithDefaultAuthService() error = %v", err)
	}
	if cfg.Auth.ServiceName != "resources-monitor" {
		t.Fatalf("Auth.ServiceName = %q, want resources-monitor", cfg.Auth.ServiceName)
	}
}

func TestLoadConfigWithDefaultAuthServiceKeepsConfiguredName(t *testing.T) {
	t.Parallel()

	configPath := writeConfigFixture(t, "preconfigured-service")

	cfg, err := loadConfigWithDefaultAuthService(configPath, "resources-monitor")
	if err != nil {
		t.Fatalf("loadConfigWithDefaultAuthService() error = %v", err)
	}
	if cfg.Auth.ServiceName != "preconfigured-service" {
		t.Fatalf("Auth.ServiceName = %q, want preconfigured-service", cfg.Auth.ServiceName)
	}
}

func TestAuthTokenClientPublishInjectsAccessToken(t *testing.T) {
	t.Parallel()

	clientStub := &messagingClientStub{}
	tokenManager := &tokenManagerStub{
		tokenValue: "token-1",
		tokenErrs:  []error{nil},
	}
	client := authTokenClient{client: clientStub, tokenManager: tokenManager}

	err := client.Publish(context.Background(), "system-alert", loggingmanagercontract.AlertPayload{
		Category: "disk",
	})
	if err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	if len(clientStub.published) != 1 {
		t.Fatalf("publish count = %d, want 1", len(clientStub.published))
	}
	payload, ok := clientStub.published[0].payload.(loggingmanagercontract.AlertPayload)
	if !ok {
		t.Fatalf("payload type = %T, want AlertPayload", clientStub.published[0].payload)
	}
	if payload.AccessToken != "token-1" {
		t.Fatalf("AccessToken = %q, want token-1", payload.AccessToken)
	}
}

func TestAuthTokenClientRequestInjectsAccessToken(t *testing.T) {
	t.Parallel()

	clientStub := &messagingClientStub{}
	tokenManager := &tokenManagerStub{
		tokenValue: "token-2",
		tokenErrs:  []error{nil},
	}
	client := authTokenClient{client: clientStub, tokenManager: tokenManager}

	var response any
	err := client.Request(context.Background(), "system-logging-manager.updateAlertState", loggingmanagercontract.UpdateAlertStateInput{
		EventID: "evt-1",
		Status:  "active",
	}, &response)
	if err != nil {
		t.Fatalf("Request() error = %v", err)
	}

	if len(clientStub.requests) != 1 {
		t.Fatalf("request count = %d, want 1", len(clientStub.requests))
	}
	requestPayload, ok := clientStub.requests[0].request.(loggingmanagercontract.UpdateAlertStateInput)
	if !ok {
		t.Fatalf("request type = %T, want UpdateAlertStateInput", clientStub.requests[0].request)
	}
	if requestPayload.AccessToken != "token-2" {
		t.Fatalf("AccessToken = %q, want token-2", requestPayload.AccessToken)
	}
}

func TestAuthTokenClientDrainDelegates(t *testing.T) {
	t.Parallel()

	clientStub := &messagingClientStub{}
	client := authTokenClient{client: clientStub, tokenManager: &tokenManagerStub{}}

	if err := client.Drain(); err != nil {
		t.Fatalf("Drain() error = %v", err)
	}
	if clientStub.drainCalls != 1 {
		t.Fatalf("drain calls = %d, want 1", clientStub.drainCalls)
	}
}

func TestAuthTokenClientCloseDelegates(t *testing.T) {
	t.Parallel()

	clientStub := &messagingClientStub{}
	client := authTokenClient{client: clientStub, tokenManager: &tokenManagerStub{}}

	client.Close()
	if clientStub.closeCalls != 1 {
		t.Fatalf("close calls = %d, want 1", clientStub.closeCalls)
	}
}

func TestCurrentTokenReturnsExistingToken(t *testing.T) {
	t.Parallel()

	tokenManager := &tokenManagerStub{
		tokenValue: "token-existing",
		tokenErrs:  []error{nil},
	}
	client := authTokenClient{client: &messagingClientStub{}, tokenManager: tokenManager}

	token, err := client.currentToken(context.Background())
	if err != nil {
		t.Fatalf("currentToken() error = %v", err)
	}
	if token != "token-existing" {
		t.Fatalf("token = %q, want token-existing", token)
	}
}

func TestCurrentTokenFallsBackToRefresh(t *testing.T) {
	t.Parallel()

	tokenManager := &tokenManagerStub{
		tokenValue: "token-refreshed",
		tokenErrs:  []error{errors.New("missing"), nil},
		refreshErr: nil,
	}
	client := authTokenClient{client: &messagingClientStub{}, tokenManager: tokenManager}

	token, err := client.currentToken(context.Background())
	if err != nil {
		t.Fatalf("currentToken() error = %v", err)
	}
	if token != "token-refreshed" {
		t.Fatalf("token = %q, want token-refreshed", token)
	}
	if tokenManager.refreshCalls != 1 {
		t.Fatalf("refresh calls = %d, want 1", tokenManager.refreshCalls)
	}
}

func TestCurrentTokenFallsBackToLogin(t *testing.T) {
	t.Parallel()

	tokenManager := &tokenManagerStub{
		tokenValue: "token-login",
		tokenErrs:  []error{errors.New("missing"), nil},
		refreshErr: errors.New("refresh-failed"),
		loginErr:   nil,
	}
	client := authTokenClient{client: &messagingClientStub{}, tokenManager: tokenManager}

	token, err := client.currentToken(context.Background())
	if err != nil {
		t.Fatalf("currentToken() error = %v", err)
	}
	if token != "token-login" {
		t.Fatalf("token = %q, want token-login", token)
	}
	if tokenManager.loginCalls != 1 {
		t.Fatalf("login calls = %d, want 1", tokenManager.loginCalls)
	}
}

func TestCurrentTokenReturnsLoginError(t *testing.T) {
	t.Parallel()

	tokenManager := &tokenManagerStub{
		tokenErrs:  []error{errors.New("missing")},
		refreshErr: errors.New("refresh-failed"),
		loginErr:   errors.New("login-failed"),
	}
	client := authTokenClient{client: &messagingClientStub{}, tokenManager: tokenManager}

	_, err := client.currentToken(context.Background())
	if !errors.Is(err, tokenManager.loginErr) {
		t.Fatalf("currentToken() error = %v, want %v", err, tokenManager.loginErr)
	}
}

func TestWithAccessTokenReturnsOriginalForUnknownPayload(t *testing.T) {
	t.Parallel()

	input := struct{ Value string }{Value: "x"}
	output := withAccessToken(input, "token-ignored")

	typed, ok := output.(struct{ Value string })
	if !ok {
		t.Fatalf("output type = %T, want anonymous struct", output)
	}
	if typed.Value != "x" {
		t.Fatalf("Value = %q, want x", typed.Value)
	}
}

type publishCall struct {
	subject string
	payload any
}

type requestCall struct {
	subject string
	request any
}

type messagingClientStub struct {
	published  []publishCall
	requests   []requestCall
	publishErr error
	requestErr error
	drainCalls int
	closeCalls int
}

func (stub *messagingClientStub) Publish(_ context.Context, subject string, payload any) error {
	stub.published = append(stub.published, publishCall{subject: subject, payload: payload})
	return stub.publishErr
}

func (stub *messagingClientStub) Request(_ context.Context, subject string, request any, _ any) error {
	stub.requests = append(stub.requests, requestCall{subject: subject, request: request})
	return stub.requestErr
}

func (stub *messagingClientStub) Drain() error {
	stub.drainCalls++
	return nil
}

func (stub *messagingClientStub) Close() {
	stub.closeCalls++
}

type tokenManagerStub struct {
	tokenValue   string
	tokenErrs    []error
	tokenCalls   int
	refreshErr   error
	refreshCalls int
	loginErr     error
	loginCalls   int
}

func (stub *tokenManagerStub) Token() (string, time.Time, error) {
	var err error
	if stub.tokenCalls < len(stub.tokenErrs) {
		err = stub.tokenErrs[stub.tokenCalls]
	}
	stub.tokenCalls++
	if err != nil {
		return "", time.Time{}, err
	}
	return stub.tokenValue, time.Unix(1, 0), nil
}

func (stub *tokenManagerStub) Refresh(context.Context) error {
	stub.refreshCalls++
	return stub.refreshErr
}

func (stub *tokenManagerStub) Login(context.Context) error {
	stub.loginCalls++
	return stub.loginErr
}

func writeConfigFixture(t *testing.T, authServiceName string) string {
	t.Helper()

	configPath := filepath.Join(t.TempDir(), "resources-monitor.conf")
	configData := []byte(
		"[messaging]\n" +
			"url=tls://127.0.0.1:4222\n" +
			"client_name=resources-monitor\n" +
			"ca=/etc/lite-nas/certificates/transport/root-ca.crt\n" +
			"cert=/etc/lite-nas/certificates/transport/lite-nas-resources-monitor/client.crt\n" +
			"key=/etc/lite-nas/certificates/transport/lite-nas-resources-monitor/client.key\n" +
			"timeout=5s\n" +
			"[auth]\n" +
			"ca=/etc/lite-nas/certificates/identities/root-ca.crt\n" +
			"cert=/etc/lite-nas/certificates/identities/lite-nas-resources-monitor/client.crt\n" +
			"key=/etc/lite-nas/certificates/identities/lite-nas-resources-monitor/client.key\n" +
			"service_name=" + authServiceName + "\n" +
			"[rules]\n" +
			"files=/tmp/rules.json\n" +
			"[logging]\n" +
			"level=warn\n" +
			"format=rfc5424\n" +
			"output=file\n" +
			"file_path=/tmp/resources-monitor.log\n",
	)
	if err := os.WriteFile(configPath, configData, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return configPath
}
