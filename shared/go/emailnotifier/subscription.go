package emailnotifier

import (
	"context"

	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	sharedloggingmanager "lite-nas/shared/loggingmanager"
	sharedmessaging "lite-nas/shared/messaging"
)

// NewAlertSubscriptionHandler constructs a validated subscription handler that
// forwards decoded alert payloads into the worker input channel.
func NewAlertSubscriptionHandler(
	validator sharedloggingmanager.InputValidator,
	output chan<- loggingmanagercontract.AlertPayload,
) sharedmessaging.MessageHandler {
	codec := sharedmessaging.NewJSONCodec()

	return func(ctx context.Context, envelope sharedmessaging.Envelope) error {
		var payload loggingmanagercontract.AlertPayload
		if err := codec.Unmarshal(envelope.Payload, &payload); err != nil {
			return err
		}
		if err := validator.Struct(payload); err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case output <- payload:
			return nil
		}
	}
}
