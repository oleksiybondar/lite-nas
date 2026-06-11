package cliargs

import (
	"errors"
	"strings"
)

// ErrHelpRequested indicates that usage/help output was requested.
var ErrHelpRequested = errors.New("help requested")

// ApplyHelpAndConfigArg handles common CLI flags and updates config path.
//
// Returns handled=true when arg is recognized as either help or config flag.
func ApplyHelpAndConfigArg(arg string, configPath *string) (handled bool, err error) {
	if arg == "-h" || arg == "--help" {
		return true, ErrHelpRequested
	}

	if strings.HasPrefix(arg, "--config=") {
		*configPath = strings.TrimPrefix(arg, "--config=")
		return true, nil
	}

	return false, nil
}
