package main

import (
	"context"
	"errors"
	"testing"
	"time"

	networkstate "lite-nas/services/network-metrics/state"
	networkmetricscontract "lite-nas/shared/contracts/networkmetrics"
	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/messaging"
	"lite-nas/shared/metrics"
)

func TestRegisterRPCHandlersReturnsUnavailableSnapshotWhenStoreEmpty(t *testing.T) {
	t.Parallel()

	store := networkstate.NewHistoryStore(2)
	server := &recordingServer{}

	if err := registerRPCHandlers(server, store); err != nil {
		t.Fatalf("registerRPCHandlers() error = %v", err)
	}

	handler := server.handlers[networkmetricscontract.SnapshotRPCSubject]
	response, err := handler(context.Background(), messaging.Envelope{})
	if err != nil {
		t.Fatalf("snapshot RPC handler error = %v", err)
	}

	typed, ok := response.(networkmetricscontract.GetSnapshotResponse)
	if !ok {
		t.Fatalf("snapshot RPC response type = %T, want GetSnapshotResponse", response)
	}
	if typed.Available {
		t.Fatal("snapshot response Available = true, want false")
	}
}

func TestRegisterRPCHandlersReturnsServerError(t *testing.T) {
	t.Parallel()

	server := &recordingServer{registerErr: errors.New("register failed")}
	err := registerRPCHandlers(server, networkstate.NewHistoryStore(1))
	if !errors.Is(err, server.registerErr) {
		t.Fatalf("registerRPCHandlers() error = %v, want %v", err, server.registerErr)
	}
}

func TestRegisterRPCHandlersReturnsLatestSnapshotAndHistory(t *testing.T) {
	t.Parallel()

	store := networkstate.NewHistoryStore(2)
	snapshot := metrics.NetworkMetricsSnapshot{}
	store.Add(snapshot)
	server := &recordingServer{}

	if err := registerRPCHandlers(server, store); err != nil {
		t.Fatalf("registerRPCHandlers() error = %v", err)
	}

	currentResponse, err := server.handlers[networkmetricscontract.SnapshotRPCSubject](context.Background(), messaging.Envelope{})
	if err != nil {
		t.Fatalf("snapshot RPC handler error = %v", err)
	}
	historyResponse, err := server.handlers[networkmetricscontract.HistoryRPCSubject](context.Background(), messaging.Envelope{})
	if err != nil {
		t.Fatalf("history RPC handler error = %v", err)
	}

	current := currentResponse.(networkmetricscontract.GetSnapshotResponse)
	if !current.Available {
		t.Fatal("snapshot response Available = false, want true")
	}
	history := historyResponse.(networkmetricscontract.GetHistoryResponse)
	if len(history.Items) != 1 {
		t.Fatalf("history length = %d, want 1", len(history.Items))
	}
}

func TestHandlePollErrorLogsOnlyWhenChannelOpen(t *testing.T) {
	t.Parallel()

	log := &recordingLogger{}
	handlePollError(log, errors.New("boom"), true)
	handlePollError(log, errors.New("ignored"), false)

	if len(log.errors) != 1 {
		t.Fatalf("Error logs = %d, want 1", len(log.errors))
	}
}

func TestHandleSnapshotStoresAndPublishesSnapshot(t *testing.T) {
	t.Parallel()

	store := networkstate.NewHistoryStore(2)
	client := &recordingClient{}
	log := &recordingLogger{}
	snapshot := metrics.NetworkMetricsSnapshot{}

	stopped := handleSnapshot(context.Background(), store, client, log, snapshot, true)
	if stopped {
		t.Fatal("handleSnapshot() stopped = true, want false")
	}

	if client.publishSubject != networkmetricscontract.SnapshotEventSubject {
		t.Fatalf("Publish() subject = %q, want %q", client.publishSubject, networkmetricscontract.SnapshotEventSubject)
	}
	if _, ok := store.Latest(); !ok {
		t.Fatal("Latest() ok = false, want stored snapshot")
	}
}

func TestHandleSnapshotReturnsTrueWhenInputChannelClosed(t *testing.T) {
	t.Parallel()

	stopped := handleSnapshot(context.Background(), networkstate.NewHistoryStore(1), &recordingClient{}, &recordingLogger{}, metrics.NetworkMetricsSnapshot{}, false)
	if !stopped {
		t.Fatal("handleSnapshot() stopped = false, want true")
	}
}

