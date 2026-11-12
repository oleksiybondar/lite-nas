// Package hostcommandtest records external command calls made by host services.
package hostcommandtest

import "context"

// RecordingRunner captures the exact executable and arguments without host effects.
type RecordingRunner struct {
	Name        string
	Args        []string
	OutputBytes []byte
	Results     []Result
	Calls       int
}

// Result supplies one ordered command output or failure.
type Result struct {
	Output []byte
	Err    error
}

// Output records one invocation and returns configured output.
func (r *RecordingRunner) Output(_ context.Context, name string, args ...string) ([]byte, error) {
	r.Name, r.Args = name, args
	if r.Calls < len(r.Results) {
		result := r.Results[r.Calls]
		r.Calls++
		return result.Output, result.Err
	}
	r.Calls++
	return r.OutputBytes, nil
}
