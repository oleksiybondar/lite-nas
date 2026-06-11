package checks

import (
	"context"
	"os/exec"
)

// Runner executes external commands used by RBAC permission checks.
type Runner interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

// ExecRunner executes commands through os/exec.
type ExecRunner struct{}

// Run executes one command and returns combined stdout and stderr output.
func (ExecRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	// #nosec G204 -- Command arguments are provided by trusted internal RBAC flows.
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}
