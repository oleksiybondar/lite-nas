package workers

import (
	"errors"
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
	return cliargs.Process(
		args,
		Invocation{
			ConfigPath: p.defaultConfigPath,
			Mode:       ModeCurrent,
		},
		applyArg,
		finalizeInvocation,
	)
}

func applyArg(invocation Invocation, arg string) (Invocation, error) {
	if handled, err := cliargs.ApplyHelpAndConfigArg(arg, &invocation.ConfigPath); handled {
		if err != nil {
			return Invocation{}, err
		}
		return invocation, nil
	}

	switch arg {
	case "--history":
		invocation.Mode = ModeHistory
		return invocation, nil
	default:
		return Invocation{}, fmt.Errorf("unknown argument: %s", arg)
	}
}

func finalizeInvocation(invocation Invocation) (Invocation, error) {
	if invocation.ConfigPath == "" {
		return Invocation{}, errors.New("config path must not be empty")
	}

	return invocation, nil
}
