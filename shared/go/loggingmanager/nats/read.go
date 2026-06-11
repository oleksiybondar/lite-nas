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
		page, err := core.ListEventsPage(input.ToDTO())
		if err != nil {
			return nil, err
		}
		return loggingmanagercontract.ListAlertsResponse{
			Items:      loggingmanagercontract.BuildListAlertItems(page.Items),
			TotalCount: page.TotalCount,
		}, nil
	}
}

func handleGetAlertRPC(core *sharedloggingmanager.Core) sharedmessaging.RPCHandler {
	return func(_ context.Context, envelope sharedmessaging.Envelope) (any, error) {
		input, err := decodePayload[loggingmanagercontract.GetAlertInput](envelope)
		if err != nil {
			return nil, err
		}
		item, found, err := core.GetEvent(input.ToDTO())
		if err != nil {
			return nil, err
		}
		if !found {
			return loggingmanagercontract.GetAlertResponse{}, nil
		}
		flatItem := loggingmanagercontract.BuildListAlertItem(item)
		return loggingmanagercontract.GetAlertResponse{
			Item: &flatItem,
		}, nil
	}
}

func handleGetActiveAlertsRPC(core *sharedloggingmanager.Core) sharedmessaging.RPCHandler {
	return func(_ context.Context, envelope sharedmessaging.Envelope) (any, error) {
		input, err := decodeListInput(envelope)
		if err != nil {
			return nil, err
		}
		page, err := core.ListActiveEventsPage(input.ToDTO())
		if err != nil {
			return nil, err
		}
		return loggingmanagercontract.ListAlertsResponse{
			Items:      loggingmanagercontract.BuildListAlertItems(page.Items),
			TotalCount: page.TotalCount,
		}, nil
	}
}

func handleGetUnacknowledgedActiveAlertsRPC(core *sharedloggingmanager.Core) sharedmessaging.RPCHandler {
	return func(_ context.Context, envelope sharedmessaging.Envelope) (any, error) {
		input, err := decodeListInput(envelope)
		if err != nil {
			return nil, err
		}
		page, err := core.ListActiveUnacknowledgedEventsPage(input.ToDTO())
		if err != nil {
			return nil, err
		}
		return loggingmanagercontract.ListAlertsResponse{
			Items:      loggingmanagercontract.BuildListAlertItems(page.Items),
			TotalCount: page.TotalCount,
		}, nil
	}
}
