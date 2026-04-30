package main

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"lite-nas/services/auth/modules"
	"lite-nas/services/auth/pamauth"
	"lite-nas/services/auth/sessions"
	"lite-nas/shared/authtoken"
	authcontract "lite-nas/shared/contracts/auth"
	"lite-nas/shared/messaging"
	"lite-nas/shared/testutil/authtokentest"
)

type runtimeAuthenticatorStub struct {
	result pamauth.Result
	err    error
}

func (a runtimeAuthenticatorStub) Authenticate(pamauth.AuthenticateRequest) (pamauth.Result, error) {
	return a.result, a.err
}

func (a runtimeAuthenticatorStub) ChangePassword(pamauth.PasswordChangeRequest) (pamauth.Result, error) {
	return a.result, a.err
}

type runtimeRecordingServer struct {
	handlers map[string]messaging.RPCHandler
	errs     map[string]error
}

func (s *runtimeRecordingServer) Subscribe(string, messaging.MessageHandler) error {
	return nil
}

func (s *runtimeRecordingServer) RegisterRPC(subject string, handler messaging.RPCHandler) error {
	if err := s.errs[subject]; err != nil {
		return err
	}
	if s.handlers == nil {
		s.handlers = make(map[string]messaging.RPCHandler)
	}
	s.handlers[subject] = handler
	return nil
}

func (s *runtimeRecordingServer) Drain() error { return nil }
func (s *runtimeRecordingServer) Close()       {}

func TestHandleLoginRPCIssuesTokens(t *testing.T) {
	t.Parallel()

	runtimeDeps := authRuntimeFixture(t, pamauth.Result{
		Code:     pamauth.OutcomeAuthenticated,
		Username: "testuser",
	})

	response, err := handleLoginRPC(runtimeDeps, rpcEnvelope(t, authcontract.LoginRequest{
		Username:  "testuser",
		Password:  "testpassword",
		UserAgent: "lite-nas-test",
	}))
	if err != nil {
		t.Fatalf("handleLoginRPC() error = %v", err)
	}

	if response.Status != authcontract.StatusAuthenticated {
		t.Fatalf("Status = %q, want %q", response.Status, authcontract.StatusAuthenticated)
	}
	if response.AccessToken == "" || response.RefreshToken == "" {
		t.Fatalf("tokens = (%q, %q), want both set", response.AccessToken, response.RefreshToken)
	}
	if runtimeDeps.RefreshStore.Len() != 1 {
		t.Fatalf("refresh store len = %d, want 1", runtimeDeps.RefreshStore.Len())
	}
}

func TestHandleLoginRPCReturnsDeniedForPAMDenial(t *testing.T) {
	t.Parallel()

	runtimeDeps := authRuntimeFixture(t, pamauth.Result{
		Code:     pamauth.OutcomeDenied,
		Username: "testuser",
	})

	response, err := handleLoginRPC(runtimeDeps, rpcEnvelope(t, authcontract.LoginRequest{
		Username:  "testuser",
		Password:  "wrong",
		UserAgent: "lite-nas-test",
	}))
	if err != nil {
		t.Fatalf("handleLoginRPC() error = %v", err)
	}
	if response.Status != authcontract.StatusDenied {
		t.Fatalf("Status = %q, want %q", response.Status, authcontract.StatusDenied)
	}
}

func TestHandleRefreshRPCRotatesRefreshToken(t *testing.T) {
	t.Parallel()

	runtimeDeps, loginResponse := loginRPCFixture(t)

	refreshResponse, err := handleRefreshRPC(runtimeDeps, rpcEnvelope(t, authcontract.RefreshRequest{
		RefreshToken: loginResponse.RefreshToken,
		UserAgent:    "lite-nas-test",
	}))
	if err != nil {
		t.Fatalf("handleRefreshRPC() error = %v", err)
	}
	if refreshResponse.AccessToken == "" || refreshResponse.RefreshToken == "" {
		t.Fatalf("refresh tokens = (%q, %q), want both set", refreshResponse.AccessToken, refreshResponse.RefreshToken)
	}
}

func TestHandleLogoutRPCRevokesRefreshToken(t *testing.T) {
	t.Parallel()

	runtimeDeps, loginResponse := loginRPCFixture(t)

	logoutResponse, err := handleLogoutRPC(runtimeDeps, rpcEnvelope(t, authcontract.LogoutRequest{
		RefreshToken: loginResponse.RefreshToken,
		UserAgent:    "lite-nas-test",
	}))
	if err != nil {
		t.Fatalf("handleLogoutRPC() error = %v", err)
	}
	if !logoutResponse.LoggedOut {
		t.Fatal("LoggedOut = false, want true")
	}
}

