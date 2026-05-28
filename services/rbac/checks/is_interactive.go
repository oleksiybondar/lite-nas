package checks

import (
	"context"
	"fmt"
	"strings"
)

// IsInteractiveUser reports whether one UID appears to be an interactive user.
func IsInteractiveUser(ctx context.Context, runner Runner, uid string) (bool, error) {
	output, err := runner.Run(ctx, "getent", "passwd", uid)
	if err != nil {
		return false, fmt.Errorf("getent passwd failed: %w", err)
	}

	line := strings.TrimSpace(string(output))
	if line == "" {
		return false, nil
	}

	fields := strings.Split(line, ":")
	if len(fields) < 7 {
		return false, fmt.Errorf("invalid passwd entry format")
	}

	shell := strings.TrimSpace(fields[6])
	return isInteractiveShell(shell), nil
}

func isInteractiveShell(shell string) bool {
	switch shell {
	case "", "/sbin/nologin", "/usr/sbin/nologin", "/bin/false", "/usr/bin/false":
		return false
	default:
		return true
	}
}
