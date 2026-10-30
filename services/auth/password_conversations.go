package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"lite-nas/services/auth/pamauth"
	authcontract "lite-nas/shared/contracts/auth"
	"lite-nas/shared/logger"
	"lite-nas/shared/workers"
)

const (
	passwordConversationLifetime = 5 * time.Minute
	maxPasswordConversations     = 128
)

// passwordFlow owns one PAM transaction and its one-to-one prompt channels.
// The worker owns event production; RPC handlers own answer production. Neither
// channel is closed: cancellation stops both sides without racing a send.
type passwordFlow struct {
	id        string
	login     authcontract.LoginRequest
	onDemand  bool
	expiresAt time.Time
	answers   chan string
	events    chan passwordEvent
	stepMu    sync.Mutex
	sequence  atomic.Int32
	traced    atomic.Bool
	cancel    context.CancelFunc
	ctx       context.Context
}

// passwordEvent carries the next PAM prompt or a terminal result to an RPC.
type passwordEvent struct {
	result pamauth.Result
	prompt *pamauth.Prompt
}

// passwordConversations routes bounded RPC intents to independent PAM workers.
type passwordConversations struct {
	mu          sync.Mutex
	flows       map[string]*passwordFlow
	serviceName string
	ctx         context.Context
	log         logger.Logger
}

// newPasswordConversations binds PAM workers to the auth-service process lifetime.
func newPasswordConversations(ctx context.Context, serviceName string, log logger.Logger) *passwordConversations {
	return &passwordConversations{ctx: ctx, serviceName: serviceName, log: log, flows: make(map[string]*passwordFlow)}
}

// start creates a worker for one login or on-demand password change.
func (m *passwordConversations) start(ctx context.Context, login authcontract.LoginRequest, onDemand bool) (passwordEvent, *passwordFlow, bool) {
	var idBytes [24]byte
	if _, err := rand.Read(idBytes[:]); err != nil {
		return passwordEvent{}, nil, false
	}
	id := base64.RawURLEncoding.EncodeToString(idBytes[:])
	workerCtx, cancel := context.WithTimeout(m.ctx, passwordConversationLifetime)
	flow := &passwordFlow{id: id, login: login, onDemand: onDemand, expiresAt: time.Now().Add(passwordConversationLifetime), answers: make(chan string), events: make(chan passwordEvent), cancel: cancel, ctx: workerCtx}
	m.mu.Lock()
	if len(m.flows) >= maxPasswordConversations {
		m.mu.Unlock()
		cancel()
		return passwordEvent{}, nil, false
	}
	m.flows[id] = flow
	m.mu.Unlock()
	workers.NewBaseWorker(func(ctx context.Context) { m.run(ctx, flow) }).Start(workerCtx)
	if m.log != nil && onDemand {
		m.log.Info("password conversation started", "conversation_id", id, "on_demand", onDemand)
		flow.traced.Store(true)
	}
	select {
	case <-ctx.Done():
		cancel()
		return passwordEvent{}, nil, false
	case <-workerCtx.Done():
		return passwordEvent{}, nil, false
	case event := <-flow.events:
		return event, flow, true
	}
}

// step sends one answer, then waits for exactly one next prompt or outcome.
func (m *passwordConversations) step(ctx context.Context, id string, sequence int32, answer string) (passwordEvent, *passwordFlow, bool) {
	m.mu.Lock()
	flow := m.flows[id]
	m.mu.Unlock()
	if flow == nil {
		return passwordEvent{}, nil, false
	}
	flow.stepMu.Lock()
	defer flow.stepMu.Unlock()
	if sequence != flow.sequence.Load() {
		if m.log != nil {
			m.log.Info("password conversation reply rejected", "conversation_id", id, "reason", "sequence")
		}
		return passwordEvent{}, nil, false
	}
	if !time.Now().Before(flow.expiresAt) {
		if m.log != nil {
			m.log.Info("password conversation reply rejected", "conversation_id", id, "reason", "expired")
		}
		flow.cancel()
		return passwordEvent{}, nil, false
	}
	if m.log != nil && flow.traced.Load() {
		m.log.Info("password conversation reply accepted", "conversation_id", id, "sequence", sequence)
	}
	select {
	case <-ctx.Done():
		return passwordEvent{}, nil, false
	case <-flow.ctx.Done():
		return passwordEvent{}, nil, false
	case flow.answers <- answer:
	}
	if m.log != nil && flow.traced.Load() {
		m.log.Info("password conversation reply delivered to PAM", "conversation_id", id, "sequence", sequence)
	}
	select {
	case <-ctx.Done():
		return passwordEvent{}, nil, false
	case <-flow.ctx.Done():
		return passwordEvent{}, nil, false
	case event := <-flow.events:
		if m.log != nil && flow.traced.Load() {
			m.log.Info("password conversation reply completed", "conversation_id", id, "sequence", sequence, "next_prompt", event.prompt != nil, "status", event.result.Code)
		}
		return event, flow, true
	}
}

// run owns the PAM handle, emits prompts, and waits on the flow's answer channel.
func (m *passwordConversations) run(ctx context.Context, flow *passwordFlow) {
	defer flow.cancel()
	defer func() {
		m.mu.Lock()
		delete(m.flows, flow.id)
		m.mu.Unlock()
	}()
	// PAM owns thread-local state in some modules. Keep this transaction on one OS thread.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	result, _ := pamauth.RunConversation(m.serviceName, pamauth.AuthenticateRequest{Username: flow.login.Username, Password: flow.login.Password}, flow.onDemand, func(prompt pamauth.Prompt) (string, error) {
		flow.sequence.Add(1)
		if m.log != nil {
			if flow.traced.CompareAndSwap(false, true) {
				m.log.Info("password conversation started", "conversation_id", flow.id, "on_demand", flow.onDemand)
			}
			m.log.Info("password conversation prompt", "conversation_id", flow.id, "secret", prompt.Secret)
			for _, message := range prompt.Messages {
				if message.Level == pamauth.MessageLevelError {
					m.log.Info("password conversation retry", "conversation_id", flow.id)
					break
				}
			}
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case flow.events <- passwordEvent{prompt: &prompt}:
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case answer := <-flow.answers:
			return answer, nil
		}
	}, func(stage string) {
		if m.log != nil && flow.traced.Load() {
			m.log.Info("password conversation PAM stage", "conversation_id", flow.id, "stage", stage)
		}
	})
	if m.log != nil && flow.traced.Load() {
		m.log.Info("password conversation ended", "conversation_id", flow.id, "status", result.Code, "expired", ctx.Err() != nil)
	}
	select {
	case <-ctx.Done():
	case flow.events <- passwordEvent{result: result}:
	}
}
