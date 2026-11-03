package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"lite-nas/services/security-automation/detection"
	"lite-nas/shared/authtoken"
	sharedcontracts "lite-nas/shared/contracts"
	"lite-nas/shared/logger"
	"lite-nas/shared/messaging"
	"lite-nas/shared/modules"
	"lite-nas/shared/workers"
)

// automationRuntime owns operator-visible rule state and the validated NATS boundaries.
type automationRuntime struct {
	rules              map[string]*compiledAutomationRule
	order              []string
	client             messaging.Client
	verifier           authtoken.Verifier
	log                logger.Logger
	mutex              sync.RWMutex
	incidents          *incidentStore
	runUFW             func(context.Context, string, string, string, string, string) error
	runWebRevocation   func(context.Context, string, string) (int, error)
	terminateSession   sessionTermination
	usb                usbEffects
	physicalInterfaces func() ([]string, error)
	operatorActions    chan effectCommand
	audit              incidentAudit
	report             correlationAlert
}

// compiledFSMTypes links every supported detector into the process regardless of configuration.
var compiledFSMTypes = map[detection.PluginName]func(detection.Config, func() time.Time) (*detection.Plugin, error){
	detection.Sequence:          detection.New,
	detection.CountWithinWindow: detection.New,
}

// constructAutomationInstances creates only configured enabled FSMs before listeners start.
func constructAutomationInstances(rules map[string]*compiledAutomationRule, now func() time.Time) error {
	for id, rule := range rules {
		if !*rule.spec.Enabled {
			continue
		}
		constructor := compiledFSMTypes[rule.spec.FSM.Type]
		if constructor == nil {
			return fmt.Errorf("unknown FSM plugin %q for rule %q", rule.spec.FSM.Type, id)
		}
		plugin, err := constructor(rule.detectorConfig, now)
		if err != nil {
			return fmt.Errorf("rule %s: %w", id, err)
		}
		rule.plugin = plugin
	}
	return nil
}

