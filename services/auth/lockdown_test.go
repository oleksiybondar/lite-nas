package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"lite-nas/services/auth/pamauth"
	authcontract "lite-nas/shared/contracts/auth"
	"lite-nas/shared/messaging"
)

// lockdownClient records state-change events and can simulate transport loss.
type lockdownClient struct {
	events []authcontract.LockdownChangedEvent
	err    error
}

func (c *lockdownClient) Publish(_ context.Context, subject string, payload any) error {
	if c.err != nil {
		return c.err
	}
	if subject == authcontract.LockdownChangedEventSubject {
		c.events = append(c.events, payload.(authcontract.LockdownChangedEvent))
	}
	return nil
}
func (c *lockdownClient) Request(context.Context, string, any, any) error { return nil }
func (c *lockdownClient) Drain() error                                    { return nil }
func (c *lockdownClient) Close()                                          {}

// newLockdownFixture starts with a missing state file in a private test dir.
func newLockdownFixture(t *testing.T) *lockdownStore {
	t.Helper()
	store, err := openLockdownStore(filepath.Join(t.TempDir(), "lockdown.json"))
	if err != nil {
		t.Fatalf("openLockdownStore: %v", err)
	}
	return store
}

// mustSetLockdown applies a valid policy transition for a test fixture.
func mustSetLockdown(t *testing.T, store *lockdownStore, client messaging.Client, active bool) {
	t.Helper()
	_, err := store.set(context.Background(), authcontract.SetLockdownRequest{Lockdown: active, Reason: "security incident", Initiator: "log-analysis"}, client)
	if err != nil {
		t.Fatalf("set lockdown: %v", err)
	}
}

// Requirements: task.md/SVC-001.
func TestLockdownSurvivesRestart(t *testing.T) {
	t.Parallel()
	store := newLockdownFixture(t)
	mustSetLockdown(t, store, &lockdownClient{}, true)
	reloaded, err := openLockdownStore(store.path)
	if err != nil {
		t.Fatalf("reload lockdown: %v", err)
	}
	if !reloaded.active() {
		t.Fatal("lockdown was lost across restart")
	}
}

// Requirements: task.md/SVC-001.
func TestLockdownRejectsMalformedState(t *testing.T) {
	t.Parallel()
	store := newLockdownFixture(t)
	if err := os.WriteFile(store.path, []byte("{"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	if _, err := openLockdownStore(store.path); err == nil {
		t.Fatal("malformed lockdown state accepted")
	}
}

// Requirements: task.md/SVC-001.
func TestLockdownBlocksLogin(t *testing.T) {
	t.Parallel()
	runtimeDeps := authRuntimeFixture(t, pamauth.Result{Code: pamauth.OutcomeAuthenticated, Username: "alice"})
	runtimeDeps.Lockdown = newLockdownFixture(t)
	mustSetLockdown(t, runtimeDeps.Lockdown, &lockdownClient{}, true)
	response, err := handleLoginRPC(runtimeDeps, rpcEnvelope(t, authcontract.LoginRequest{Username: "alice", Password: "password"}))
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if response.Status != authcontract.StatusDenied {
		t.Fatalf("status = %s", response.Status)
	}
}

// Requirements: task.md/SVC-001.
func TestLockdownBlocksRefresh(t *testing.T) {
	t.Parallel()
	runtimeDeps, login := loginRPCFixture(t)
	runtimeDeps.Lockdown = newLockdownFixture(t)
	mustSetLockdown(t, runtimeDeps.Lockdown, &lockdownClient{}, true)
	response, err := handleRefreshRPC(runtimeDeps, rpcEnvelope(t, authcontract.RefreshRequest{RefreshToken: login.RefreshToken}))
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if response.AccessToken != "" {
		t.Fatal("refresh issued access token during lockdown")
	}
}

// Requirements: task.md/SVC-001.
func TestLockdownKeepsServiceAuthenticationAvailable(t *testing.T) {
	t.Parallel()
	runtimeDeps := authRuntimeFixture(t, pamauth.Result{})
	runtimeDeps.Lockdown = newLockdownFixture(t)
	mustSetLockdown(t, runtimeDeps.Lockdown, &lockdownClient{}, true)
	response, err := handleServiceTokenLoginRPC(runtimeDeps, rpcEnvelope(t, authcontract.ServiceTokenLoginRequest{Service: "lite-nas-test"}))
	if err != nil {
		t.Fatalf("service login: %v", err)
	}
	if response.AccessToken == "" {
		t.Fatal("service authentication blocked")
	}
}

// Requirements: task.md/SVC-001.
func TestLockdownRetriesFailedEvent(t *testing.T) {
	t.Parallel()
	store := newLockdownFixture(t)
	client := &lockdownClient{err: errors.New("nats unavailable")}
	_, err := store.set(context.Background(), authcontract.SetLockdownRequest{Lockdown: true, Reason: "incident", Initiator: "log-analysis"}, client)
	if err == nil {
		t.Fatal("publish failure was not reported")
	}
	client.err = nil
	if err := store.replayPending(context.Background(), client); err != nil {
		t.Fatalf("replay: %v", err)
	}
	if len(client.events) != 1 {
		t.Fatalf("published events = %d", len(client.events))
	}
}

// Requirements: task.md/SVC-001.
func TestInvalidLockdownRPCDoesNotChangeState(t *testing.T) {
	t.Parallel()
	runtimeDeps := authRuntimeFixture(t, pamauth.Result{})
	runtimeDeps.Lockdown = newLockdownFixture(t)
	_, err := handleSetLockdownRPC(context.Background(), runtimeDeps, messaging.Envelope{Payload: []byte(`{"lockdown":true,"reason":"","initiator":"test"}`)})
	if err == nil {
		t.Fatal("invalid lockdown request accepted")
	}
	if runtimeDeps.Lockdown.active() {
		t.Fatal("invalid request changed lockdown")
	}
}