func TestHandleValidateAccessTokenRPCReturnsAuthenticated(t *testing.T) {
	t.Parallel()

	runtimeDeps := authRuntimeFixture(t, pamauth.Result{})
	accessToken, issued := issueAccessToken(runtimeDeps, "testuser", "testuser")
	if !issued {
		t.Fatal("issueAccessToken() failed")
	}

	response, err := handleValidateAccessTokenRPC(runtimeDeps, rpcEnvelope(t, authcontract.ValidateAccessTokenRequest{
		AccessToken: accessToken,
	}))
	if err != nil {
		t.Fatalf("handleValidateAccessTokenRPC() error = %v", err)
	}

	if !response.Valid {
		t.Fatal("Valid = false, want true")
	}
	if response.Username != "testuser" {
		t.Fatalf("Username = %q, want testuser", response.Username)
	}
}

func TestHandleRefreshRPCReturnsEmptyForUnknownRefreshToken(t *testing.T) {
	t.Parallel()

	response, err := handleRefreshRPC(
		authRuntimeFixture(t, pamauth.Result{}),
		rpcEnvelope(t, authcontract.RefreshRequest{RefreshToken: "unknown", UserAgent: "lite-nas-test"}),
	)
	if err != nil {
		t.Fatalf("handleRefreshRPC() error = %v", err)
	}
	if response.AccessToken != "" || response.RefreshToken != "" {
		t.Fatalf("response = %#v, want empty", response)
	}
}

func TestHandleLogoutRPCReturnsFalseForUnknownRefreshToken(t *testing.T) {
	t.Parallel()

	response, err := handleLogoutRPC(
		authRuntimeFixture(t, pamauth.Result{}),
		rpcEnvelope(t, authcontract.LogoutRequest{RefreshToken: "unknown", UserAgent: "lite-nas-test"}),
	)
	if err != nil {
		t.Fatalf("handleLogoutRPC() error = %v", err)
	}
	if response.LoggedOut {
		t.Fatal("LoggedOut = true, want false")
	}
}

func TestHandleValidateAccessTokenRPCReturnsDeniedForInvalidToken(t *testing.T) {
	t.Parallel()

	response, err := handleValidateAccessTokenRPC(
		authRuntimeFixture(t, pamauth.Result{}),
		rpcEnvelope(t, authcontract.ValidateAccessTokenRequest{AccessToken: "invalid"}),
	)
	if err != nil {
		t.Fatalf("handleValidateAccessTokenRPC() error = %v", err)
	}
	if response.Valid || response.Status != authcontract.StatusDenied {
		t.Fatalf("response = %#v, want denied", response)
	}
}

func TestHandleLoginRPCRejectsMalformedPayload(t *testing.T) {
	t.Parallel()

	runtimeDeps := authRuntimeFixture(t, pamauth.Result{})
	badEnvelope := messaging.Envelope{Payload: []byte("{")}

	loginResponse, err := handleLoginRPC(runtimeDeps, badEnvelope)
	if err != nil {
		t.Fatalf("handleLoginRPC() error = %v", err)
	}
	if loginResponse.Status != authcontract.StatusDenied {
		t.Fatalf("login status = %q, want denied", loginResponse.Status)
	}
}

func TestHandleRefreshRPCRejectsMalformedPayload(t *testing.T) {
	t.Parallel()

	runtimeDeps := authRuntimeFixture(t, pamauth.Result{})
	badEnvelope := messaging.Envelope{Payload: []byte("{")}

	refreshResponse, err := handleRefreshRPC(runtimeDeps, badEnvelope)
	if err != nil {
		t.Fatalf("handleRefreshRPC() error = %v", err)
	}
	if refreshResponse.AccessToken != "" || refreshResponse.RefreshToken != "" {
		t.Fatalf("refresh response = %#v, want empty", refreshResponse)
	}
}

func TestHandleLogoutRPCRejectsMalformedPayload(t *testing.T) {
	t.Parallel()

	runtimeDeps := authRuntimeFixture(t, pamauth.Result{})
	badEnvelope := messaging.Envelope{Payload: []byte("{")}

	logoutResponse, err := handleLogoutRPC(runtimeDeps, badEnvelope)
	if err != nil {
		t.Fatalf("handleLogoutRPC() error = %v", err)
	}
	if logoutResponse.LoggedOut {
		t.Fatal("logout LoggedOut = true, want false")
	}
}

