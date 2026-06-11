package cliargs

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
