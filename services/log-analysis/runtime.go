package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"lite-nas/services/log-analysis/plugins"
	sharedcontracts "lite-nas/shared/contracts"
	authcontract "lite-nas/shared/contracts/auth"
	loggingcontract "lite-nas/shared/contracts/loggingmanager"
	securitycontract "lite-nas/shared/contracts/securityloggingmanager"
	"lite-nas/shared/fileio"
	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/loggingmanager/enum"
	"lite-nas/shared/loggingmanagerauth"
	"lite-nas/shared/messaging"
	"lite-nas/shared/modules"
	"lite-nas/shared/workers"
)

const packagedAnalysisConfig = "/etc/lite-nas/log-analysis.conf"

// pluginConstructor is a fixed Go constructor selected by startup code only.
type pluginConstructor func(plugins.Config, chan<- plugins.Detection, sharedlogger.Logger) (*plugins.Plugin, error)

var compiledPluginTypes = map[string]pluginConstructor{
	"immediate":       plugins.NewImmediate,
	"threshold":       plugins.NewThreshold,
	"keyed_threshold": plugins.NewKeyedThreshold,
}

// run constructs all configured plugin instances before the rsyslog receiver
// or dispatcher starts. It then waits for the bounded pipeline to drain on stop.
func run(ctx context.Context) error {
	reader, err := fileio.NewFileReader(packagedAnalysisConfig)
	if err != nil {
		return err
	}
	cfg, err := loadAnalysisConfig(reader)
	if err != nil {
		return err
	}
	spec, err := loadRulesFile(cfg.RulesFile)
	if err != nil {
		return err
	}
	senders, err := resolveSenderUIDs(cfg.SenderUsers)
	if err != nil {
		return err
	}
	infra, err := modules.NewCoreClientAuthInfra(sharedcontracts.ServiceLogAnalysis, cfg.Logging, cfg.Messaging, cfg.Auth, 24*time.Hour)
	if err != nil {
		return err
	}
	defer infra.Close()
	effects := make(chan plugins.Detection, 64)
	instances, err := constructInstances(spec.Instances, spec.EventIDPrefix, effects, infra.Logger)
	if err != nil {
		return err
	}
	routes, err := buildRoutes(spec, instances)
	if err != nil {
		return err
	}
	conn, err := activatedDatagramSocket()
	if err != nil {
		return err
	}
	// The receiver is the only producer; the runtime closes records after it stops.
	records := make(chan logRecord, 64)
	receiver, err := newRsyslogReceiver(conn, spec.Sources, senders, records, infra.Logger)
	if err != nil {
		_ = conn.Close()
		return err
	}
	dispatcher := newDispatchWorker(records, routes, instances, infra.Logger)
	reporterDone := make(chan struct{})
	operatorClient := loggingmanagerauth.NewClient(infra.Client, infra.AuthTokenManager, loggingmanagerauth.InjectAlertAccessToken)
	workers.NewBaseWorker(func(workerCtx context.Context) {
		reportDetections(workerCtx, effects, operatorClient, infra.Client, infra.Logger)
		close(reporterDone)
	}).Start(ctx)
	for _, plugin := range instances {
		plugin.Start(ctx)
	}
	dispatcher.Start(ctx)
	receiver.Start(ctx)
	<-ctx.Done()
	<-receiver.Done()
	close(records)      // The receiver producer has stopped.
	<-dispatcher.Done() // Dispatcher has closed each plugin input.
	for _, plugin := range instances {
		<-plugin.Done()
	}
	close(effects) // All plugin producers have stopped.
	<-reporterDone
	return ctx.Err()
}

// constructInstances uses the static registry once at the composition root.
// Distinct IDs remain distinct even when they select the same compiled type.
func constructInstances(specs []instanceSpec, eventIDPrefix string, output chan<- plugins.Detection, log sharedlogger.Logger) (map[string]*plugins.Plugin, error) {
	instances := make(map[string]*plugins.Plugin)
	for _, spec := range specs {
		if !spec.Enabled {
			continue
		}
		constructor := compiledPluginTypes[spec.Type]
		if constructor == nil {
			return nil, fmt.Errorf("unknown plugin type %q", spec.Type)
		}
		cfg, err := spec.pluginConfig(eventIDPrefix)
		if err != nil {
			return nil, fmt.Errorf("plugin instance %s: %w", spec.ID, err)
		}
		plugin, err := constructor(cfg, output, log)
		if err != nil {
			return nil, fmt.Errorf("plugin instance %s: %w", spec.ID, err)
		}
		if instances[spec.ID] != nil {
			return nil, fmt.Errorf("duplicate plugin instance %s", spec.ID)
		}
		instances[spec.ID] = plugin
	}
	return instances, nil
}

// reportDetections retries typed effects in order. Transport failures remain
// visible and block later effects rather than silently dropping observations.
func reportDetections(ctx context.Context, effects <-chan plugins.Detection, alertClient, authClient messaging.Client, log sharedlogger.Logger) {
	for effect := range effects {
		alertSent := false
		for {
			attemptCtx := ctx
			var cancel context.CancelFunc
			if ctx.Err() != nil {
				attemptCtx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
			}
			err := deliverDetection(attemptCtx, effect, &alertSent, alertClient, authClient)
			if cancel != nil {
				cancel()
			}
			if err == nil {
				break
			}
			if ctx.Err() != nil {
				log.Error("log detection undelivered at shutdown", "rule", effect.RuleID, "event_id", effect.EventID, "error", err)
				break
			}
			log.Error("log detection publication failed; retrying", "rule", effect.RuleID, "event_id", effect.EventID, "error", err)
			select {
			case <-ctx.Done():
			case <-time.After(2 * time.Second):
			}
		}
	}
}

