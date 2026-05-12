package nats

import (
	"context"

	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	sharedloggingmanager "lite-nas/shared/loggingmanager"
	sharedmessaging "lite-nas/shared/messaging"
)

func handleUpdateAlertStateRPC(core *sharedloggingmanager.Core) sharedmessaging.RPCHandler {
	return func(_ context.Context, envelope sharedmessaging.Envelope) (any, error) {
		input, err := decodePayload[loggingmanagercontract.UpdateAlertStateInput](envelope)
		if err != nil {
			return nil, err
		}
		if err = core.SetState(input); err != nil {
			return nil, err
		}
		return loggingmanagercontract.OKResponse{OK: true}, nil
	}
}
