package modules

import "lite-nas/apps/zfs-metrics-cli/workers"

// Workers groups stateless workers used by the CLI runtime.
type Workers struct {
	ArgsProcessor workers.ArgsProcessor
	OutputWriter  workers.OutputWriter
}

// NewWorkersModule assembles workers required by the CLI runtime.
func NewWorkersModule(defaultConfigPath string) Workers {
	return Workers{
		ArgsProcessor: workers.NewArgsProcessor(defaultConfigPath),
		OutputWriter:  workers.NewOutputWriter(),
	}
}
