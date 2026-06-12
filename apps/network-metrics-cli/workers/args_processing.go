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

// CurrentSelection describes which current-snapshot sections to render.
type CurrentSelection struct {
	Interfaces bool
	Protocols  bool
	Sockets    bool
	Pressure   bool
}

// Invocation contains the parsed CLI execution settings.
type Invocation struct {
	ConfigPath       string
	Mode             Mode
	CurrentSelection CurrentSelection
}

// ArgsProcessor parses the CLI arguments into an invocation.
type ArgsProcessor interface {
	Process(args []string) (Invocation, error)
}

type argsProcessor struct {
	defaultConfigPath string
}

// NewArgsProcessor creates an argument processing worker.
func NewArgsProcessor(defaultConfigPath string) ArgsProcessor {
	return argsProcessor{defaultConfigPath: defaultConfigPath}
}

// Process parses the CLI arguments into a validated invocation.
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
	case "--interfaces", "--protocols", "--sockets", "--pressure":
		invocation.CurrentSelection.selectFlag(arg)
	default:
		return Invocation{}, fmt.Errorf("unknown argument: %s", arg)
	}

	return invocation, nil
}

func finalizeInvocation(invocation Invocation) (Invocation, error) {
	if invocation.ConfigPath == "" {
		return Invocation{}, errors.New("config path must not be empty")
	}

	if invocation.Mode == ModeHistory {
		return finalizeHistoryInvocation(invocation)
	}

	if !invocation.CurrentSelection.anySelected() {
		invocation.CurrentSelection.selectAll()
	}

	return invocation, nil
}

func finalizeHistoryInvocation(invocation Invocation) (Invocation, error) {
	if invocation.CurrentSelection.anySelected() {
		return Invocation{}, errors.New("--history cannot be combined with section flags")
	}

	return invocation, nil
}

func (selection CurrentSelection) anySelected() bool {
	return selection.Interfaces || selection.Protocols || selection.Sockets || selection.Pressure
}

func (selection *CurrentSelection) selectAll() {
	selection.Interfaces = true
	selection.Protocols = true
	selection.Sockets = true
	selection.Pressure = true
}

func (selection *CurrentSelection) selectFlag(flag string) {
	switch flag {
	case "--interfaces":
		selection.Interfaces = true
	case "--protocols":
		selection.Protocols = true
	case "--sockets":
		selection.Sockets = true
	case "--pressure":
		selection.Pressure = true
	}
}
