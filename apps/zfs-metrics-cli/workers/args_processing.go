package workers

import (
	"fmt"

	"lite-nas/shared/cliargs"
)

var ErrHelpRequested = cliargs.ErrHelpRequested

// Mode defines the selected CLI action mode.
type Mode string

const (
	ModeCurrent Mode = "current"
	ModeHistory Mode = "history"
)

// Invocation contains parsed CLI execution settings.
type Invocation struct {
	ConfigPath string
	Mode       Mode
}

// ArgsProcessor parses CLI arguments into an invocation.
type ArgsProcessor interface {
	Process(args []string) (Invocation, error)
}

type argsProcessor struct {
	defaultConfigPath string
}

// NewArgsProcessor creates an argument parsing worker.
func NewArgsProcessor(defaultConfigPath string) ArgsProcessor {
	return argsProcessor{defaultConfigPath: defaultConfigPath}
}

// Process parses CLI args into one validated invocation.
func (p argsProcessor) Process(args []string) (Invocation, error) {
	return cliargs.ProcessWithConfig(
		args,
		Invocation{
			ConfigPath: p.defaultConfigPath,
			Mode:       ModeCurrent,
		},
		func(invocation *Invocation) *string { return &invocation.ConfigPath },
		applyArg,
		nil,
	)
}

func applyArg(invocation Invocation, arg string) (Invocation, error) {
	switch arg {
	case "--history":
		invocation.Mode = ModeHistory
		return invocation, nil
	default:
		return Invocation{}, fmt.Errorf("unknown argument: %s", arg)
	}
}
