// Command lite-nas-file-writer writes bounded upload data as its effective Unix user.
package main

import (
	"encoding/json"
	"errors"
	"os"
)

// main maps a failed invocation to a nonzero process exit.
func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		failure := struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
				Cleanup string `json:"cleanup,omitempty"`
			} `json:"error"`
		}{}
		failure.Error.Code = "operation_failed"
		failure.Error.Message = err.Error()
		var operationFailure *operationError
		if errors.As(err, &operationFailure) {
			failure.Error.Code = operationFailure.Code
			failure.Error.Cleanup = operationFailure.Cleanup
		}
		_ = json.NewEncoder(os.Stderr).Encode(failure)
		os.Exit(2)
	}
}
