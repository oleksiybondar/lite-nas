package main

import (
	"context"

	"lite-nas/shared/fsm"
)

type (
	integrityState  string
	integrityInput  string
	integrityEffect string
)

const (
	integrityNormal     integrityState  = "normal"
	integrityIncident   integrityState  = "incident"
	integrityFailure    integrityInput  = "failure"
	integrityRecovered  integrityInput  = "recovered"
	integrityCreate     integrityEffect = "create"
	integrityOccurrence integrityEffect = "occurrence"
	integrityNormalize  integrityEffect = "normalize"
)

// integrityLifecycle is the validated transition table for AIDE incidents.
// The active incident pointer remains authoritative because it advances only
// after the logging manager operation succeeds.
var integrityLifecycle = newIntegrityLifecycle()

func newIntegrityLifecycle() *fsm.Machine[integrityState, integrityInput, struct{}, integrityEffect] {
	action := func(effect integrityEffect) fsm.Action[struct{}, integrityInput, integrityEffect] {
		return func(_ context.Context, data struct{}, _ integrityInput) (struct{}, []integrityEffect, error) {
			return data, []integrityEffect{effect}, nil
		}
	}
	machine, err := fsm.New(fsm.Definition[integrityState, integrityInput, struct{}, integrityEffect]{
		Initial: integrityNormal,
		States:  []integrityState{integrityNormal, integrityIncident},
		Inputs:  []integrityInput{integrityFailure, integrityRecovered},
		Rows: []fsm.Row[integrityState, integrityInput]{
			{From: integrityNormal, On: integrityFailure, To: integrityIncident, Action: string(integrityCreate)},
			{From: integrityNormal, On: integrityRecovered, To: integrityNormal},
			{From: integrityIncident, On: integrityFailure, To: integrityIncident, Action: string(integrityOccurrence)},
			{From: integrityIncident, On: integrityRecovered, To: integrityNormal, Action: string(integrityNormalize)},
		},
		Actions: map[string]fsm.Action[struct{}, integrityInput, integrityEffect]{
			string(integrityCreate):     action(integrityCreate),
			string(integrityOccurrence): action(integrityOccurrence),
			string(integrityNormalize):  action(integrityNormalize),
		},
	})
	if err != nil {
		panic(err)
	} // The table is fixed source code.
	return machine
}

// integrityTransition chooses the next logging-manager effect from the
// current successful incident cache and one factual AIDE observation.
func integrityTransition(ctx context.Context, active bool, input integrityInput) (integrityEffect, error) {
	state := integrityNormal
	if active {
		state = integrityIncident
	}
	result, err := integrityLifecycle.Step(ctx, fsm.Instance[integrityState, struct{}]{State: state}, input)
	if err != nil {
		return "", err
	}
	if len(result.Effects) == 0 {
		return "", nil
	}
	return result.Effects[0], nil
}