func TestHandleSnapshotLogsPublishError(t *testing.T) {
	t.Parallel()

	log := &recordingLogger{}
	client := &recordingClient{publishErr: errors.New("publish failed")}

	handleSnapshot(context.Background(), networkstate.NewHistoryStore(1), client, log, metrics.NetworkMetricsSnapshot{}, true)

	if len(log.errors) != 1 {
		t.Fatalf("Error logs = %d, want 1", len(log.errors))
	}
}

func TestHandleShutdownReturnsContextError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := handleShutdown(ctx, &recordingLogger{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("handleShutdown() error = %v, want context.Canceled", err)
	}
}

func TestServeSnapshotsReturnsOnContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := serveSnapshots(ctx, make(chan metrics.NetworkMetricsSnapshot), make(chan error), networkstate.NewHistoryStore(1), &recordingClient{}, &recordingLogger{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("serveSnapshots() error = %v, want context.Canceled", err)
	}
}

func TestServeSnapshotsProcessesPollErrorsAndSnapshots(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	input := make(chan metrics.NetworkMetricsSnapshot, 1)
	pollErrors := make(chan error, 1)
	store := networkstate.NewHistoryStore(1)
	client := &recordingClient{published: make(chan struct{}, 1)}
	log := &recordingLogger{}
	snapshot := metrics.NetworkMetricsSnapshot{}

	pollErrors <- errors.New("poll failed")
	input <- snapshot

	done := make(chan error, 1)
	go func() {
		done <- serveSnapshots(ctx, input, pollErrors, store, client, log)
	}()

	waitForPublishedSnapshot(t, client)
	cancel()

	err := <-done
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("serveSnapshots() error = %v, want context.Canceled", err)
	}
	if len(log.errors) == 0 {
		t.Fatal("expected poll error to be logged")
	}
}

type recordingServer struct {
	handlers    map[string]messaging.RPCHandler
	registerErr error
}

func (s *recordingServer) Subscribe(string, messaging.MessageHandler) error { return nil }

func (s *recordingServer) RegisterRPC(subject string, handler messaging.RPCHandler) error {
	if s.registerErr != nil {
		return s.registerErr
	}
	if s.handlers == nil {
		s.handlers = make(map[string]messaging.RPCHandler)
	}
	s.handlers[subject] = handler
	return nil
}

func (s *recordingServer) UseSubscriptionMiddleware(...messaging.SubscriptionMiddleware) {}
func (s *recordingServer) UseRPCMiddleware(...messaging.RPCMiddleware)                   {}
func (s *recordingServer) Drain() error                                                  { return nil }
func (s *recordingServer) Close()                                                        {}

type recordingClient struct {
	publishSubject string
	publishPayload any
	publishErr     error
	published      chan struct{}
}

func (c *recordingClient) Publish(_ context.Context, subject string, payload any) error {
	c.publishSubject = subject
	c.publishPayload = payload
	if c.published != nil {
		select {
		case c.published <- struct{}{}:
		default:
		}
	}
	return c.publishErr
}

func (c *recordingClient) Request(context.Context, string, any, any) error { return nil }
func (c *recordingClient) Drain() error                                    { return nil }
func (c *recordingClient) Close()                                          {}

type recordingLogger struct {
	errors []string
	infos  []string
}

func (l *recordingLogger) Debug(string, ...any) {}
func (l *recordingLogger) Warn(string, ...any)  {}

func (l *recordingLogger) Info(msg string, _ ...any) {
	l.infos = append(l.infos, msg)
}

func (l *recordingLogger) Error(msg string, _ ...any) {
	l.errors = append(l.errors, msg)
}

func (l *recordingLogger) With(...any) sharedlogger.Logger { return l }

func waitForPublishedSnapshot(t *testing.T, client *recordingClient) {
	t.Helper()

	if client.published == nil {
		t.Fatal("recording client has no publish notification channel")
	}
	select {
	case <-client.published:
		return
	case <-time.After(time.Second):
		t.Fatal("snapshot publish was not observed")
	}
}
