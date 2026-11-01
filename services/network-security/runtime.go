package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/netip"
	"time"

	"github.com/go-playground/validator/v10"

	sharedcontracts "lite-nas/shared/contracts"
	loggingcontract "lite-nas/shared/contracts/loggingmanager"
	networkmetricscontract "lite-nas/shared/contracts/networkmetrics"
	networksecuritycontract "lite-nas/shared/contracts/networksecurity"
	securitycontract "lite-nas/shared/contracts/securityloggingmanager"
	"lite-nas/shared/eventmanager"
	"lite-nas/shared/fileio"
	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/loggingmanager/enum"
	"lite-nas/shared/loggingmanagerauth"
	"lite-nas/shared/messaging"
	"lite-nas/shared/modules"
	"lite-nas/shared/workers"
)

const packagedSecurityConfig = "/etc/lite-nas/network-security.conf"

// alertIntent carries one immutable port scan finding to the outbound worker.
type alertIntent struct {
	ID         string
	Detection  scanDetection
	Window     time.Duration
	SnapshotAt time.Time
}

var scanEventIDs = eventmanager.NewManager(uint64(time.Now().Unix()) % 99_999_999)

// run initializes listeners and the compiled detector before receiving NATS
// messages. Channel owners drain in order: callbacks, coordinator, reporter.
func run(ctx context.Context) error {
	reader, err := fileio.NewFileReader(packagedSecurityConfig)
	if err != nil {
		return err
	}
	cfg, err := loadSecurityConfig(reader)
	if err != nil {
		return err
	}
	infra, err := modules.NewCoreClientAuthInfra(sharedcontracts.ServiceNetworkSecurity, cfg.Logging, cfg.Messaging, cfg.Auth, 24*time.Hour)
	if err != nil {
		return err
	}
	defer infra.Close()
	server, err := messaging.NewServer(cfg.Messaging, infra.Logger, messaging.NewJSONCodec())
	if err != nil {
		return err
	}
	defer server.Close()
	operatorClient := loggingmanagerauth.NewClient(infra.Client, infra.AuthTokenManager, loggingmanagerauth.InjectAlertAccessToken)
	return runWorkers(ctx, cfg, server, operatorClient, infra.Logger)
}

// runWorkers owns the bounded input and effect queues after transport startup.
// The caller supplies the authenticated alert client and closes both clients.
func runWorkers(ctx context.Context, cfg securityConfig, server messaging.Server, operatorClient messaging.Client, log sharedlogger.Logger) error {
	plugin, err := newPortScanPlugin(cfg.Threshold, cfg.MaxPairs, cfg.Window, cfg.Cooldown, time.Now)
	if err != nil {
		return err
	}
	horizontal, err := newPatternPlugin("horizontal_scan", cfg.Horizontal.Threshold, cfg.Horizontal.MaxPairs, cfg.Horizontal.Window, cfg.Horizontal.Cooldown, time.Now)
	if err != nil {
		return err
	}
	repeated, err := newPatternPlugin("repeated_attempts", cfg.Repeated.Threshold, cfg.Repeated.MaxPairs, cfg.Repeated.Window, cfg.Repeated.Cooldown, time.Now)
	if err != nil {
		return err
	}
	plugins := []*portScanPlugin{plugin, horizontal, repeated}
	attempts := make(chan networksecuritycontract.FirewallAttempt, 128)
	snapshots := make(chan networkmetricscontract.SnapshotUpdatedEvent, 8)
	effects := make(chan alertIntent, 64)
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := server.Subscribe(networksecuritycontract.FirewallAttemptSubject, func(handlerCtx context.Context, envelope messaging.Envelope) error {
		attempt, err := decodeAttempt(envelope, validate)
		if err != nil {
			log.Warn("invalid firewall attempt", "error", err)
			return err
		}
		select {
		case attempts <- attempt:
			return nil
		case <-handlerCtx.Done():
			return handlerCtx.Err()
		case <-ctx.Done():
			return ctx.Err()
		}
	}); err != nil {
		return err
	}
	if err := server.Subscribe(networkmetricscontract.SnapshotEventSubject, func(handlerCtx context.Context, envelope messaging.Envelope) error {
		snapshot, err := decodeSnapshot(envelope)
		if err != nil {
			log.Warn("invalid network snapshot", "error", err)
			return err
		}
		select {
		case snapshots <- snapshot:
			return nil
		case <-handlerCtx.Done():
			return handlerCtx.Err()
		case <-ctx.Done():
			return ctx.Err()
		}
	}); err != nil {
		return err
	}
	timer, ticks, err := workers.NewPollingTimerWorker(time.Second, 1)
	if err != nil {
		return err
	}
	coordinatorDone := make(chan struct{})
	reporterDone := make(chan struct{})
	workers.NewBaseWorker(func(workerCtx context.Context) {
		coordinate(workerCtx, plugins, attempts, snapshots, ticks, effects, log)
		close(coordinatorDone)
	}).Start(ctx)
	workers.NewBaseWorker(func(workerCtx context.Context) {
		reportAlerts(workerCtx, effects, operatorClient, log)
		close(reporterDone)
	}).Start(ctx)
	timer.Start(ctx)
	<-ctx.Done()
	if err := server.Drain(); err != nil {
		log.Error("network security NATS drain failed", "error", err)
	}
	close(attempts) // Subscription callbacks have stopped after Drain.
	close(snapshots)
	<-coordinatorDone // Coordinator owns effect production.
	close(effects)
	<-reporterDone
	return ctx.Err()
}

