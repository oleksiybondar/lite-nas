// Package fsm evaluates typed, table-driven state transitions. It does not
// perform transport, persistence, or privileged effects for its callers.
package fsm

import (
	"context"
	"errors"
	"fmt"
)

// Outcome distinguishes a transition from a deliberate nonmatch or failure.
type Outcome string

const (
	// Matched means a row changed the state.
	Matched Outcome = "matched"
	// Stayed means a row matched and retained the current state.
	Stayed Outcome = "stayed"
	// NotApplicable means no row accepted the input in the current state.
	NotApplicable Outcome = "not_applicable"
	// Reset means a row returned to the machine's initial state.
	Reset Outcome = "reset"
	// EvaluationError means a guard or action failed.
	EvaluationError Outcome = "evaluation_error"
)

// Row names one transition. Guards and actions are registered Go functions;
// configuration may select a machine but cannot supply executable functions.
type Row[S comparable, I comparable] struct {
	From   S
	On     I
	Guard  string
	To     S
	Action string
}

// Guard decides whether a row applies to the current state data and input.
type Guard[D any, I comparable] func(context.Context, D, I) (bool, error)

// Action computes new state data and effect intents without delivering them.
type Action[D any, I comparable, E any] func(context.Context, D, I) (D, []E, error)

// Definition is an immutable machine specification after New validates it.
type Definition[S comparable, I comparable, D any, E any] struct {
	Initial S
	States  []S
	Inputs  []I
	Rows    []Row[S, I]
	Guards  map[string]Guard[D, I]
	Actions map[string]Action[D, I, E]
}

// Instance holds a caller-owned state and its domain data.
type Instance[S comparable, D any] struct {
	State S
	Data  D
}

// Result reports the selected row and typed effects. The caller must commit
// the next instance and account for effects before accepting another input.
type Result[S comparable, I comparable, D any, E any] struct {
	Next    Instance[S, D]
	Outcome Outcome
	Row     *Row[S, I]
	Effects []E
}

// Machine evaluates a validated table. It is safe to share between workers
// because Step only reads its definition and operates on caller-owned data.
type Machine[S comparable, I comparable, D any, E any] struct {
	definition Definition[S, I, D, E]
	states     map[S]struct{}
	inputs     map[I]struct{}
}

// New validates state, input and function references before accepting a table.
// Rows with the same state and input are allowed only when all have guards;
// Step rejects overlapping guards rather than depending on row order.
func New[S comparable, I comparable, D any, E any](definition Definition[S, I, D, E]) (*Machine[S, I, D, E], error) {
	if len(definition.States) == 0 || len(definition.Inputs) == 0 || len(definition.Rows) == 0 {
		return nil, errors.New("fsm requires states, inputs and rows")
	}
	machine := &Machine[S, I, D, E]{definition: definition, states: make(map[S]struct{}), inputs: make(map[I]struct{})}
	for _, state := range definition.States {
		if _, exists := machine.states[state]; exists {
			return nil, fmt.Errorf("duplicate state %v", state)
		}
		machine.states[state] = struct{}{}
	}
	for _, input := range definition.Inputs {
		if _, exists := machine.inputs[input]; exists {
			return nil, fmt.Errorf("duplicate input %v", input)
		}
		machine.inputs[input] = struct{}{}
	}
	if _, exists := machine.states[definition.Initial]; !exists {
		return nil, fmt.Errorf("initial state %v is unknown", definition.Initial)
	}
	type pair struct {
		from S
		on   I
	}
	seen := make(map[pair]int)
	unguarded := make(map[pair]bool)
	for _, row := range definition.Rows {
		if _, exists := machine.states[row.From]; !exists {
			return nil, fmt.Errorf("row source state %v is unknown", row.From)
		}
		if _, exists := machine.states[row.To]; !exists {
			return nil, fmt.Errorf("row target state %v is unknown", row.To)
		}
		if _, exists := machine.inputs[row.On]; !exists {
			return nil, fmt.Errorf("row input %v is unknown", row.On)
		}
		if row.Guard != "" && definition.Guards[row.Guard] == nil {
			return nil, fmt.Errorf("guard %q is unregistered", row.Guard)
		}
		if row.Action != "" && definition.Actions[row.Action] == nil {
			return nil, fmt.Errorf("action %q is unregistered", row.Action)
		}
		key := pair{row.From, row.On}
		seen[key]++
		unguarded[key] = unguarded[key] || row.Guard == ""
		if seen[key] > 1 && unguarded[key] {
			return nil, fmt.Errorf("ambiguous unguarded rows for state %v and input %v", row.From, row.On)
		}
	}
	reachable := map[S]bool{definition.Initial: true}
	for changed := true; changed; {
		changed = false
		for _, row := range definition.Rows {
			if reachable[row.From] && !reachable[row.To] {
				reachable[row.To] = true
				changed = true
			}
		}
	}
	for _, state := range definition.States {
		if !reachable[state] {
			return nil, fmt.Errorf("state %v is unreachable", state)
		}
	}
	// Copy collection headers and maps so callers cannot replace registry
	// entries after validation. Function closures must remain caller-safe.
	machine.definition.States = append([]S(nil), definition.States...)
	machine.definition.Inputs = append([]I(nil), definition.Inputs...)
	machine.definition.Rows = append([]Row[S, I](nil), definition.Rows...)
	machine.definition.Guards = make(map[string]Guard[D, I], len(definition.Guards))
	for name, guard := range definition.Guards {
		machine.definition.Guards[name] = guard
	}
	machine.definition.Actions = make(map[string]Action[D, I, E], len(definition.Actions))
	for name, action := range definition.Actions {
		machine.definition.Actions[name] = action
	}
	return machine, nil
}

