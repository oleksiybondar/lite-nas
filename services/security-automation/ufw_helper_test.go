package main

import (
	"errors"
	"reflect"
	"testing"
)

// TestUFWHelperArgumentsBuildsScopedRule checks the reviewed command shape.
// Requirements: security-automation/SR-002.
func TestUFWHelperArgumentsBuildsScopedRule(t *testing.T) {
	args, err := ufwHelperArguments("apply", "A1B2C3D4E5F6G7H8", "physical0", "198.51.100.10", "block_source_ip", func(string) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"insert", "1", "deny", "in", "on", "physical0", "proto", "tcp", "from", "198.51.100.10", "to", "any", "port", "443", "comment", "lite-nas-security-automation-A1B2C3D4E5F6G7H8"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("UFW arguments = %v", args)
	}
}

// TestUFWHelperArgumentsBuildsAllTrafficRule checks the scan action blocks
// every inbound port and protocol on one verified physical interface.
// Requirements: security-automation/SR-002.
func TestUFWHelperArgumentsBuildsAllTrafficRule(t *testing.T) {
	args, err := ufwHelperArguments("apply", "A1B2C3D4E5F6G7H8", "physical0", "198.51.100.10", "block_scan_source_ip", func(string) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"insert", "1", "deny", "in", "on", "physical0", "from", "198.51.100.10", "to", "any", "comment", "lite-nas-security-automation-A1B2C3D4E5F6G7H8"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("scan UFW arguments = %v", args)
	}
}

// TestUFWHelperArgumentsRejectsUnknownAction keeps the privileged helper from
// turning a caller-provided label into an arbitrary firewall rule shape.
// Requirements: security-automation/SR-002.
func TestUFWHelperArgumentsRejectsUnknownAction(t *testing.T) {
	_, err := ufwHelperArguments("apply", "A1B2C3D4E5F6G7H8", "physical0", "198.51.100.10", "--all", func(string) error { return nil })
	if err == nil {
		t.Fatal("unknown UFW action accepted")
	}
}

// TestUFWHelperArgumentsRestoresAllTrafficRule checks inverse matching uses
// the same validated source, interface, and ownership comment.
// Requirements: security-automation/SR-002.
func TestUFWHelperArgumentsRestoresAllTrafficRule(t *testing.T) {
	args, err := ufwHelperArguments("restore", "A1B2C3D4E5F6G7H8", "physical0", "198.51.100.10", "block_scan_source_ip", func(string) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"delete", "deny", "in", "on", "physical0", "from", "198.51.100.10", "to", "any", "comment", "lite-nas-security-automation-A1B2C3D4E5F6G7H8"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("scan restore UFW arguments = %v", args)
	}
}

// TestUFWHelperArgumentsRejectsInvalidInputs checks option-shaped and malformed operands.
// Requirements: security-automation/SR-002.
func TestUFWHelperArgumentsRejectsInvalidInputs(t *testing.T) {
	for _, test := range []struct {
		name, operation, id, iface, source string
	}{
		{"unknown operation", "reset", "A1B2C3D4E5F6G7H8", "physical0", "198.51.100.10"},
		{"option-shaped interface", "apply", "A1B2C3D4E5F6G7H8", "--force", "198.51.100.10"},
		{"space in interface", "apply", "A1B2C3D4E5F6G7H8", "physical 0", "198.51.100.10"},
		{"option-shaped source", "apply", "A1B2C3D4E5F6G7H8", "physical0", "--force"},
		{"noncanonical source", "apply", "A1B2C3D4E5F6G7H8", "physical0", "2001:0db8::1"},
		{"multicast source", "apply", "A1B2C3D4E5F6G7H8", "physical0", "224.0.0.1"},
		{"option-shaped incident", "apply", "--force", "physical0", "198.51.100.10"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := ufwHelperArguments(test.operation, test.id, test.iface, test.source, "block_source_ip", func(string) error { return nil })
			if err == nil {
				t.Fatal("invalid UFW operand accepted")
			}
		})
	}
}

// TestUFWHelperArgumentsRequiresVerifiedIngress checks uncertain interfaces skip effects.
// Requirements: security-automation/SR-002.
func TestUFWHelperArgumentsRequiresVerifiedIngress(t *testing.T) {
	_, err := ufwHelperArguments("apply", "A1B2C3D4E5F6G7H8", "virtual0", "198.51.100.10", "block_source_ip", func(string) error { return errors.New("not physical") })
	if err == nil {
		t.Fatal("unverified interface accepted")
	}
}

// TestUFWHelperArgumentsAllowsRecordedMissingInterfaceOnRestore verifies an
// unplugged adapter cannot strand its service-owned UFW deny.
// Requirements: security-automation/SR-002.
func TestUFWHelperArgumentsAllowsRecordedMissingInterfaceOnRestore(t *testing.T) {
	_, err := ufwHelperArguments("restore", "A1B2C3D4E5F6G7H8", "physical0", "198.51.100.10", "block_source_ip", func(string) error { return errors.New("interface missing") })
	if err != nil {
		t.Fatal(err)
	}
}

