package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"lite-nas/shared/authtoken"
	sharedcontracts "lite-nas/shared/contracts"
	contract "lite-nas/shared/contracts/operationalsafety"
	"lite-nas/shared/logger"
	"lite-nas/shared/messaging"
	"lite-nas/shared/modules"
	"lite-nas/shared/servicetoken"
	"lite-nas/shared/workers"
)

// safetyRuntime owns in-memory tasks, startup-compiled rules, and NATS workers.
type safetyRuntime struct {
	rules          map[string]*compiledRule
	order          []string
	tasks          map[string]*safetyTask
	stages         map[string]*stageTask
	pendingNormal  map[string]bool
	events         map[string]*observedEvent
	client         messaging.Client
	verifier       authtoken.Verifier
	tokens         *servicetoken.Manager
	log            logger.Logger
	mu             sync.Mutex
	intents        chan<- effectIntent
	apply          func(context.Context, uint64, uint64) error
	applyManual    func(context.Context, uint64) error
	now            func() time.Time
	publish        func(context.Context, effectIntent) error
	publishOwned   func(context.Context, string, string) (string, error)
	normalizeOwned func(context.Context, string) error
	lastCap        uint64
	ids            *eventIDAllocator
}

// run initializes compiled rules before accepting NATS observations or operator RPCs.
func run(ctx context.Context) error {
	cfg, err := loadServiceConfig(packagedConfig)
	if err != nil {
		return err
	}
	rules, order, err := loadRules(cfg.RulesFile)
	if err != nil {
		return err
	}
	certificate, err := os.ReadFile(cfg.Tokens.VerificationCert)
	if err != nil {
		return err
	}
	key, err := authtoken.ParseEd25519CertificatePublicKeyPEM(certificate)
	if err != nil {
		return err
	}
	verifier, err := authtoken.NewVerifier(authtoken.VerifierOptions{Issuer: cfg.Tokens.Issuer, Audience: cfg.Tokens.Audience, ClockSkew: cfg.Tokens.ClockSkew}, key)
	if err != nil {
		return err
	}
	infra, err := modules.NewCoreClientServerInfra(sharedcontracts.ServiceOperationalSafetyAutomation, cfg.Logging, cfg.Messaging)
	if err != nil {
		return err
	}
	defer infra.Close()
	serviceLogin := cfg.Auth.ServiceLogin
	if serviceLogin == "" {
		serviceLogin = cfg.Auth.ServiceName
	}
	tokens, err := servicetoken.NewManager(infra.Client, servicetoken.Options{Service: serviceLogin})
	if err != nil {
		return err
	}
	ids, err := newEventIDAllocator(eventSequenceDirectory)
	if err != nil {
		return err
	}
	runtime := &safetyRuntime{ids: ids, rules: rules, order: order, tasks: make(map[string]*safetyTask), stages: make(map[string]*stageTask), pendingNormal: make(map[string]bool), events: make(map[string]*observedEvent), client: infra.Client, verifier: verifier, tokens: tokens, log: infra.Logger, apply: applyCPUCap, applyManual: applyManualCPUCap, now: time.Now}
	runtime.publish = runtime.publishIntent
	runtime.publishOwned = runtime.publishStageAlert
	runtime.normalizeOwned = runtime.normalizeStageAlert
	return runtime.runWorkers(ctx, infra.Server)
}

// runWorkers owns bounded observation and effect queues and drains accepted work on shutdown.
func (runtime *safetyRuntime) runWorkers(ctx context.Context, server messaging.Server) error {
	observations := make(chan safetyObservation, 128) // NATS callbacks and timer produce; dispatcher consumes; owner closes after Drain and timer stop.
	effects := make(chan effectIntent, 64)            // Dispatcher and RPC handlers produce; outbound worker consumes; owner closes after server Drain and dispatcher stop.
	runtime.intents = effects
	dispatcherDone := make(chan struct{})
	outboundDone := make(chan struct{})
	workerCtx, stop := context.WithCancel(ctx)
	defer stop()
	workers.NewBaseWorker(func(context.Context) {
		defer close(dispatcherDone)
		for item := range observations {
			if err := runtime.processObservation(workerCtx, item); err != nil {
				runtime.log.Error("operational safety observation failed", "event_id", item.eventID, "error", err)
			}
		}
	}).Start(workerCtx)
	workers.NewBaseWorker(func(context.Context) {
		defer close(outboundDone)
		for intent := range effects {
			runtime.applyIntent(workerCtx, intent)
		}
	}).Start(workerCtx)
	if err := runtime.register(server); err != nil {
		_ = server.Drain()
		close(observations)
		<-dispatcherDone
		close(effects)
		<-outboundDone
		return err
	}
	if err := runtime.subscribe(server, observations); err != nil {
		_ = server.Drain()
		close(observations)
		<-dispatcherDone
		close(effects)
		<-outboundDone
		return err
	}
	timer, ticks, err := workers.NewPollingTimerWorker(5*time.Second, 1)
	if err != nil {
		_ = server.Drain()
		close(observations)
		<-dispatcherDone
		close(effects)
		<-outboundDone
		return err
	}
	timer.Start(workerCtx)
	pollDone := make(chan struct{})
	workers.NewBaseWorker(func(context.Context) {
		defer close(pollDone)
		for {
			select {
			case <-workerCtx.Done():
				return
			case _, ok := <-ticks:
				if !ok {
					return
				}
				runtime.pollAlertStates(workerCtx, observations)
				runtime.advanceCooldowns(workerCtx)
			}
		}
	}).Start(workerCtx)
	<-ctx.Done()
	stop()
	<-pollDone
	if err := server.Drain(); err != nil {
		runtime.log.Error("operational safety NATS drain failed", "error", err)
	}
	close(observations)
	<-dispatcherDone
	close(effects)
	<-outboundDone
	return ctx.Err()
}

// serviceToken refreshes the service credential used for system alert and state RPCs.
func (runtime *safetyRuntime) serviceToken(ctx context.Context) (string, error) {
	token, expires, err := runtime.tokens.Token()
	if err == nil && expires.After(time.Now().Add(time.Minute)) {
		return token, nil
	}
	if err == nil && runtime.tokens.Refresh(ctx) == nil {
		token, _, err = runtime.tokens.Token()
		return token, err
	}
	if err = runtime.tokens.Login(ctx); err != nil {
		return "", err
	}
	token, _, err = runtime.tokens.Token()
	return token, err
}

// taskViews copies active task state for the operator RPC without exposing mutable maps.
func (runtime *safetyRuntime) taskViews() []contract.OperationalSafetyTask {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	views := make([]contract.OperationalSafetyTask, 0, len(runtime.tasks))
	for _, task := range runtime.tasks {
		views = append(views, task.view())
	}
	for _, task := range runtime.stages {
		views = append(views, task.view(runtime.rules[task.ruleID]))
	}
	sort.Slice(views, func(i, j int) bool { return views[i].ID < views[j].ID })
	return views
}

// enqueueEffect bounds a task effect request and honors request cancellation.
func (runtime *safetyRuntime) enqueueEffect(ctx context.Context, intent effectIntent) error {
	if runtime.intents == nil {
		return errors.New("effect worker unavailable")
	}
	select {
	case runtime.intents <- intent:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// mustRule returns a configured rule or a concrete operator error.
func (runtime *safetyRuntime) mustRule(id string) (*compiledRule, error) {
	rule := runtime.rules[id]
	if rule == nil {
		return nil, fmt.Errorf("unknown operational safety rule %q", id)
	}
	return rule, nil
}
