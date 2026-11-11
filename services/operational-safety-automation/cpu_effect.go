package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	cpuPolicyRoot     = "/sys/devices/system/cpu/cpufreq"
	serviceExecutable = "/usr/libexec/lite-nas/operational-safety-automation"
)

var cpuPolicyName = regexp.MustCompile(`^policy[0-9]+$`)

// cpuPolicy records one kernel CPU frequency policy before a change.
type cpuPolicy struct {
	path string
	old  uint64
}

// applyCPUCap invokes only the fixed root helper with a validated numeric cap.
func applyCPUCap(ctx context.Context, capKHz, expectedKHz uint64) error {
	if capKHz == 0 {
		return errors.New("CPU cap must be positive")
	}
	commandCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	output, err := exec.CommandContext(commandCtx, "/usr/bin/sudo", "-n", "--", serviceExecutable, "cpu-cap-helper", strconv.FormatUint(capKHz, 10), strconv.FormatUint(expectedKHz, 10)).CombinedOutput()
	if err != nil {
		return fmt.Errorf("CPU cap helper: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

// applyManualCPUCap invokes the fixed privileged helper for a separately authorized operator request.
func applyManualCPUCap(ctx context.Context, capKHz uint64) error {
	if capKHz == 0 {
		return errors.New("CPU cap must be positive")
	}
	commandCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	output, err := exec.CommandContext(commandCtx, "/usr/bin/sudo", "-n", "--", serviceExecutable, "manual-cpu-cap-helper", strconv.FormatUint(capKHz, 10)).CombinedOutput()
	if err != nil {
		return fmt.Errorf("manual CPU cap helper: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

// runCPUCapHelper is the root-only validated sysfs entry point.
func runCPUCapHelper(raw, expectedRaw string) error {
	if os.Geteuid() != 0 {
		return errors.New("CPU cap helper requires root")
	}
	return runCPUCapHelperAt(cpuPolicyRoot, raw, expectedRaw)
}

// runManualCPUCapHelper applies an operator-selected cap without claiming prior ownership.
func runManualCPUCapHelper(raw string) error {
	if os.Geteuid() != 0 {
		return errors.New("manual CPU cap helper requires root")
	}
	return runCPUCapHelperAtMode(cpuPolicyRoot, raw, "0", true)
}

// runCPUCapHelperAt applies the fixed policy operation at an injected root for tests.
func runCPUCapHelperAt(root, raw, expectedRaw string) error {
	return runCPUCapHelperAtMode(root, raw, expectedRaw, false)
}

// runCPUCapHelperAtMode shares policy validation while restricting cap relaxation by mode.
func runCPUCapHelperAtMode(root, raw, expectedRaw string, manual bool) error {
	capKHz, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || capKHz == 0 || strconv.FormatUint(capKHz, 10) != raw {
		return errors.New("invalid CPU cap")
	}
	expectedKHz, err := strconv.ParseUint(expectedRaw, 10, 64)
	if err != nil || strconv.FormatUint(expectedKHz, 10) != expectedRaw {
		return errors.New("invalid expected CPU cap")
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	policies := make([]cpuPolicy, 0)
	for _, entry := range entries {
		if !cpuPolicyName.MatchString(entry.Name()) || !entry.IsDir() {
			continue
		}
		path := filepath.Join(root, entry.Name())
		old, err := readPolicyNumber(path, "scaling_max_freq")
		if err != nil {
			return err
		}
		min, err := readPolicyNumber(path, "scaling_min_freq")
		if err != nil {
			return err
		}
		hwMax, err := readPolicyNumber(path, "cpuinfo_max_freq")
		if err != nil {
			return err
		}
		hwMin, err := readPolicyNumber(path, "cpuinfo_min_freq")
		if err != nil {
			return err
		}
		if !manual && capKHz > old && (expectedKHz == 0 || old != expectedKHz) {
			return fmt.Errorf("refusing to relax unowned policy %s", entry.Name())
		}
		if capKHz < min || capKHz < hwMin || capKHz > hwMax {
			return fmt.Errorf("cap %d unsupported by %s", capKHz, entry.Name())
		}
		policies = append(policies, cpuPolicy{path: path, old: old})
	}
	if len(policies) == 0 {
		return errors.New("no CPU frequency policy available")
	}
	changed := make([]cpuPolicy, 0, len(policies))
	for _, policy := range policies {
		if policy.old == capKHz {
			continue
		}
		if err := writePolicyNumber(policy.path, capKHz); err != nil {
			return rollbackCPUPolicies(changed, err)
		}
		changed = append(changed, policy)
		got, err := readPolicyNumber(policy.path, "scaling_max_freq")
		if err != nil || got != capKHz {
			return rollbackCPUPolicies(changed, fmt.Errorf("CPU policy verification failed for %s: %w", policy.path, err))
		}
	}
	return nil
}

// readPolicyNumber parses one trusted kernel policy attribute in kHz.
func readPolicyNumber(policyPath, name string) (uint64, error) {
	data, err := os.ReadFile(filepath.Join(policyPath, name))
	if err != nil {
		return 0, err
	}
	value, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
	if err != nil || value == 0 {
		return 0, fmt.Errorf("invalid %s for %s", name, policyPath)
	}
	return value, nil
}

// writePolicyNumber changes one policy's maximum and verifies it at the caller.
func writePolicyNumber(policyPath string, capKHz uint64) error {
	file, err := os.OpenFile(filepath.Join(policyPath, "scaling_max_freq"), os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	_, writeErr := file.WriteString(strconv.FormatUint(capKHz, 10))
	closeErr := file.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}

// rollbackCPUPolicies attempts to restore every policy modified in this helper call.
func rollbackCPUPolicies(changed []cpuPolicy, cause error) error {
	for index := len(changed) - 1; index >= 0; index-- {
		policy := changed[index]
		if err := writePolicyNumber(policy.path, policy.old); err != nil {
			cause = errors.Join(cause, fmt.Errorf("rollback %s: %w", policy.path, err))
		}
	}
	return cause
}
