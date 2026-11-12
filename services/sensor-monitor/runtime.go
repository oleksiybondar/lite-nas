package main

import (
	"bytes"
	"context"
	"fmt"
	"time"

	sharedconfig "lite-nas/shared/config"
	sharedcontracts "lite-nas/shared/contracts"
	sensorcontract "lite-nas/shared/contracts/sensormonitor"
	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/messaging"
	"lite-nas/shared/metrics"
	"lite-nas/shared/modules"
	"lite-nas/shared/workers"
)

const packagedConfigPath = "/etc/lite-nas/sensor-monitor.conf"

// run assembles and serves one process of sensor collection and NATS access.
func run(ctx context.Context) error {
	infra, cfg, err := modules.LoadCoreInfra(
		packagedConfigPath, sharedcontracts.ServiceSensorMonitor, LoadConfig,
		modules.NewCoreClientServerInfra,
		func(config Config) sharedconfig.MessagingConfig { return config.Messaging },
		func(config Config) sharedconfig.LoggingConfig { return config.Logging },
	)
	if err != nil {
		return err
	}
	defer infra.Close()

	names, err := LoadNames(cfg.Sensors.NamesFile)
	if err != nil {
		return fmt.Errorf("load sensor names: %w", err)
	}
	collectors := []Collector{NewLinuxCollector(names)}
	if cfg.Sensors.FirmwareEnabled {
		if firmware := NewRaspberryPiCollector(names); firmware != nil {
			collectors = append(collectors, firmware)
		}
	}
	store := NewHistoryStore(cfg.Sensors.HistorySize)
	if err := registerRPCHandlers(infra.Server, store); err != nil {
		return err
	}
	timer, ticks, err := workers.NewPollingTimerWorker(cfg.Sensors.PollInterval, 1)
	if err != nil {
		return err
	}
	snapshots := make(chan metrics.SensorSnapshot, 1)
	poller := newSensorPollingWorker(ticks, snapshots, collectors, cfg.Sensors, infra.Logger)
	pollerDone := make(chan struct{})
	publisherDone := make(chan struct{})
	go func() {
		defer close(publisherDone)
		serveSnapshots(ctx, snapshots, store, infra.Client, infra.Logger)
	}()
	go func() {
		defer close(pollerDone)
		for poller.WaitNextPoll(ctx) {
			poller.PollAndSend(ctx)
		}
		close(snapshots) // This worker is the only snapshot producer.
	}()
	timer.Start(ctx)
	infra.Logger.Info("sensor monitor started", "config", packagedConfigPath)
	<-ctx.Done()
	<-pollerDone
	<-publisherDone
	return ctx.Err()
}

// newSensorPollingWorker assembles a partial snapshot from startup collectors.
func newSensorPollingWorker(
	ticks <-chan struct{}, output chan<- metrics.SensorSnapshot,
	collectors []Collector, config SensorConfig, log sharedlogger.Logger,
) workers.PollingWorker[metrics.SensorSnapshot] {
	lastError := make([]string, len(collectors))
	return workers.NewPollingWorker(ticks, output, nil, func(ctx context.Context) (metrics.SensorSnapshot, error) {
		snapshot := emptySnapshot()
		snapshot.Timestamp = time.Now().UTC()
		for index, collector := range collectors {
			section, err := collector.Collect(ctx)
			if err != nil {
				if message := err.Error(); message != lastError[index] {
					log.Warn("sensor collection failed", "collector", index, "error", err)
					lastError[index] = message
				}
			} else if lastError[index] != "" {
				log.Info("sensor collection recovered", "collector", index)
				lastError[index] = ""
			}
			snapshot.Temperatures = append(snapshot.Temperatures, section.Temperatures...)
			snapshot.Voltages = append(snapshot.Voltages, section.Voltages...)
			snapshot.Fans = append(snapshot.Fans, section.Fans...)
			snapshot.Clocks = append(snapshot.Clocks, section.Clocks...)
			snapshot.Throttling = append(snapshot.Throttling, section.Throttling...)
		}
		snapshot.Temperatures = keepNumeric(snapshot.Temperatures, config.TemperatureKeys)
		snapshot.Voltages = keepNumeric(snapshot.Voltages, config.VoltageKeys)
		snapshot.Fans = keepNumeric(snapshot.Fans, config.FanKeys)
		snapshot.Clocks = keepNumeric(snapshot.Clocks, config.ClockKeys)
		snapshot.Throttling = keepStates(snapshot.Throttling, config.ThrottlingKeys)
		return snapshot, nil
	})
}

// keepNumeric selects requested stable keys while preserving discovery order.
func keepNumeric(readings []metrics.SensorNumericReading, keys []string) []metrics.SensorNumericReading {
	if len(keys) == 0 {
		return readings
	}
	selected := readings[:0]
	for _, reading := range readings {
		for _, key := range keys {
			if key == reading.Key {
				selected = append(selected, reading)
				break
			}
		}
	}
	return selected
}

// keepStates selects requested condition keys while preserving discovery order.
func keepStates(readings []metrics.SensorStateReading, keys []string) []metrics.SensorStateReading {
	if len(keys) == 0 {
		return readings
	}
	selected := readings[:0]
	for _, reading := range readings {
		for _, key := range keys {
			if key == reading.Key {
				selected = append(selected, reading)
				break
			}
		}
	}
	return selected
}

// registerRPCHandlers exposes the latest and retained snapshot contracts.
func registerRPCHandlers(server messaging.Server, store *HistoryStore) error {
	if err := server.RegisterRPC(sensorcontract.SnapshotRPCSubject, func(_ context.Context, envelope messaging.Envelope) (any, error) {
		if err := validateEmptyRequest(envelope); err != nil {
			return nil, err
		}
		snapshot, ok := store.Latest()
		return sensorcontract.GetSnapshotResponse{Available: ok, Snapshot: snapshot}, nil
	}); err != nil {
		return err
	}
	return server.RegisterRPC(sensorcontract.HistoryRPCSubject, func(_ context.Context, envelope messaging.Envelope) (any, error) {
		if err := validateEmptyRequest(envelope); err != nil {
			return nil, err
		}
		return sensorcontract.GetHistoryResponse{Items: store.List()}, nil
	})
}

// validateEmptyRequest rejects unexpected RPC payload fields before reading state.
func validateEmptyRequest(envelope messaging.Envelope) error {
	payload := bytes.TrimSpace(envelope.Payload)
	if len(payload) == 0 || bytes.Equal(payload, []byte("{}")) {
		return nil
	}
	return fmt.Errorf("sensor monitor RPC request must be an empty object")
}

// serveSnapshots stores and publishes every completed snapshot, including the
// bounded backlog remaining when collection stops.
func serveSnapshots(
	ctx context.Context, input <-chan metrics.SensorSnapshot, store *HistoryStore,
	client messaging.Client, log sharedlogger.Logger,
) {
	for snapshot := range input {
		store.Add(snapshot)
		publishCtx := ctx
		cancel := func() {}
		if ctx.Err() != nil {
			publishCtx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		}
		err := client.Publish(publishCtx, sensorcontract.SnapshotEventSubject, sensorcontract.SnapshotUpdatedEvent{Snapshot: snapshot})
		cancel()
		if err != nil {
			log.Error("sensor snapshot publication failed", "error", err)
		}
	}
}