// deliverDetection follows the manager lifecycle and invokes the verified auth
// RPC only after an explicitly configured critical create effect succeeds.
func deliverDetection(ctx context.Context, effect plugins.Detection, alertSent *bool, alertClient, authClient messaging.Client) error {
	if !*alertSent {
		switch effect.Operation {
		case "create":
			message := detectionMessage("Security rule triggered", effect)
			email := effect.Email
			err := alertClient.Publish(ctx, securitycontract.AlertSubject, loggingcontract.AlertPayload{
				EventID: effect.EventID, Category: effect.Category, Severity: enum.Severity(effect.Severity),
				CreatedAt: effect.At.UTC().Format(time.RFC3339), Source: sharedcontracts.ServiceLogAnalysis,
				Message: message, Email: &email,
			})
			if err != nil {
				return err
			}
			// Publish is asynchronous: the manager may not have committed the
			// event before the following state RPC. Remember the successful create
			// so retries can wait on resolution without recreating the active row.
			*alertSent = true
		case "occurrence":
			value := "repeated matching observations"
			if err := alertClient.Publish(ctx, securitycontract.AlertOccurrenceSubject, loggingcontract.AlertOccurrencePayload{
				EventID: effect.EventID, Timestamp: effect.At.UTC().Format(time.RFC3339),
				ValueType: enum.ValueTypeText, ValueText: &value,
			}); err != nil {
				return err
			}
			*alertSent = true
		case "normalize":
			message := detectionMessage("Security detection cooldown elapsed or successful login observed", effect)
			response := loggingcontract.OKResponse{}
			err := alertClient.Request(ctx, securitycontract.UpdateAlertStateRPCSubject,
				loggingcontract.UpdateAlertStateInput{EventID: effect.EventID, Status: enum.StatusNormal, Message: &message}, &response)
			if err != nil {
				return err
			}
			if !response.OK {
				return errors.New("security logging manager rejected normalization")
			}
			*alertSent = true
		default:
			return fmt.Errorf("unknown detection operation %q", effect.Operation)
		}
	}
	if effect.Operation == "create" && effect.ResolveImmediately {
		if err := confirmAlertPersisted(ctx, effect.EventID, alertClient); err != nil {
			return err
		}
		message := detectionMessage("Security detection recorded", effect)
		response := loggingcontract.OKResponse{}
		if err := alertClient.Request(ctx, securitycontract.UpdateAlertStateRPCSubject,
			loggingcontract.UpdateAlertStateInput{EventID: effect.EventID, Status: enum.StatusNormal, Message: &message}, &response); err != nil {
			return err
		}
		if !response.OK {
			return errors.New("security logging manager rejected immediate detection resolution")
		}
	}
	if effect.Operation != "create" || !effect.Lockdown {
		return nil
	}
	if effect.Severity != "critical" {
		return errors.New("lockdown effect requires critical severity")
	}
	response := authcontract.SetLockdownResponse{}
	request := authcontract.SetLockdownRequest{
		Lockdown:  true,
		Reason:    "Security rule " + effect.RuleID + " triggered incident " + effect.EventID,
		Initiator: sharedcontracts.ServiceLogAnalysis,
	}
	if err := authClient.Request(ctx, authcontract.SetLockdownRPCSubject, request, &response); err != nil {
		return err
	}
	if !response.Lockdown {
		return errors.New("auth service did not apply lockdown")
	}
	return nil
}

// confirmAlertPersisted waits for the manager's asynchronous alert consumer to
// commit a newly published event before requesting its immediate resolution.
func confirmAlertPersisted(ctx context.Context, eventID string, alertClient messaging.Client) error {
	response := loggingcontract.GetAlertResponse{}
	if err := alertClient.Request(ctx, securitycontract.GetAlertRPCSubject,
		loggingcontract.GetAlertInput{EventID: eventID}, &response); err != nil {
		return err
	}
	if response.Item == nil {
		return errors.New("security logging manager has not persisted the new alert")
	}
	return nil
}

// detectionMessage retains the account or device correlation key in bounded
// alert text, while never including credential material from the source log.
func detectionMessage(prefix string, effect plugins.Detection) string {
	message := prefix + ": " + effect.RuleID
	if channel := authenticationChannel(effect.Service); channel != "" {
		message += " via " + channel
	}
	if effect.Account != "" {
		message += " for account " + effect.Account
	}
	if effect.SourceAddress != "" {
		message += " from " + effect.SourceAddress
	}
	if effect.Device != "" {
		message += " device " + effect.Device
	}
	if len(message) > 256 {
		message = message[:256]
	}
	return message
}

// authenticationChannel maps known PAM services to a human-readable login path.
func authenticationChannel(service string) string {
	switch service {
	case "litenas-auth":
		return "Web/API"
	case "sshd":
		return "SSH/CLI"
	case "login", "su", "su-l":
		return "local CLI"
	case "sudo":
		return "sudo"
	default:
		return ""
	}
}
