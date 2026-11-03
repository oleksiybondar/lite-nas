package main

import (
	"context"
	"testing"

	"lite-nas/services/security-automation/detection"
	"lite-nas/shared/logger"
)

const syntheticUSBInventory = "7: allow id 1234:abcd serial \"redacted\" name \"Synthetic device\" hash \"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=\" parent-hash \"BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB=\" via-port \"3-2\" with-interface 03:00:00"

// TestParseUSBInventoryReadsScopedDevice verifies USBGuard's numeric ID and
// stable identity fields are parsed without storing production inventory.
func TestParseUSBInventoryReadsScopedDevice(t *testing.T) {
	devices, err := parseUSBInventory(syntheticUSBInventory)
	if err != nil || len(devices) != 1 || devices[0].ID != "7" || devices[0].Port != "3-2" {
		t.Fatalf("USBGuard inventory = %+v, %v", devices, err)
	}
}

// TestUSBBlockUsesRecordedIdentity verifies an automatic block is retained
// for manual restoration only after USBGuard reports the same device blocked.
func TestUSBBlockUsesRecordedIdentity(t *testing.T) {
	store := mustIncidentStore(t)
	current := "allow"
	runtime := &automationRuntime{
		incidents: store, operatorActions: make(chan effectCommand, 1), log: logger.NopLogger{},
		usb: usbEffects{
			list: func(context.Context) ([]usbDevice, error) {
				return []usbDevice{{ID: "7", Policy: current, Product: "1234:abcd", Hash: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=", Port: "3-2"}}, nil
			},
			set: func(_ context.Context, action, _ string) error { current = "block"; return nil },
		},
		audit: func(context.Context, incidentRecord, bool) error { return nil },
	}
	intent := detectionIntent{RuleID: "usb_suspicious_activity", Match: detection.Match{Key: "3-2|1234:abcd", AlertIDs: []string{"secusb1", "secusb2"}, LastMessage: "Security rule triggered: usb_device_connected device 3-2 1234:abcd"}}
	if err := runtime.queueUSBBlock(intent); err != nil {
		t.Fatal(err)
	}
	command := <-runtime.operatorActions
	runtime.applyUSBIncident(context.Background(), command.id)
	if got := store.list(); len(got) != 1 || got[0].DeviceHash == "" || got[0].State != "mitigating" {
		t.Fatalf("USBGuard block journal = %+v", got)
	}
}
