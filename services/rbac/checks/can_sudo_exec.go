package checks

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
)

// CanSudoExec reports whether one UID is allowed to run one command through sudo policy.
func CanSudoExec(ctx context.Context, runner Runner, uid string, command string) (bool, error) {
	if err := ValidateSudoArguments(uid, command); err != nil {
		return false, err
	}
	identity, err := ResolveIdentityByUID(ctx, runner, uid)
	if err != nil {
		return false, err
	}

	username := identity.Username
	if err := ValidateIdentityOperand(username); err != nil {
		return false, err
	}

	_, err = runner.Run(ctx, "sudo", "-n", "-l", "-U", username, "--", command)
	if err == nil {
		return true, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return false, nil
	}

	return false, fmt.Errorf("sudo policy check failed: %w", err)
}
