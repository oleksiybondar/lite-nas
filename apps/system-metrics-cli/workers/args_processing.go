package workers

import (
	"errors"
	"fmt"
	"strings"
)

var ErrHelpRequested = errors.New("help requested")

type Mode string

const (
	ModeCurrent Mode = "current"
	ModeHistory Mode = "history"
)

// CurrentSelection describes which current-snapshot sections to render.
type CurrentSelection struct {
	CPU bool
	RAM bool
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

type argUpdateFunc func(Invocation) Invocation

var exactArgUpdaters = map[string]argUpdateFunc{
	"--history": setHistoryMode,
	"--cpu":     selectCPUSection,
	"--ram":     selectRAMSection,
}

// NewArgsProcessor creates an argument processing worker.
func NewArgsProcessor(defaultConfigPath string) ArgsProcessor {
	return argsProcessor{defaultConfigPath: defaultConfigPath}
}

// Process parses the CLI arguments into a validated invocation.
func (p argsProcessor) Process(args []string) (Invocation, error) {
	invocation := Invocation{
		ConfigPath: p.defaultConfigPath,
		Mode:       ModeCurrent,
	}

	for _, arg := range args {
		nextInvocation, err := applyArg(invocation, arg)
		if err != nil {
			return Invocation{}, err
		}

		invocation = nextInvocation
	}

	return finalizeInvocation(invocation)
}

func applyArg(invocation Invocation, arg string) (Invocation, error) {
	if arg == "-h" || arg == "--help" {
		return Invocation{}, ErrHelpRequested
	}

	if strings.HasPrefix(arg, "--config=") {
		invocation.ConfigPath = strings.TrimPrefix(arg, "--config=")
		return invocation, nil
	}

	updater, ok := exactArgUpdaters[arg]
	if !ok {
		return Invocation{}, fmt.Errorf("unknown argument: %s", arg)
	}

	return updater(invocation), nil
}

func finalizeInvocation(invocation Invocation) (Invocation, error) {
	if invocation.ConfigPath == "" {
		return Invocation{}, errors.New("config path must not be empty")
	}

	if invocation.Mode == ModeHistory {
		return finalizeHistoryInvocation(invocation)
	}

	if !invocation.CurrentSelection.CPU && !invocation.CurrentSelection.RAM {
		invocation.CurrentSelection.CPU = true
		invocation.CurrentSelection.RAM = true
	}

	return invocation, nil
}

func finalizeHistoryInvocation(invocation Invocation) (Invocation, error) {
	if invocation.CurrentSelection.CPU || invocation.CurrentSelection.RAM {
		return Invocation{}, errors.New("--history cannot be combined with --cpu or --ram")
	}

	return invocation, nil
}

func setHistoryMode(invocation Invocation) Invocation {
	invocation.Mode = ModeHistory
	return invocation
}

func selectCPUSection(invocation Invocation) Invocation {
	invocation.CurrentSelection.CPU = true
	return invocation
}

func selectRAMSection(invocation Invocation) Invocation {
	invocation.CurrentSelection.RAM = true
	return invocation
}
