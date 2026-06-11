package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	rbaccache "lite-nas/services/rbac/cache"
	sharedcontracts "lite-nas/shared/contracts/rbac"
	"lite-nas/shared/messaging"

	"github.com/go-playground/validator/v10"
)

type rpcTestServer struct {
	handlers    map[string]messaging.RPCHandler
	registerErr error
}

func (server *rpcTestServer) Subscribe(string, messaging.MessageHandler) error { return nil }
func (server *rpcTestServer) RegisterRPC(subject string, handler messaging.RPCHandler) error {
	if server.registerErr != nil {
		return server.registerErr
	}
	if server.handlers == nil {
		server.handlers = make(map[string]messaging.RPCHandler)
	}
	server.handlers[subject] = handler
	return nil
}
func (server *rpcTestServer) UseSubscriptionMiddleware(...messaging.SubscriptionMiddleware) {}
func (server *rpcTestServer) UseRPCMiddleware(...messaging.RPCMiddleware)                   {}
func (server *rpcTestServer) Drain() error                                                  { return nil }
func (server *rpcTestServer) Close()                                                        {}

type scriptedRunner struct {
	results map[string]scriptedResult
}

type scriptedResult struct {
	output string
	err    error
}

func (runner scriptedRunner) Run(_ context.Context, name string, args ...string) ([]byte, error) {
	key := name
	for _, arg := range args {
		key += "\x00" + arg
	}

	result, ok := runner.results[key]
	if !ok {
		return nil, errors.New("unexpected command")
	}

	return []byte(result.output), result.err
}

func TestRegisterRPCHandlersHandlesRead(t *testing.T) {
	t.Parallel()

	server, resolvedPath := mustRegisterHandlersWithFixtures(t)
	assertDecisionRPC(
		t,
		server.handlers[sharedcontracts.CanReadPathRPCSubject],
		sharedcontracts.CheckPathRequest{UID: "1002", Path: resolvedPath},
		true,
	)
}

func TestRegisterRPCHandlersHandlesWriteAndExec(t *testing.T) {
	t.Parallel()

	server, resolvedPath := mustRegisterHandlersWithFixtures(t)
	assertDecisionRPC(
		t,
		server.handlers[sharedcontracts.CanWritePathRPCSubject],
		sharedcontracts.CheckPathRequest{UID: "1002", Path: resolvedPath},
		false,
	)
	assertDecisionRPC(
		t,
		server.handlers[sharedcontracts.CanExecPathRPCSubject],
		sharedcontracts.CheckPathRequest{UID: "1002", Path: resolvedPath},
		false,
	)
}

func TestRegisterRPCHandlersHandlesSudoRolesAndInvalidation(t *testing.T) {
	t.Parallel()

	server, _ := mustRegisterHandlersWithFixtures(t)
	assertDecisionRPC(
		t,
		server.handlers[sharedcontracts.CanSudoExecRPCSubject],
		sharedcontracts.CheckSudoExecRequest{UID: "1002", Command: "/usr/bin/zfs"},
		true,
	)
	assertRolesRPC(t, server.handlers[sharedcontracts.GetSubjectRolesRPCSubject])
	assertInvalidateRPC(t, server.handlers[sharedcontracts.InvalidateCacheRPCSubject])
}

func TestRegisterRPCHandlersReturnsRegisterError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("register failed")
	err := registerRPCHandlers(
		&rpcTestServer{registerErr: wantErr},
		validator.New(validator.WithRequiredStructEnabled()),
		newDecisionService(rbaccache.NewStore(make(chan struct{}, 1)), scriptedRunner{}, time.Minute, time.Minute),
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("registerRPCHandlers() error = %v, want %v", err, wantErr)
	}
}

func TestDecodeRPCRequest(t *testing.T) {
	t.Parallel()

	validatorInstance := validator.New(validator.WithRequiredStructEnabled())
	validPayload, _ := json.Marshal(sharedcontracts.CheckPathRequest{UID: "1002", Path: "/tmp/x"})
	request := sharedcontracts.CheckPathRequest{}
	if !decodeRPCRequest(messaging.Envelope{Payload: validPayload}, &request, validatorInstance) {
		t.Fatalf("decodeRPCRequest() returned false for valid payload")
	}
	if decodeRPCRequest(messaging.Envelope{Payload: []byte("{bad")}, &request, validatorInstance) {
		t.Fatalf("decodeRPCRequest() returned true for invalid payload")
	}
}

