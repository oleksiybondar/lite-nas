package modules

import "lite-nas/apps/system-metrics-cli/workers"

// Workers groups the stateless workers used by the CLI runtime.
//
// The fields are populated once during startup and are expected to be treated
// as logically read-only after construction.
type Workers struct {
	ArgsProcessor workers.ArgsProcessor
	OutputWriter  workers.OutputWriter
}

// NewWorkersModule assembles the workers required by the CLI runtime.
//
// Parameters:
//   - defaultConfigPath: fallback config path used by argument processing
func NewWorkersModule(defaultConfigPath string) Workers {
	return Workers{
		ArgsProcessor: workers.NewArgsProcessor(defaultConfigPath),
		OutputWriter:  workers.NewOutputWriter(),
	}
}
