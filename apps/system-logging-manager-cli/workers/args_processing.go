package workers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrHelpRequested = errors.New("help requested")

type Command string

const (
	CommandCreateEvent                   Command = "createEvent"
	CommandCreateOccurrence              Command = "createOccurrence"
	CommandGetEvent                      Command = "getEvent"
	CommandGetAlerts                     Command = "getAlerts"
	CommandGetActiveEvents               Command = "getActiveEvents"
	CommandGetActiveUnacknowledgedEvents Command = "getActiveUnacknowledgedEvents"
	CommandUpdateEventState              Command = "updateEventState"
	CommandAcknowledgeEvent              Command = "acknowledgeEvent"
	CommandMuteEvent                     Command = "muteEvent"
	defaultListPage                              = 1
	commandCreateOccurrenceAlias         Command = "createOccurence"
	commandGetEventsAlias                Command = "getEvents"
	commandGetActiveUnacknowledgedAlias  Command = "getActiveUnacknowladgedEvents"
)

// Invocation contains validated CLI execution settings.
type Invocation struct {
	ConfigPath string
	Command    Command
	Data       string
	EventID    string
	Page       int
	PageSize   int
	JSONOutput bool
}

// ArgsProcessor parses raw CLI arguments into an Invocation.
type ArgsProcessor interface {
	Process(args []string) (Invocation, error)
}

type argsProcessor struct {
	defaultConfigPath string
}

type invocationValidatorFunc func(Invocation) (Invocation, error)

var invocationValidators = map[Command]invocationValidatorFunc{
	CommandCreateEvent:                   validateDataCommand,
	CommandCreateOccurrence:              validateCreateOccurrenceCommand,
	CommandGetEvent:                      validateGetEventCommand,
	CommandGetAlerts:                     validateListCommand,
	CommandGetActiveEvents:               validateListCommand,
	CommandGetActiveUnacknowledgedEvents: validateListCommand,
	CommandUpdateEventState:              validateDataCommand,
	CommandAcknowledgeEvent:              validateDataCommand,
	CommandMuteEvent:                     validateDataCommand,
}

// NewArgsProcessor creates argument parsing worker.
func NewArgsProcessor(defaultConfigPath string) ArgsProcessor {
	return argsProcessor{defaultConfigPath: defaultConfigPath}
}

// Process parses and validates CLI arguments.
func (p argsProcessor) Process(args []string) (Invocation, error) {
	invocation := Invocation{
		ConfigPath: p.defaultConfigPath,
		Page:       defaultListPage,
	}

	for i := 0; i < len(args); i++ {
		nextInvocation, nextIndex, err := applyArgument(args, i, invocation)
		if err != nil {
			return Invocation{}, err
		}
		invocation = nextInvocation
		i = nextIndex
	}

	return finalizeInvocation(invocation)
}

func applyArgument(args []string, index int, invocation Invocation) (Invocation, int, error) {
	arg := args[index]
	if arg == "-h" || arg == "--help" {
		return Invocation{}, index, ErrHelpRequested
	}
	if arg == "--json" {
		invocation.JSONOutput = true
		return invocation, index, nil
	}
	if strings.HasPrefix(arg, "--config=") {
		invocation.ConfigPath = strings.TrimPrefix(arg, "--config=")
		return invocation, index, nil
	}

	handler, ok := flagHandlers[arg]
	if !ok {
		return Invocation{}, index, fmt.Errorf("unknown argument: %s", arg)
	}
	return handler(args, index, invocation)
}

type flagHandlerFunc func(args []string, index int, invocation Invocation) (Invocation, int, error)

var flagHandlers = map[string]flagHandlerFunc{
	"--config":   handleConfigFlag,
	"--cmd":      handleCommandFlag,
	"--data":     handleDataFlag,
	"--eventID":  handleEventIDFlag,
	"--page":     handlePageFlag,
	"--pageSize": handlePageSizeFlag,
}

func handleConfigFlag(args []string, index int, invocation Invocation) (Invocation, int, error) {
	value, nextIndex, err := readFlagValue(args, index, "--config")
	if err != nil {
		return Invocation{}, index, err
	}
	invocation.ConfigPath = value
	return invocation, nextIndex, nil
}

