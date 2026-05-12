package nats

import (
	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	sharedmessaging "lite-nas/shared/messaging"
)

func decodePayload[T any](envelope sharedmessaging.Envelope) (T, error) {
	var payload T
	if err := sharedmessaging.NewJSONCodec().Unmarshal(envelope.Payload, &payload); err != nil {
		return payload, err
	}
	return payload, nil
}

func decodeListInput(envelope sharedmessaging.Envelope) (loggingmanagercontract.ListAlertsInput, error) {
	input, err := decodePayload[loggingmanagercontract.ListAlertsInput](envelope)
	if err != nil {
		return loggingmanagercontract.ListAlertsInput{}, err
	}
	if input.Page == 0 {
		input.Page = 1
	}
	return input, nil
}