// run compiles configured instances before registering subscriptions and drains every accepted observation on stop.
func run(ctx context.Context) error {
	cfg, err := loadAutomationConfig(packagedAutomationConfig)
	if err != nil {
		return err
	}
	rules, err := loadAutomationRules(cfg.RulesFile)
	if err != nil {
		return err
	}
	if err := constructAutomationInstances(rules, time.Now); err != nil {
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
	infra, err := modules.NewCoreClientServerInfra(sharedcontracts.ServiceSecurityAutomation, cfg.Logging, cfg.Messaging)
	if err != nil {
		return err
	}
	defer infra.Close()
	incidents, err := openIncidentStore(incidentStateDirectory)
	if err != nil {
		return err
	}
	serviceLogin := cfg.Auth.ServiceLogin
	if serviceLogin == "" {
		serviceLogin = cfg.Auth.ServiceName
	}
	audit, err := newIncidentAudit(infra.Client, serviceLogin)
	if err != nil {
		return err
	}
	report, err := newCorrelationAlert(infra.Client, serviceLogin)
	if err != nil {
		return err
	}
	terminateSession, err := newSessionTermination(infra.Client, serviceLogin)
	if err != nil {
		return err
	}
	usb, err := newUSBEffects(infra.Client, serviceLogin)
	if err != nil {
		return err
	}
	return runAutomationWorkers(ctx, infra.Server, infra.Client, verifier, infra.Logger, rules, incidents, audit, report, terminateSession, usb)
}

// runAutomationWorkers owns bounded alert, observation, tick, and intent channels.
// NATS owns alert producers until Drain; the listener owns observations and ticks;
// the dispatcher owns intents. Each owner closes its outputs only after its input stops.
func runAutomationWorkers(ctx context.Context, server messaging.Server, client messaging.Client,
	verifier authtoken.Verifier, log logger.Logger, rules map[string]*compiledAutomationRule, incidents *incidentStore, audit incidentAudit, report correlationAlert, terminateSession sessionTermination, usb usbEffects,
) error {
	workerCtx, stopWorkers := context.WithCancel(ctx)
	defer stopWorkers()
	runtime := &automationRuntime{
		rules: rules, client: client, verifier: verifier, log: log, incidents: incidents,
		runUFW: invokePrivilegedUFW, physicalInterfaces: discoverPhysicalIngressInterfaces, audit: audit, report: report, terminateSession: terminateSession, usb: usb,
	}
	runtime.runWebRevocation = runtime.revokeMatchedWebSessions
	var effectsDone <-chan struct{}
	if incidents != nil {
		runtime.operatorActions = make(chan effectCommand, 32)
		effectsDone = startOperatorEffectsWorker(workerCtx, runtime)
	}
	stopEffects := func() {
		if runtime.operatorActions != nil {
			close(runtime.operatorActions)
			<-effectsDone
		}
	}
	plugins := make(map[string]*detection.Plugin)
	for id, rule := range rules {
		runtime.order = append(runtime.order, id)
		if rule.plugin != nil {
			plugins[id] = rule.plugin
		}
	}
	sort.Strings(runtime.order)
	alerts := make(chan alertObservation, 128)
	timer, timerTicks, err := workers.NewPollingTimerWorker(time.Second, 1)
	if err != nil {
		stopEffects()
		return err
	}
	observations, ticks, listenerDone := startAutomationListener(workerCtx, alerts, timerTicks, rules, log)
	intents, dispatcherDone, err := startDispatcher(workerCtx, observations, ticks, plugins, 64, func(err error) { log.Error("automation dispatcher error", "error", err) })
	if err != nil {
		close(alerts)
		<-listenerDone
		stopEffects()
		return err
	}
	reporterDone := make(chan struct{})
	workers.NewBaseWorker(func(context.Context) {
		defer close(reporterDone)
		for intent := range intents {
			rule := runtime.rules[intent.RuleID]
			if rule.spec.Action.Kind == "alert" {
				if runtime.report == nil {
					log.Error("automation correlation publisher unavailable", "rule", intent.RuleID, "alerts", intent.Match.AlertIDs)
					continue
				}
				for {
					if err := runtime.report(workerCtx, rule.spec, intent.Match); err != nil {
						if workerCtx.Err() != nil {
							log.Error("automation correlation alert undelivered at shutdown", "rule", intent.RuleID, "alerts", intent.Match.AlertIDs, "error", err)
							break
						}
						log.Error("automation correlation alert publication failed; retrying", "rule", intent.RuleID, "alerts", intent.Match.AlertIDs, "error", err)
						select {
						case <-workerCtx.Done():
						case <-time.After(2 * time.Second):
						}
						continue
					}
					runtime.setRuleOutcome(intent.RuleID, fmt.Sprintf("correlation alert submitted for %s", intent.Match.Key))
					break
				}
				continue
			}
			switch rule.spec.Action.Name {
			case "revoke_web_sessions":
				if err := runtime.queueWebRevocation(intent); err != nil {
					log.Error("automation web-session revocation not queued", "rule", intent.RuleID, "alerts", intent.Match.AlertIDs, "error", err)
				}
			case "block_source_ip":
				// Only the verified Web/API rule establishes direct HTTPS ingress.
				if intent.RuleID != "web_https_source_block" {
					runtime.setRuleOutcome(intent.RuleID, fmt.Sprintf("detected %s; automatic action not enrolled", intent.Match.Key))
					log.Warn("automation match requires enrollment before effect", "rule", intent.RuleID, "target", intent.Match.Key, "alerts", intent.Match.AlertIDs)
					continue
				}
				if err := runtime.queueAutomaticUFW(intent); err != nil {
					log.Error("automation direct HTTPS block not queued", "rule", intent.RuleID, "alerts", intent.Match.AlertIDs, "error", err)
				}
			case "block_scan_source_ip":
				if err := runtime.queueAutomaticScanUFW(intent); err != nil {
					log.Error("automation scan-source block not queued", "rule", intent.RuleID, "alerts", intent.Match.AlertIDs, "error", err)
				}
			case "login_terminate_session":
				if err := runtime.queueSessionTermination(intent); err != nil {
					log.Error("automation login session termination not queued", "rule", intent.RuleID, "alerts", intent.Match.AlertIDs, "error", err)
				}
			case "usb_block_device":
				if err := runtime.queueUSBBlock(intent); err != nil {
					log.Error("automation USB block not queued", "rule", intent.RuleID, "alerts", intent.Match.AlertIDs, "error", err)
				}
			case "":
				runtime.setRuleOutcome(intent.RuleID, fmt.Sprintf("alert-only match for %s", intent.Match.Key))
				log.Info("automation alert-only match", "rule", intent.RuleID, "target", intent.Match.Key, "alerts", intent.Match.AlertIDs)
			}
		}
	}).Start(workerCtx)
	timer.Start(workerCtx)
	if err := runtime.register(server); err != nil {
		_ = server.Drain()
		close(alerts)
		<-listenerDone
		<-dispatcherDone
		<-reporterDone
		stopEffects()
		return err
	}
	if err := subscribeAlertCreations(workerCtx, server, alerts); err != nil {
		_ = server.Drain()
		close(alerts)
		<-listenerDone
		<-dispatcherDone
		<-reporterDone
		stopEffects()
		return err
	}
	log.Info("security automation service started", "rules", len(rules), "enabled_instances", len(plugins))
	<-ctx.Done()
	if err := server.Drain(); err != nil {
		log.Error("automation NATS drain failed", "error", err)
	}
	close(alerts)
	<-listenerDone
	<-dispatcherDone
	<-reporterDone
	stopEffects() // Every producer of effect commands has stopped.
	return ctx.Err()
}
