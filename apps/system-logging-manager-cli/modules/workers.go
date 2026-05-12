package modules

import "lite-nas/apps/system-logging-manager-cli/workers"

// Workers groups stateless worker dependencies.
type Workers struct {
	ArgsProcessor workers.ArgsProcessor
	OutputWriter  workers.OutputWriter
}

// NewWorkersModule assembles runtime workers.
func NewWorkersModule(defaultConfigPath string) Workers {
	return Workers{
		ArgsProcessor: workers.NewArgsProcessor(defaultConfigPath),
		OutputWriter:  workers.NewOutputWriter(),
	}
}
