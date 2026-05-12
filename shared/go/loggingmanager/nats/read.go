package nats

import (
	"context"

	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	sharedloggingmanager "lite-nas/shared/loggingmanager"
	sharedmessaging "lite-nas/shared/messaging"
)

func handleGetAlertsRPC(core *sharedloggingmanager.Core) sharedmessaging.RPCHandler {
	return func(_ context.Context, envelope sharedmessaging.Envelope) (any, error) {
		input, err := decodeListInput(envelope)
		if err != nil {
			return nil, err
		}
		items, err := core.ListEvents(input)
		if err != nil {
			return nil, err
		}
		return loggingmanagercontract.ListAlertsResponse{Items: items}, nil
	}
}

func handleGetActiveAlertsRPC(core *sharedloggingmanager.Core) sharedmessaging.RPCHandler {
	return func(_ context.Context, envelope sharedmessaging.Envelope) (any, error) {
		input, err := decodeListInput(envelope)
		if err != nil {
			return nil, err
		}
		items, err := core.ListActiveEvents(input)
		if err != nil {
			return nil, err
		}
		return loggingmanagercontract.ListAlertsResponse{Items: items}, nil
	}
}

func handleGetUnacknowledgedActiveAlertsRPC(core *sharedloggingmanager.Core) sharedmessaging.RPCHandler {
	return func(_ context.Context, envelope sharedmessaging.Envelope) (any, error) {
		input, err := decodeListInput(envelope)
		if err != nil {
			return nil, err
		}
		items, err := core.ListActiveUnacknowledgedEvents(input)
		if err != nil {
			return nil, err
		}
		return loggingmanagercontract.ListAlertsResponse{Items: items}, nil
	}
}