// TestUFWHelperArgumentsRestoresByExactComment checks deletion cannot follow a
// stale UFW rule number after another owner changes the ruleset.
// Requirements: security-automation/SR-002.
func TestUFWHelperArgumentsRestoresByExactComment(t *testing.T) {
	args, err := ufwHelperArguments("restore", "A1B2C3D4E5F6G7H8", "physical0", "198.51.100.10", "block_source_ip", func(string) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"delete", "deny", "in", "on", "physical0", "proto", "tcp", "from", "198.51.100.10", "to", "any", "port", "443", "comment", "lite-nas-security-automation-A1B2C3D4E5F6G7H8"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("restore UFW arguments = %v", args)
	}
}

// TestFindOwnedUFWRuleAcceptsOnlyMatchingComment checks exact owned-rule selection.
// Requirements: security-automation/SR-002.
func TestFindOwnedUFWRuleAcceptsOnlyMatchingComment(t *testing.T) {
	status := "Status: active\n\n[ 1] 443/tcp on physical0 DENY IN 198.51.100.10 # lite-nas-security-automation-A1B2C3D4E5F6G7H8\n[ 2] 443/tcp ALLOW IN Anywhere\n"
	number, present, err := findOwnedUFWRule(status, "A1B2C3D4E5F6G7H8", "physical0", "198.51.100.10", "block_source_ip")
	if err != nil || !present || number != 1 {
		t.Fatalf("owned UFW rule = %d, %t, %v", number, present, err)
	}
}

// TestFindOwnedUFWRuleRejectsScopeMismatch protects another operator's rule.
// Requirements: security-automation/SR-002.
func TestFindOwnedUFWRuleRejectsScopeMismatch(t *testing.T) {
	status := "Status: active\n\n[ 1] 22/tcp on physical0 DENY IN 198.51.100.10 # lite-nas-security-automation-A1B2C3D4E5F6G7H8\n"
	_, _, err := findOwnedUFWRule(status, "A1B2C3D4E5F6G7H8", "physical0", "198.51.100.10", "block_source_ip")
	if err == nil {
		t.Fatal("mismatched UFW rule accepted")
	}
}

// TestPhysicalDevicePathRejectsVirtualInterfaces checks tunnel-like devices
// cannot be selected as a UFW ingress target.
// Requirements: security-automation/SR-002.
func TestPhysicalDevicePathRejectsVirtualInterfaces(t *testing.T) {
	if physicalDevicePath("/sys/devices/virtual/net/example0/device") {
		t.Fatal("virtual interface accepted as physical")
	}
}

// TestPhysicalDevicePathAcceptsBusBackedInterface checks generic hardware paths.
// Requirements: security-automation/SR-002.
func TestPhysicalDevicePathAcceptsBusBackedInterface(t *testing.T) {
	if !physicalDevicePath("/sys/devices/platform/example/ethernet") {
		t.Fatal("bus-backed interface rejected")
	}
}

// TestHasScopedUFWRuleFindsUnownedDeny protects a matching manual rule.
// Requirements: security-automation/SR-002.
func TestHasScopedUFWRuleFindsUnownedDeny(t *testing.T) {
	status := "Status: active\n[ 7] 443/tcp on physical0 DENY IN 198.51.100.10 # manual-rule\n"
	if !hasScopedUFWRule(status, "physical0", "198.51.100.10", "block_source_ip") {
		t.Fatal("matching manual deny was not found")
	}
}

// TestFindOwnedUFWRuleAcceptsIPv6NumberedStatus verifies the IPv6 marker does
// not hide an owned deny from restoration.
// Requirements: security-automation/SR-002.
func TestFindOwnedUFWRuleAcceptsIPv6NumberedStatus(t *testing.T) {
	status := "Status: active\n[ 1] 443/tcp (v6) on physical0 DENY IN 2001:db8::10 # lite-nas-security-automation-A1B2C3D4E5F6G7H8\n"
	_, present, err := findOwnedUFWRule(status, "A1B2C3D4E5F6G7H8", "physical0", "2001:db8::10", "block_source_ip")
	if err != nil || !present {
		t.Fatalf("IPv6 owned UFW rule = %t, %v", present, err)
	}
}

// TestFindOwnedUFWRuleAcceptsAllTrafficScope verifies the helper recognizes
// only the recorded source-wide deny after UFW renders its numbered status.
// Requirements: security-automation/SR-002.
func TestFindOwnedUFWRuleAcceptsAllTrafficScope(t *testing.T) {
	status := "Status: active\n[ 1] Anywhere on physical0 DENY IN 198.51.100.10 # lite-nas-security-automation-A1B2C3D4E5F6G7H8\n"
	_, present, err := findOwnedUFWRule(status, "A1B2C3D4E5F6G7H8", "physical0", "198.51.100.10", "block_scan_source_ip")
	if err != nil || !present {
		t.Fatalf("all-traffic owned rule = %t, %v", present, err)
	}
}

// TestFindOwnedUFWRuleAcceptsIPv6AllTrafficScope verifies UFW's IPv6 marker
// does not prevent exact source-wide rule ownership checks.
// Requirements: security-automation/SR-002.
func TestFindOwnedUFWRuleAcceptsIPv6AllTrafficScope(t *testing.T) {
	status := "Status: active\n[ 1] Anywhere (v6) on physical0 DENY IN 2001:db8::10 # lite-nas-security-automation-A1B2C3D4E5F6G7H8\n"
	_, present, err := findOwnedUFWRule(status, "A1B2C3D4E5F6G7H8", "physical0", "2001:db8::10", "block_scan_source_ip")
	if err != nil || !present {
		t.Fatalf("IPv6 all-traffic owned rule = %t, %v", present, err)
	}
}