// Initial returns a fresh instance using the declared initial state.
func (m *Machine[S, I, D, E]) Initial(data D) Instance[S, D] {
	return Instance[S, D]{State: m.definition.Initial, Data: data}
}

// Rows returns a copy of the declarative table for diagnostics and demos.
func (m *Machine[S, I, D, E]) Rows() []Row[S, I] {
	return append([]Row[S, I](nil), m.definition.Rows...)
}

// Step evaluates all eligible guards, then executes exactly one action. An
// error leaves the original instance unchanged and returns no effects.
func (m *Machine[S, I, D, E]) Step(ctx context.Context, current Instance[S, D], input I) (Result[S, I, D, E], error) {
	result := Result[S, I, D, E]{Next: current, Outcome: NotApplicable}
	if _, exists := m.states[current.State]; !exists {
		return result, fmt.Errorf("unknown current state %v", current.State)
	}
	if _, exists := m.inputs[input]; !exists {
		return result, fmt.Errorf("unknown input %v", input)
	}
	var selected *Row[S, I]
	for index := range m.definition.Rows {
		row := &m.definition.Rows[index]
		if row.From != current.State || row.On != input {
			continue
		}
		if row.Guard != "" {
			accept, err := m.definition.Guards[row.Guard](ctx, current.Data, input)
			if err != nil {
				result.Outcome = EvaluationError
				return result, fmt.Errorf("guard %q in state %v on input %v: %w", row.Guard, current.State, input, err)
			}
			if !accept {
				continue
			}
		}
		if selected != nil {
			result.Outcome = EvaluationError
			return result, fmt.Errorf("multiple rows matched state %v and input %v", current.State, input)
		}
		selected = row
	}
	if selected == nil {
		return result, nil
	}
	data := current.Data
	var effects []E
	if selected.Action != "" {
		var err error
		data, effects, err = m.definition.Actions[selected.Action](ctx, current.Data, input)
		if err != nil {
			result.Outcome = EvaluationError
			return result, fmt.Errorf("action %q from state %v to %v on input %v: %w", selected.Action, current.State, selected.To, input, err)
		}
	}
	result.Next = Instance[S, D]{State: selected.To, Data: data}
	selectedCopy := *selected
	result.Row = &selectedCopy
	result.Effects = effects
	switch {
	case selected.To == m.definition.Initial && current.State != m.definition.Initial:
		result.Outcome = Reset
	case selected.To == current.State:
		result.Outcome = Stayed
	default:
		result.Outcome = Matched
	}
	return result, nil
}
