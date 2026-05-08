package main

import (
	"context"
	"testing"

	"lite-nas/services/system-metrics/modules"
	systemmetricscontract "lite-nas/shared/contracts/systemmetrics"
	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/messaging"
)

type publishCall struct {
	subject string
	payload any
}

type recordingClient struct {
	publishCalls []publishCall
	publishErr   error
	publishHook  func()
}

func (c *recordingClient) Publish(_ context.Context, subject string, payload any) error {
	c.publishCalls = append(c.publishCalls, publishCall{subject: subject, payload: payload})
	if c.publishHook != nil {
		c.publishHook()
	}
	return c.publishErr
}

func (c *recordingClient) Request(context.Context, string, any, any) error {
	return nil
}

func (c *recordingClient) Drain() error {
	return nil
}

func (c *recordingClient) Close() {}

type recordingServer struct {
	rpcHandlers       map[string]messaging.RPCHandler
	registerRPCErrors map[string]error
}

func (s *recordingServer) Subscribe(string, messaging.MessageHandler) error {
	return nil
}

func (s *recordingServer) RegisterRPC(subject string, handler messaging.RPCHandler) error {
	if err, ok := s.registerRPCErrors[subject]; ok {
		return err
	}

	if s.rpcHandlers == nil {
		s.rpcHandlers = make(map[string]messaging.RPCHandler)
	}

	s.rpcHandlers[subject] = handler
	return nil
}

func (s *recordingServer) Drain() error {
	return nil
}

func (s *recordingServer) Close() {}

type recordingLogger struct {
	infos []string
	warns []string
}

func (l *recordingLogger) Debug(string, ...any) {}

func (l *recordingLogger) Info(msg string, _ ...any) {
	l.infos = append(l.infos, msg)
}

func (l *recordingLogger) Warn(msg string, _ ...any) {
	l.warns = append(l.warns, msg)
}

func (l *recordingLogger) Error(string, ...any) {}

func (l *recordingLogger) With(...any) sharedlogger.Logger {
	return l
}

func newSnapshotStore(size int) *modules.SnapshotStore {
	return modules.NewStateModule(size).SnapshotStore
}

func mustRegisterRPCHandlers(t *testing.T, server *recordingServer, store *modules.SnapshotStore) {
	t.Helper()

	if err := registerRPCHandlers(server, store); err != nil {
		t.Fatalf("registerRPCHandlers() error = %v", err)
	}
}

func mustInvokeRPCHandler(t *testing.T, server *recordingServer, subject string) any {
	t.Helper()

	handler, ok := server.rpcHandlers[subject]
	if !ok {
		t.Fatalf("rpcHandlers[%q] is not registered", subject)
	}

	response, err := handler(context.Background(), messaging.Envelope{})
	if err != nil {
		t.Fatalf("rpcHandlers[%q]() error = %v", subject, err)
	}

	return response
}

func mustInvokeSnapshotRPC(t *testing.T, server *recordingServer) systemmetricscontract.GetSnapshotResponse {
	t.Helper()

	response := mustInvokeRPCHandler(t, server, systemmetricscontract.SnapshotRPCSubject)

	snapshotResponse, ok := response.(systemmetricscontract.GetSnapshotResponse)
	if !ok {
		t.Fatalf("stats response type = %T, want systemmetrics.GetSnapshotResponse", response)
	}

	return snapshotResponse
}

func mustInvokeHistoryRPC(t *testing.T, server *recordingServer) systemmetricscontract.GetHistoryResponse {
	t.Helper()

	response := mustInvokeRPCHandler(t, server, systemmetricscontract.HistoryRPCSubject)

	historyResponse, ok := response.(systemmetricscontract.GetHistoryResponse)
	if !ok {
		t.Fatalf("history response type = %T, want systemmetrics.GetHistoryResponse", response)
	}

	return historyResponse
}