func TestForwardInvalidateTicks(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	ticks := make(chan struct{}, 1)
	invalidate := make(chan struct{}, 1)
	done := make(chan struct{})
	go func() {
		forwardInvalidateTicks(ctx, ticks, invalidate)
		close(done)
	}()

	ticks <- struct{}{}
	select {
	case <-invalidate:
	case <-time.After(time.Second):
		t.Fatalf("expected invalidate tick")
	}

	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("forwardInvalidateTicks did not stop")
	}
}

func mustRegisterHandlersWithFixtures(t *testing.T) (*rpcTestServer, string) {
	t.Helper()

	tmpFile, err := os.CreateTemp(t.TempDir(), "rbac-check-*")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	if closeErr := tmpFile.Close(); closeErr != nil {
		t.Fatalf("Close() error = %v", closeErr)
	}
	resolvedPath := tmpFile.Name()

	server := &rpcTestServer{}
	validatorInstance := validator.New(validator.WithRequiredStructEnabled())
	service := newDecisionService(
		rbaccache.NewStore(make(chan struct{}, 1)),
		scriptedRunner{
			results: map[string]scriptedResult{
				"id\x001002": {output: "uid=1002(testuser) gid=100(testgroup) groups=100(testgroup),10(wheel)\n"},
				"getfacl\x00-p\x00" + resolvedPath: {
					output: "# owner: root\n# group: testgroup\nuser::rwx\ngroup::r--\nother::---\n",
				},
				"sudo\x00-n\x00-l\x00-U\x00testuser\x00/usr/bin/zfs": {output: "ok\n"},
				"getent\x00passwd\x001002":                           {output: "testuser:x:1002:100::/home/testuser:/bin/bash\n"},
			},
		},
		time.Hour,
		24*time.Hour,
	)

	if err = registerRPCHandlers(server, validatorInstance, service); err != nil {
		t.Fatalf("registerRPCHandlers() error = %v", err)
	}

	return server, resolvedPath
}

func assertRolesRPC(t *testing.T, handler messaging.RPCHandler) {
	t.Helper()

	rolesPayload, _ := json.Marshal(sharedcontracts.GetSubjectRolesRequest{Username: "1002"})
	rolesResponse, err := handler(t.Context(), messaging.Envelope{Payload: rolesPayload})
	if err != nil {
		t.Fatalf("roles handler error = %v", err)
	}

	roles, ok := rolesResponse.(sharedcontracts.GetSubjectRolesResponse)
	if !ok || roles.UID != "1002" || len(roles.Groups) == 0 {
		t.Fatalf("roles handler response = %#v", rolesResponse)
	}
}

func assertInvalidateRPC(t *testing.T, handler messaging.RPCHandler) {
	t.Helper()

	invalidatePayload, _ := json.Marshal(sharedcontracts.InvalidateCacheRequest{UID: "1002"})
	invalidateResponse, err := handler(t.Context(), messaging.Envelope{Payload: invalidatePayload})
	if err != nil {
		t.Fatalf("invalidate handler error = %v", err)
	}

	invalidate, ok := invalidateResponse.(sharedcontracts.InvalidateCacheResponse)
	if !ok || !invalidate.OK {
		t.Fatalf("invalidate handler response = %#v", invalidateResponse)
	}
}

func assertDecisionRPC(t *testing.T, handler messaging.RPCHandler, request any, want bool) {
	t.Helper()

	payload, _ := json.Marshal(request)
	response, err := handler(t.Context(), messaging.Envelope{Payload: payload})
	if err != nil {
		t.Fatalf("handler error = %v", err)
	}

	decision, ok := response.(sharedcontracts.DecisionResponse)
	if !ok {
		t.Fatalf("unexpected response type: %#v", response)
	}
	if decision.Allowed != want {
		t.Fatalf("decision = %v, want %v", decision.Allowed, want)
	}
}
