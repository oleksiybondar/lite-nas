package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const ufwExecutable = "/usr/sbin/ufw"

var (
	ufwIncidentID    = regexp.MustCompile(`^[A-Za-z0-9]{16,64}$`)
	ufwInterfaceName = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,15}$`)
	numberedUFWRule  = regexp.MustCompile(`^\[\s*([0-9]+)\]\s+(.+)$`)
)

// validateUFWSource accepts one canonical unicast address as a single UFW operand.
func validateUFWSource(value string) (netip.Addr, error) {
	address, err := netip.ParseAddr(value)
	if err != nil || !address.IsGlobalUnicast() || address.Zone() != "" || address.Unmap().String() != value {
		return netip.Addr{}, errors.New("source must be a canonical unicast IP address")
	}
	return address, nil
}

// physicalIngressInterface confirms the interface is a host device backed
// by a sysfs device link outside the virtual-device tree.
func physicalIngressInterface(name string) error {
	if !ufwInterfaceName.MatchString(name) || name[0] == '-' {
		return errors.New("invalid ingress interface name")
	}
	iface, err := net.InterfaceByName(name)
	if err != nil || iface.Flags&net.FlagLoopback != 0 {
		return errors.New("ingress interface is unavailable or loopback")
	}
	devicePath, err := filepath.EvalSymlinks(filepath.Join("/sys/class/net", name, "device"))
	if err != nil || !physicalDevicePath(devicePath) {
		return errors.New("ingress interface is not a verified physical device")
	}
	return nil
}

// physicalDevicePath distinguishes bus-backed devices from virtual net devices.
func physicalDevicePath(path string) bool {
	return strings.HasPrefix(path, "/sys/devices/") && !strings.HasPrefix(path, "/sys/devices/virtual/")
}

// ufwHelperArguments constructs one of the two privileged rule forms permitted
// to the service. Restore may target a recorded interface that has disappeared;
// its numbered UFW rule and incident comment are verified before deletion.
func ufwHelperArguments(operation, incidentID, iface, source, actionName string, verifyInterface func(string) error) ([]string, error) {
	if !ufwIncidentID.MatchString(incidentID) {
		return nil, errors.New("invalid incident ID")
	}
	if verifyInterface == nil {
		return nil, errors.New("missing interface verification")
	}
	if !ufwInterfaceName.MatchString(iface) || iface[0] == '-' {
		return nil, errors.New("invalid ingress interface name")
	}
	if operation == "apply" {
		if err := verifyInterface(iface); err != nil {
			return nil, err
		}
	}
	if _, err := validateUFWSource(source); err != nil {
		return nil, err
	}
	if actionName != "" && actionName != "block_source_ip" && actionName != "block_scan_source_ip" {
		return nil, errors.New("unsupported UFW action")
	}
	comment := "lite-nas-security-automation-" + incidentID
	rule := []string{"deny", "in", "on", iface, "from", source, "to", "any", "comment", comment}
	if actionName != "block_scan_source_ip" {
		rule = []string{"deny", "in", "on", iface, "proto", "tcp", "from", source, "to", "any", "port", "443", "comment", comment}
	}
	switch operation {
	case "apply":
		return append([]string{"insert", "1"}, rule...), nil
	case "restore":
		return append([]string{"delete"}, rule...), nil
	default:
		return nil, errors.New("unsupported UFW helper operation")
	}
}

// runUFWHelper is the root-only entry point invoked through a fixed sudoers
// command. It never accepts executable paths, arbitrary UFW verbs, or shell text.
func runUFWHelper(operation, incidentID, iface, source, actionName string) error {
	if os.Geteuid() != 0 {
		return errors.New("UFW helper requires root")
	}
	args, err := ufwHelperArguments(operation, incidentID, iface, source, actionName, physicalIngressInterface)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	before, err := readUFWStatus(ctx)
	if err != nil {
		return err
	}
	_, present, err := findOwnedUFWRule(before, incidentID, iface, source, actionName)
	if err != nil {
		return err
	}
	if operation == "apply" && present {
		return errors.New("incident UFW rule already exists")
	}
	if operation == "apply" && hasScopedUFWRule(before, iface, source, actionName) {
		return errors.New("matching source deny already exists; refusing to claim another owner's rule")
	}
	if operation == "restore" && !present {
		return nil // A pending record may have been written before apply began.
	}
	output, err := invokeUFW(ctx, args...)
	if err != nil {
		return fmt.Errorf("UFW %s failed: %w", operation, err)
	}
	if operation == "apply" && !strings.Contains(output, "Rule inserted") && !strings.Contains(output, "Rule added") {
		return errors.New("UFW did not confirm insertion of a new rule")
	}
	after, err := readUFWStatus(ctx)
	if err != nil {
		return err
	}
	updatedNumber, updatedPresent, err := findOwnedUFWRule(after, incidentID, iface, source, actionName)
	if err != nil {
		return err
	}
	if operation == "apply" && (!updatedPresent || updatedNumber != 1) || operation == "restore" && updatedPresent {
		return fmt.Errorf("UFW %s could not be verified", operation)
	}
	return nil
}

// invokeUFW runs only the fixed executable with separate validated arguments.
func invokeUFW(ctx context.Context, args ...string) (string, error) {
	command := exec.CommandContext(ctx, ufwExecutable, args...)
	command.Env = append(os.Environ(), "LC_ALL=C")
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, output)
	}
	return string(output), nil
}

// readUFWStatus refuses a disabled firewall, where a persisted rule would not
// represent an active mitigation.
func readUFWStatus(ctx context.Context) (string, error) {
	status, err := invokeUFW(ctx, "status", "numbered")
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(status, "Status: active\n") {
		return "", errors.New("UFW is not active")
	}
	return status, nil
}

// findOwnedUFWRule selects a numbered deny by its same-line incident comment,
// then verifies the source, interface, protocol, port, and direction.
func findOwnedUFWRule(status, incidentID, iface, source, actionName string) (int, bool, error) {
	comment := "lite-nas-security-automation-" + incidentID
	lines := strings.Split(status, "\n")
	found := 0
	for _, line := range lines {
		body, actualComment, hasComment := strings.Cut(line, " # ")
		if !hasComment || strings.TrimSpace(actualComment) != comment {
			continue
		}
		if found != 0 {
			return 0, false, errors.New("ambiguous incident UFW comment")
		}
		parts := numberedUFWRule.FindStringSubmatch(strings.TrimSpace(body))
		if parts == nil {
			return 0, false, errors.New("incident UFW comment lacks a numbered rule")
		}
		if !scopedUFWFields(parts[2], iface, source, actionName) {
			return 0, false, errors.New("incident UFW rule differs from its recorded scope")
		}
		number, err := strconv.Atoi(parts[1])
		if err != nil || number < 1 {
			return 0, false, errors.New("invalid incident UFW rule number")
		}
		found = number
	}
	return found, found != 0, nil
}

// hasScopedUFWRule rejects a duplicate deny even if it belongs to another
// owner, since UFW can update an existing rule's comment instead of adding one.
func hasScopedUFWRule(status, iface, source, actionName string) bool {
	for _, line := range strings.Split(status, "\n") {
		body, _, _ := strings.Cut(line, " # ")
		parts := numberedUFWRule.FindStringSubmatch(strings.TrimSpace(body))
		if parts != nil && scopedUFWFields(parts[2], iface, source, actionName) {
			return true
		}
	}
	return false
}

// scopedUFWFields accepts only the recorded inbound HTTPS or all-traffic deny.
func scopedUFWFields(body, iface, source, actionName string) bool {
	fields := strings.Fields(body)
	target := "443/tcp"
	if actionName == "block_scan_source_ip" {
		target = "Anywhere"
	} else if actionName != "" && actionName != "block_source_ip" {
		return false
	}
	if len(fields) < 5 || fields[0] != target {
		return false
	}
	index := 1
	if fields[index] == "(v6)" {
		index++
	}
	if len(fields) <= index+3 || fields[index] != "on" || fields[index+1] != iface || fields[index+2] != "DENY" {
		return false
	}
	addressIndex := index + 3
	if fields[addressIndex] == "IN" {
		addressIndex++
	}
	return len(fields) == addressIndex+1 && fields[addressIndex] == source
}
