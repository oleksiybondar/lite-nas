package modules

import "lite-nas/apps/network-metrics-cli/workers"

// Workers groups the stateless workers used by the CLI runtime.
type Workers struct {
	ArgsProcessor workers.ArgsProcessor
	OutputWriter  workers.OutputWriter
}

// NewWorkersModule assembles the workers required by the CLI runtime.
func NewWorkersModule(defaultConfigPath string) Workers {
	return Workers{
		ArgsProcessor: workers.NewArgsProcessor(defaultConfigPath),
		OutputWriter:  workers.NewOutputWriter(),
	}
}