func TestAuthStatusFromPAMMapsPasswordChangeRequired(t *testing.T) {
	t.Parallel()

	if got := authStatusFromPAM(pamauth.OutcomePasswordChangeNeeded); got != authcontract.StatusPasswordChangeRequired {
		t.Fatalf("authStatusFromPAM() = %q, want %q", got, authcontract.StatusPasswordChangeRequired)
	}
}

func TestAuthMessagesFromPAMConvertsMessages(t *testing.T) {
	t.Parallel()

	got := authMessagesFromPAM([]pamauth.Message{
		{Level: pamauth.MessageLevelInfo, Text: "hello"},
		{Level: pamauth.MessageLevelWarn, Text: "careful"},
	})
	if len(got) != 2 {
		t.Fatalf("len(messages) = %d, want 2", len(got))
	}
	if got[0].Level != authcontract.MessageLevelInfo || got[1].Level != authcontract.MessageLevelWarn {
		t.Fatalf("message levels = %#v, want info/warn", got)
	}
}

func TestRegisterRPCHandlersRegistersAuthSubjects(t *testing.T) {
	t.Parallel()

	server := &runtimeRecordingServer{}
	if err := registerRPCHandlers(server, authRuntimeFixture(t, pamauth.Result{})); err != nil {
		t.Fatalf("registerRPCHandlers() error = %v", err)
	}

	for _, subject := range []string{
		authcontract.LoginRPCSubject,
		authcontract.RefreshRPCSubject,
		authcontract.LogoutRPCSubject,
		authcontract.ValidateAccessTokenRPCSubject,
	} {
		if server.handlers[subject] == nil {
			t.Fatalf("handler for %q was not registered", subject)
		}
	}
}

func TestRegisterRPCHandlersReturnsRegistrationError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("register failed")
	server := &runtimeRecordingServer{
		errs: map[string]error{authcontract.LoginRPCSubject: expectedErr},
	}
	if err := registerRPCHandlers(server, authRuntimeFixture(t, pamauth.Result{})); !errors.Is(err, expectedErr) {
		t.Fatalf("registerRPCHandlers() error = %v, want %v", err, expectedErr)
	}
}

func authRuntimeFixture(t *testing.T, authResult pamauth.Result) authRuntime {
	t.Helper()

	publicKey, privateKey := authtokentest.MustGenerateEd25519Key(t)
	issuer, err := authtoken.NewIssuer(authtoken.IssuerOptions{
		Issuer:         "lite-nas-auth",
		Audience:       "lite-nas-management-api",
		AccessLifetime: 15 * time.Minute,
	}, privateKey)
	if err != nil {
		t.Fatalf("NewIssuer() error = %v", err)
	}
	verifier, err := authtoken.NewVerifier(authtoken.VerifierOptions{
		Issuer:    "lite-nas-auth",
		Audience:  "lite-nas-management-api",
		ClockSkew: 30 * time.Second,
	}, publicKey)
	if err != nil {
		t.Fatalf("NewVerifier() error = %v", err)
	}

	return authRuntime{
		Auth: modules.Auth{
			ServiceName: "litenas-auth",
			Authenticator: runtimeAuthenticatorStub{
				result: authResult,
			},
		},
		Tokens: authTokenRuntime{
			Issuer:   issuer,
			Verifier: verifier,
		},
		RefreshStore: sessions.NewStore(time.Now, sessions.StoreOptions{}),
	}
}

func rpcEnvelope(t *testing.T, request any) messaging.Envelope {
	t.Helper()

	payload, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	return messaging.Envelope{Payload: payload}
}

func loginRPCFixture(t *testing.T) (authRuntime, authcontract.LoginResponse) {
	t.Helper()

	runtimeDeps := authRuntimeFixture(t, pamauth.Result{
		Code:     pamauth.OutcomeAuthenticated,
		Username: "testuser",
	})
	loginResponse, err := handleLoginRPC(runtimeDeps, rpcEnvelope(t, authcontract.LoginRequest{
		Username:  "testuser",
		Password:  "testpassword",
		UserAgent: "lite-nas-test",
	}))
	if err != nil {
		t.Fatalf("handleLoginRPC() error = %v", err)
	}

	return runtimeDeps, loginResponse
}