// decodeJSON enforces a bounded, exact transport shape before domain checks.
func decodeJSON(envelope messaging.Envelope, limit int, target any) error {
	if len(envelope.Payload) == 0 || len(envelope.Payload) > limit {
		return errors.New("invalid event payload size")
	}
	decoder := json.NewDecoder(bytes.NewReader(envelope.Payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return errors.New("event contains trailing JSON")
	}
	return nil
}

// decodeAttempt validates every field of an inbound firewall observation.
func decodeAttempt(envelope messaging.Envelope, validate *validator.Validate) (networksecuritycontract.FirewallAttempt, error) {
	var attempt networksecuritycontract.FirewallAttempt
	if err := decodeJSON(envelope, 2048, &attempt); err != nil {
		return attempt, err
	}
	if err := validate.Struct(attempt); err != nil {
		return attempt, err
	}
	source, sourceErr := netip.ParseAddr(attempt.SourceIP)
	destination, destinationErr := netip.ParseAddr(attempt.DestinationIP)
	if sourceErr != nil || destinationErr != nil || source.IsUnspecified() || source.IsMulticast() || destination.IsUnspecified() || destination.IsMulticast() ||
		source.Unmap().String() != attempt.SourceIP || destination.Unmap().String() != attempt.DestinationIP {
		return attempt, errors.New("invalid canonical firewall addresses")
	}
	if attempt.Timestamp.IsZero() || attempt.Timestamp.After(time.Now().Add(5*time.Minute)) {
		return attempt, errors.New("invalid firewall timestamp")
	}
	return attempt, nil
}

// decodeSnapshot bounds the aggregate context retained alongside findings.
func decodeSnapshot(envelope messaging.Envelope) (networkmetricscontract.SnapshotUpdatedEvent, error) {
	var event networkmetricscontract.SnapshotUpdatedEvent
	if err := decodeJSON(envelope, 1024*1024, &event); err != nil {
		return event, err
	}
	if event.Snapshot.Timestamp.IsZero() || event.Snapshot.Timestamp.After(time.Now().Add(5*time.Minute)) ||
		len(event.Snapshot.Interfaces) > 256 || len(event.Snapshot.Sockets.TopLocalPorts) > 256 || len(event.Snapshot.Sockets.TopRemoteIPs) > 256 {
		return event, errors.New("invalid network snapshot")
	}
	return event, nil
}

// coordinate owns the latest snapshot and the port scan plugin. The attempt
// channel is a work queue; snapshots supply context, never a scan verdict.
func coordinate(ctx context.Context, plugins []*portScanPlugin, attempts <-chan networksecuritycontract.FirewallAttempt,
	snapshots <-chan networkmetricscontract.SnapshotUpdatedEvent, ticks <-chan struct{}, effects chan<- alertIntent, log sharedlogger.Logger,
) {
	var snapshotAt time.Time
	shutdown := ctx.Done()
	for attempts != nil || snapshots != nil {
		select {
		case attempt, ok := <-attempts:
			if !ok {
				attempts = nil
				continue
			}
			processCtx := ctx
			if ctx.Err() != nil {
				processCtx = context.Background()
			}
			for _, plugin := range plugins {
				if err := plugin.Advance(processCtx); err != nil {
					log.Error("network detector deadline failed", "kind", plugin.kind, "error", err)
					continue
				}
				detections, err := plugin.Observe(processCtx, attempt)
				if err != nil {
					log.Error("network observation failed", "kind", plugin.kind, "source", attempt.SourceIP, "error", err)
					continue
				}
				for _, detection := range detections {
					contextAt := snapshotAt
					if contextAt.IsZero() || detection.At.Sub(contextAt) > 2*plugin.window || contextAt.After(detection.At.Add(5*time.Minute)) {
						contextAt = time.Time{}
					}
					effects <- alertIntent{ID: scanEventIDs.NextEventID("secnet"), Detection: detection, Window: plugin.window, SnapshotAt: contextAt}
				}
			}
		case snapshot, ok := <-snapshots:
			if !ok {
				snapshots = nil
				continue
			}
			if snapshot.Snapshot.Timestamp.After(snapshotAt) {
				snapshotAt = snapshot.Snapshot.Timestamp
			}
		case <-ticks:
			for _, plugin := range plugins {
				if err := plugin.Advance(ctx); err != nil {
					log.Error("network detector deadline failed", "kind", plugin.kind, "error", err)
				}
			}
		case <-shutdown:
			ticks = nil
			shutdown = nil
		}
	}
}

// reportAlerts retries each alert with its stable ID. Failed shutdown delivery
// is logged with the ID so an operator can account for the pending finding.
func reportAlerts(ctx context.Context, effects <-chan alertIntent, client messaging.Client, log sharedlogger.Logger) {
	for effect := range effects {
		for {
			publishCtx := ctx
			var cancel context.CancelFunc
			if ctx.Err() != nil {
				publishCtx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
			}
			err := publishAlert(publishCtx, client, effect)
			if cancel != nil {
				cancel()
			}
			if err == nil {
				break
			}
			if ctx.Err() != nil {
				log.Error("network alert undelivered at shutdown", "event_id", effect.ID, "error", err)
				break
			}
			log.Error("network alert publication failed; retrying", "event_id", effect.ID, "error", err)
			select {
			case <-ctx.Done():
			case <-time.After(2 * time.Second):
			}
		}
	}
}

// publishAlert uses the existing Security Logging Manager create contract.
func publishAlert(ctx context.Context, client messaging.Client, effect alertIntent) error {
	var message string
	switch effect.Detection.Kind {
	case "horizontal_scan":
		message = fmt.Sprintf("Blocked TCP horizontal scan from %s on port %d: %d distinct destinations in %s", effect.Detection.SourceIP, effect.Detection.DestinationPort, effect.Detection.PortCount, effect.Window)
	case "repeated_attempts":
		message = fmt.Sprintf("Repeated blocked TCP attempts from %s to %s port %d: %d attempts in %s", effect.Detection.SourceIP, effect.Detection.DestinationIP, effect.Detection.DestinationPort, effect.Detection.PortCount, effect.Window)
	default:
		message = fmt.Sprintf("Blocked TCP port scan from %s to %s: %d distinct ports in %s", effect.Detection.SourceIP, effect.Detection.DestinationIP, effect.Detection.PortCount, effect.Window)
	}
	if !effect.SnapshotAt.IsZero() {
		message += "; network snapshot " + effect.SnapshotAt.UTC().Format(time.RFC3339)
	}
	email := false
	return client.Publish(ctx, securitycontract.AlertSubject, loggingcontract.AlertPayload{
		EventID: effect.ID, Category: detectionCategory(effect.Detection.Kind), Severity: enum.SeverityWarning,
		CreatedAt: effect.Detection.At.UTC().Format(time.RFC3339), Source: sharedcontracts.ServiceNetworkSecurity,
		Message: message, Email: &email,
	})
}

// detectionCategory preserves the existing port scan category for older intents.
func detectionCategory(kind string) string {
	if kind == "horizontal_scan" || kind == "repeated_attempts" {
		return kind
	}
	return "port_scan"
}
