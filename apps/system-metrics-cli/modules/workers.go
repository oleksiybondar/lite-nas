package modules

import "lite-nas/apps/system-metrics-cli/workers"

// Workers groups the stateless workers used by the CLI runtime.
type Workers struct {
	argsProcessor workers.ArgsProcessor
	outputWriter  workers.OutputWriter
}

// NewWorkersModule creates the workers required by the CLI runtime.
func NewWorkersModule(defaultConfigPath string) Workers {
	return Workers{
		argsProcessor: workers.NewArgsProcessor(defaultConfigPath),
		outputWriter:  workers.NewOutputWriter(),
	}
}

// ArgsProcessor returns the argument processor.
func (m Workers) ArgsProcessor() workers.ArgsProcessor {
	return m.argsProcessor
}

// OutputWriter returns the output renderer.
func (m Workers) OutputWriter() workers.OutputWriter {
	return m.outputWriter
}
