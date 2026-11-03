// Package systemd reads factual properties of explicitly named service units
// without privileged service-management access.
package systemd

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const SystemctlBinary = "/usr/bin/systemctl"

var unitNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.@-]*\.service$`)

// UnitState holds factual properties returned by systemd for one allowlisted
// service. It does not infer who stopped or changed the service.
type UnitState struct {
	Unit         string
	ActiveState  string
	SubState     string
	Result       string
	RestartCount uint64
}

// ValidUnitName limits command operands to local service units and excludes
// option-shaped, path-shaped, and template-wildcard input.
func ValidUnitName(unit string) bool {
	return len(unit) >= 9 && len(unit) <= 256 && unitNamePattern.MatchString(unit)
}

// ProbeUnit runs a fixed unprivileged systemctl command with a validated unit
// operand. No shell or sudo is involved in the observation path.
func ProbeUnit(ctx context.Context, unit string) (UnitState, error) {
	if !ValidUnitName(unit) {
		return UnitState{}, fmt.Errorf("invalid service unit %q", unit)
	}
	commandCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	command := exec.CommandContext(commandCtx, SystemctlBinary, "--system", "show", "--no-pager",
		"--property=ActiveState,SubState,Result,NRestarts", "--", unit)
	output, err := command.Output()
	if err != nil {
		return UnitState{}, fmt.Errorf("systemctl show %s: %w", unit, err)
	}
	return ParseUnitState(unit, string(output))
}

// ParseUnitState bounds and validates the four known properties before the
// monitor decides whether a condition or transition exists.
func ParseUnitState(unit, output string) (UnitState, error) {
	if !ValidUnitName(unit) || len(output) > 4096 {
		return UnitState{}, fmt.Errorf("invalid systemctl output")
	}
	state := UnitState{Unit: unit}
	seen := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		key, value, ok := strings.Cut(line, "=")
		if !ok || seen[key] {
			return UnitState{}, fmt.Errorf("malformed systemctl property")
		}
		seen[key] = true
		if len(value) > 128 || strings.ContainsAny(value, "\r\x00") {
			return UnitState{}, fmt.Errorf("invalid systemctl property value")
		}
		switch key {
		case "ActiveState":
			state.ActiveState = value
		case "SubState":
			state.SubState = value
		case "Result":
			state.Result = value
		case "NRestarts":
			count, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return UnitState{}, fmt.Errorf("invalid NRestarts")
			}
			state.RestartCount = count
		default:
			return UnitState{}, fmt.Errorf("unknown systemctl property %q", key)
		}
	}
	if !seen["ActiveState"] || !seen["SubState"] || !seen["Result"] || !seen["NRestarts"] || state.ActiveState == "" || state.SubState == "" {
		return UnitState{}, fmt.Errorf("incomplete systemctl properties")
	}
	return state, nil
}

// EnablementState returns systemctl's factual unit-file state. A nonzero exit
// is accepted for recognized disabled or masked states, which are observations
// rather than command failures.
func EnablementState(ctx context.Context, unit string) (string, error) {
	if !ValidUnitName(unit) {
		return "", fmt.Errorf("invalid service unit %q", unit)
	}
	commandCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	output, err := exec.CommandContext(commandCtx, SystemctlBinary, "--system", "is-enabled", "--", unit).Output()
	if commandCtx.Err() != nil {
		return "", commandCtx.Err()
	}
	state := strings.TrimSpace(string(output))
	if len(state) > 64 {
		return "", fmt.Errorf("systemctl enablement output too large")
	}
	switch state {
	case "enabled", "enabled-runtime", "linked", "linked-runtime", "alias", "static", "indirect", "generated", "transient", "disabled", "masked", "masked-runtime":
		return state, nil
	default:
		return "", fmt.Errorf("systemctl is-enabled %s returned %q: %v", unit, state, err)
	}
}
