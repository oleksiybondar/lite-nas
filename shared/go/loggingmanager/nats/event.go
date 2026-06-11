package nats

import (
	"context"

	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	sharedloggingmanager "lite-nas/shared/loggingmanager"
	sharedmessaging "lite-nas/shared/messaging"
)

func handleAlert(core *sharedloggingmanager.Core) sharedmessaging.MessageHandler {
	return func(_ context.Context, envelope sharedmessaging.Envelope) error {
		payload, err := decodePayload[loggingmanagercontract.AlertPayload](envelope)
		if err != nil {
			return err
		}
		_, err = core.CreateEvent(payload.ToDTO())
		return err
	}
}

func handleAlertOccurrence(core *sharedloggingmanager.Core) sharedmessaging.MessageHandler {
	return func(_ context.Context, envelope sharedmessaging.Envelope) error {
		payload, err := decodePayload[loggingmanagercontract.AlertOccurrencePayload](envelope)
		if err != nil {
			return err
		}
		return core.AddOccurrence(payload.ToDTO())
	}
}
