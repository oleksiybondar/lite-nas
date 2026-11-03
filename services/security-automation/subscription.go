package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"

	sharedcontracts "lite-nas/shared/contracts"
	alertcontract "lite-nas/shared/contracts/loggingmanager"
	securitycontract "lite-nas/shared/contracts/securityloggingmanager"
	systemcontract "lite-nas/shared/contracts/systemloggingmanager"
	"lite-nas/shared/loggingmanager"
	"lite-nas/shared/messaging"
)

// alertObservation is one validated create event before rule-specific matching.
type alertObservation struct {
	Kind  string
	Alert alertcontract.AlertPayload
}

// subscribeAlertCreations registers only create events. Occurrence and state
// payloads have different contracts and do not advance detection FSMs.
// The caller owns output and closes it only after server.Drain returns.
func subscribeAlertCreations(ctx context.Context, server messaging.Server, output chan<- alertObservation) error {
	if server == nil || output == nil {
		return errors.New("alert subscription requires server and output")
	}
	validate, err := loggingmanager.NewInputValidator()
	if err != nil {
		return err
	}
	for _, source := range []struct{ subject, kind string }{
		{securitycontract.AlertSubject, "security"}, {systemcontract.AlertSubject, "system"},
	} {
		kind := source.kind
		err := server.Subscribe(source.subject, func(handlerCtx context.Context, envelope messaging.Envelope) error {
			if len(envelope.Payload) == 0 || len(envelope.Payload) > 4096 {
				return errors.New("invalid alert payload size")
			}
			decoder := json.NewDecoder(bytes.NewReader(envelope.Payload))
			decoder.DisallowUnknownFields()
			var alert alertcontract.AlertPayload
			if err := decoder.Decode(&alert); err != nil {
				return err
			}
			var trailing any
			if err := decoder.Decode(&trailing); err != io.EOF {
				return errors.New("trailing alert JSON")
			}
			if err := validate.Struct(alert); err != nil {
				return err
			}
			if alert.EventID == "" {
				return errors.New("alert creation requires event ID for automation")
			}
			if alert.Source == sharedcontracts.ServiceSecurityAutomation {
				return nil
			}
			select {
			case output <- alertObservation{Kind: kind, Alert: alert}:
				return nil
			case <-handlerCtx.Done():
				return handlerCtx.Err()
			case <-ctx.Done():
				return ctx.Err()
			}
		})
		if err != nil {
			return err
		}
	}
	return nil
}