func handleCommandFlag(args []string, index int, invocation Invocation) (Invocation, int, error) {
	value, nextIndex, err := readFlagValue(args, index, "--cmd")
	if err != nil {
		return Invocation{}, index, err
	}
	invocation.Command = normalizeCommand(Command(value))
	return invocation, nextIndex, nil
}

func handleDataFlag(args []string, index int, invocation Invocation) (Invocation, int, error) {
	value, nextIndex, err := readFlagValue(args, index, "--data")
	if err != nil {
		return Invocation{}, index, err
	}
	invocation.Data = value
	return invocation, nextIndex, nil
}

func handleEventIDFlag(args []string, index int, invocation Invocation) (Invocation, int, error) {
	value, nextIndex, err := readFlagValue(args, index, "--eventID")
	if err != nil {
		return Invocation{}, index, err
	}
	invocation.EventID = value
	return invocation, nextIndex, nil
}

func handlePageFlag(args []string, index int, invocation Invocation) (Invocation, int, error) {
	value, nextIndex, err := readFlagValue(args, index, "--page")
	if err != nil {
		return Invocation{}, index, err
	}
	page, parseErr := strconv.Atoi(value)
	if parseErr != nil {
		return Invocation{}, index, fmt.Errorf("invalid --page value: %w", parseErr)
	}
	invocation.Page = page
	return invocation, nextIndex, nil
}

func handlePageSizeFlag(args []string, index int, invocation Invocation) (Invocation, int, error) {
	value, nextIndex, err := readFlagValue(args, index, "--pageSize")
	if err != nil {
		return Invocation{}, index, err
	}
	pageSize, parseErr := strconv.Atoi(value)
	if parseErr != nil {
		return Invocation{}, index, fmt.Errorf("invalid --pageSize value: %w", parseErr)
	}
	invocation.PageSize = pageSize
	return invocation, nextIndex, nil
}

func readFlagValue(args []string, index int, flag string) (string, int, error) {
	nextIndex := index + 1
	if nextIndex >= len(args) {
		return "", index, fmt.Errorf("%s requires a value", flag)
	}
	return args[nextIndex], nextIndex, nil
}

func normalizeCommand(command Command) Command {
	switch command {
	case commandCreateOccurrenceAlias:
		return CommandCreateOccurrence
	case commandGetEventsAlias:
		return CommandGetAlerts
	case commandGetActiveUnacknowledgedAlias:
		return CommandGetActiveUnacknowledgedEvents
	default:
		return command
	}
}

func finalizeInvocation(invocation Invocation) (Invocation, error) {
	if invocation.ConfigPath == "" {
		return Invocation{}, errors.New("config path must not be empty")
	}
	if invocation.Command == "" {
		return Invocation{}, errors.New("--cmd is required")
	}

	validator, ok := invocationValidators[invocation.Command]
	if !ok {
		return Invocation{}, fmt.Errorf("unsupported --cmd value: %s", invocation.Command)
	}
	return validator(invocation)
}

func validateCreateOccurrenceCommand(invocation Invocation) (Invocation, error) {
	if invocation.EventID == "" {
		return Invocation{}, errors.New("--eventID is required for createOccurrence")
	}
	return validateDataCommand(invocation)
}

func validateGetEventCommand(invocation Invocation) (Invocation, error) {
	if invocation.EventID == "" {
		return Invocation{}, errors.New("--eventID is required for getEvent")
	}
	return invocation, nil
}

func validateListCommand(invocation Invocation) (Invocation, error) {
	if invocation.Page < 1 {
		return Invocation{}, errors.New("--page must be greater than or equal to 1")
	}
	if invocation.PageSize < 0 {
		return Invocation{}, errors.New("--pageSize must be greater than or equal to 0")
	}
	return invocation, nil
}

func validateDataCommand(invocation Invocation) (Invocation, error) {
	if invocation.Data == "" {
		return Invocation{}, errors.New("--data is required for this command")
	}
	return invocation, nil
}
