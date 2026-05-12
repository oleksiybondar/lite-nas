package nats

import (
	"context"

	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	sharedloggingmanager "lite-nas/shared/loggingmanager"
	sharedmessaging "lite-nas/shared/messaging"
)

func handleAcknowledgeAlertRPC(core *sharedloggingmanager.Core) sharedmessaging.RPCHandler {
	return func(_ context.Context, envelope sharedmessaging.Envelope) (any, error) {
		input, err := decodePayload[loggingmanagercontract.AcknowledgeAlertInput](envelope)
		if err != nil {
			return nil, err
		}
		if err = core.AcknowledgeEvent(input); err != nil {
			return nil, err
		}
		return loggingmanagercontract.OKResponse{OK: true}, nil
	}
}

func handleMuteAlertRPC(core *sharedloggingmanager.Core) sharedmessaging.RPCHandler {
	return func(_ context.Context, envelope sharedmessaging.Envelope) (any, error) {
		input, err := decodePayload[loggingmanagercontract.MuteAlertInput](envelope)
		if err != nil {
			return nil, err
		}
		if err = core.MuteEvent(input); err != nil {
			return nil, err
		}
		return loggingmanagercontract.OKResponse{OK: true}, nil
	}
}
