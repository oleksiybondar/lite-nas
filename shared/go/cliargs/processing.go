package cliargs

import "errors"

// ApplyFunc applies one argument to an invocation value.
type ApplyFunc[T any] func(invocation T, arg string) (T, error)

// FinalizeFunc validates/finalizes the invocation after all args are applied.
type FinalizeFunc[T any] func(invocation T) (T, error)

// Process applies all CLI args in order and then runs finalize.
func Process[T any](
	args []string,
	initial T,
	apply ApplyFunc[T],
	finalize FinalizeFunc[T],
) (T, error) {
	invocation := initial

	for _, arg := range args {
		nextInvocation, err := apply(invocation, arg)
		if err != nil {
			var zero T
			return zero, err
		}

		invocation = nextInvocation
	}

	return finalize(invocation)
}

// ProcessWithConfig handles shared help/config flags and requires a nonempty
// config path before running service-specific finalization.
func ProcessWithConfig[T any](
	args []string,
	initial T,
	configPath func(*T) *string,
	apply ApplyFunc[T],
	finalize FinalizeFunc[T],
) (T, error) {
	return Process(args, initial, func(invocation T, arg string) (T, error) {
		if handled, err := ApplyHelpAndConfigArg(arg, configPath(&invocation)); handled {
			return invocation, err
		}
		return apply(invocation, arg)
	}, func(invocation T) (T, error) {
		if *configPath(&invocation) == "" {
			var zero T
			return zero, errors.New("config path must not be empty")
		}
		if finalize == nil {
			return invocation, nil
		}
		return finalize(invocation)
	})
}
