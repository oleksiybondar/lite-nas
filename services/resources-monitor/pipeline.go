package main

import (
	"context"

	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	systemloggingmanagercontract "lite-nas/shared/contracts/systemloggingmanager"
	"lite-nas/shared/messaging"
)

// snapshotInput is the validated local form of one metrics subscription event.
type snapshotInput struct {
	subject string
	payload map[string]any
}

// alertIntent carries one typed logging-manager effect to the outbound worker.
type alertIntent struct {
	create     *loggingmanagercontract.AlertPayload
	occurrence *loggingmanagercontract.AlertOccurrencePayload
	normalize  *loggingmanagercontract.UpdateAlertStateInput
	result     chan alertResult
}

// alertResult acknowledges delivery before the processor advances local state.
type alertResult struct {
	ok  bool
	err error
}

// channelAlertSender connects the single processor worker to the outbound worker.
type channelAlertSender struct{ effects chan<- alertIntent }

// Create queues an alert creation and waits for its delivery result.
func (sender channelAlertSender) Create(ctx context.Context, payload loggingmanagercontract.AlertPayload) error {
	_, err := sender.send(ctx, alertIntent{create: &payload})
	return err
}

// Occurrence queues an alert occurrence and waits for its delivery result.
func (sender channelAlertSender) Occurrence(ctx context.Context, payload loggingmanagercontract.AlertOccurrencePayload) error {
	_, err := sender.send(ctx, alertIntent{occurrence: &payload})
	return err
}

// Normalize queues a state update and returns the logging manager response.
func (sender channelAlertSender) Normalize(ctx context.Context, request loggingmanagercontract.UpdateAlertStateInput) (bool, error) {
	return sender.send(ctx, alertIntent{normalize: &request})
}

// send applies backpressure and waits for delivery acknowledgement.
func (sender channelAlertSender) send(ctx context.Context, intent alertIntent) (bool, error) {
	intent.result = make(chan alertResult, 1)
	select {
	case sender.effects <- intent:
	case <-ctx.Done():
		return false, ctx.Err()
	}
	// Once queued, wait for the outbound worker so local alert state cannot
	// diverge from a delivery that completes during cancellation.
	result := <-intent.result
	return result.ok, result.err
}

// deliverAlertIntent sends one effect and acknowledges its transport result.
func deliverAlertIntent(ctx context.Context, client messaging.Client, intent alertIntent) {
	result := alertResult{}
	switch {
	case intent.create != nil:
		result.err = client.Publish(ctx, systemloggingmanagercontract.AlertSubject, *intent.create)
	case intent.occurrence != nil:
		result.err = client.Publish(ctx, systemloggingmanagercontract.AlertOccurrenceSubject, *intent.occurrence)
	case intent.normalize != nil:
		var response loggingmanagercontract.OKResponse
		result.err = client.Request(ctx, systemloggingmanagercontract.UpdateAlertStateRPCSubject, *intent.normalize, &response)
		result.ok = response.OK
	}
	intent.result <- result
}
