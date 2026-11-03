package main

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	automationcontract "lite-nas/shared/contracts/securityautomation"
	contract "lite-nas/shared/contracts/usbmanager"
	"lite-nas/shared/messaging"
	"lite-nas/shared/servicetoken"
)

var (
	usbAlertMessage  = regexp.MustCompile(`^Security rule triggered: usb_device_connected device ([0-9]{1,3}-[0-9.]{1,32}) ([0-9a-f]{4}:[0-9a-f]{4})$`)
	usbInventoryLine = regexp.MustCompile(`^([0-9]{1,10}): (allow|block|reject) id ([0-9a-f]{4}:[0-9a-f]{4}) .* hash "([A-Za-z0-9+/=]{44})" parent-hash "[A-Za-z0-9+/=]{44}" via-port "([0-9]{1,3}-[0-9.]{1,32})"(?: |$)`)
	usbNumericID     = regexp.MustCompile(`^[0-9]{1,10}$`)
	usbHash          = regexp.MustCompile(`^[A-Za-z0-9+/=]{44}$`)
	usbAlertPort     = regexp.MustCompile(`^[0-9]{1,3}-[0-9.]{1,32}$`)
	usbProductID     = regexp.MustCompile(`^[0-9a-f]{4}:[0-9a-f]{4}$`)
)

// usbDevice is one validated USBGuard inventory row used for exact identity
// checks before and after a durable policy change.
type usbDevice struct {
	ID      string
	Policy  string
	Product string
	Hash    string
	Port    string
}

// parseUSBInventory accepts USBGuard's bounded list-devices display format.
func parseUSBInventory(raw string) ([]usbDevice, error) {
	if len(raw) > 1024*1024 {
		return nil, errors.New("USBGuard inventory exceeds 1 MiB")
	}
	devices := make([]usbDevice, 0)
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		if line == "" {
			continue
		}
		parts := usbInventoryLine.FindStringSubmatch(line)
		if parts == nil {
			continue // Root hubs and other port formats are outside this action.
		}
		devices = append(devices, usbDevice{ID: parts[1], Policy: parts[2], Product: parts[3], Hash: parts[4], Port: parts[5]})
		if len(devices) > 4096 {
			return nil, errors.New("USBGuard inventory has too many devices")
		}
	}
	return devices, nil
}

// usbEffects binds the existing USB Manager inventory and policy RPCs.
type usbEffects struct {
	list func(context.Context) ([]usbDevice, error)
	set  func(context.Context, string, string) error
}

// newUSBEffects uses a single worker-owned service token for USB Manager calls.
func newUSBEffects(client messaging.Client, service string) (usbEffects, error) {
	manager, err := servicetoken.NewManager(client, servicetoken.Options{Service: service})
	if err != nil {
		return usbEffects{}, err
	}
	return usbEffects{
		list: func(ctx context.Context) ([]usbDevice, error) {
			requestCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()
			token, err := serviceAccessToken(requestCtx, manager)
			if err != nil {
				return nil, err
			}
			var response contract.ListResponse
			if err := client.Request(requestCtx, contract.ListSubject, contract.ListRequest{AccessToken: token}, &response); err != nil {
				return nil, err
			}
			return parseUSBInventory(response.Devices)
		},
		set: func(ctx context.Context, action, id string) error {
			if !usbNumericID.MatchString(id) || action != "block" && action != "allow" {
				return errors.New("invalid USBGuard action operand")
			}
			requestCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()
			token, err := serviceAccessToken(requestCtx, manager)
			if err != nil {
				return err
			}
			subject := contract.BlockSubject
			if action == "allow" {
				subject = contract.AllowSubject
			}
			var response contract.ActionResponse
			if err := client.Request(requestCtx, subject, contract.DeviceRequest{AccessToken: token, DeviceID: id, Permanent: true}, &response); err != nil {
				return err
			}
			if !response.Completed {
				return errors.New("USBGuard policy change not acknowledged")
			}
			return nil
		},
	}, nil
}

// exactUSBDevice finds one current device with the recorded port, product,
// and hash, even when USBGuard assigned a different numeric ID after restart.
func exactUSBDevice(devices []usbDevice, record incidentRecord) (usbDevice, bool, error) {
	var matched usbDevice
	for _, device := range devices {
		if device.Port != record.DevicePort || device.Product != record.DeviceProduct || record.DeviceHash != "" && device.Hash != record.DeviceHash {
			continue
		}
		if matched.ID != "" {
			return usbDevice{}, false, errors.New("ambiguous USBGuard device identity")
		}
		matched = device
	}
	return matched, matched.ID != "", nil
}

// applyUSBIncident records USBGuard's stable device hash before changing its
// permanent policy. A verified block remains active until manual rollback.
func (runtime *automationRuntime) applyUSBIncident(ctx context.Context, id string) {
	record, ok := runtime.incidents.get(id)
	if !ok || record.ActionName != "usb_block_device" || record.State == automationcontract.StateRestoring {
		return
	}
	if record.State == automationcontract.StateMitigating {
		if !record.AlertPublished {
			_ = runtime.publishIncidentAlert(ctx, id)
		}
		return
	}
	if record.DeviceHash == "" && time.Since(record.StartedAt) > 10*time.Minute {
		record.State = automationcontract.StateMitigating
		record.Outcome = "USB connection evidence expired before device resolution; no block applied"
		if err := runtime.incidents.save(record); err == nil {
			if runtime.publishIncidentAlert(ctx, id) == nil {
				_ = runtime.incidents.delete(id)
			}
		}
		return
	}
	devices, err := runtime.usb.list(ctx)
	if err != nil {
		runtime.failUSBIncident(record, err)
		return
	}
	device, found, err := exactUSBDevice(devices, record)
	if err != nil || !found || record.DeviceHash == "" && device.Policy != "allow" {
		if record.DeviceHash != "" {
			runtime.failUSBIncident(record, errors.New("recorded USBGuard device is absent or ambiguous"))
			return
		}
		record.State = automationcontract.StateMitigating
		record.Outcome = "no unique currently allowed USBGuard device; no block applied"
		if saveErr := runtime.incidents.save(record); saveErr != nil {
			runtime.log.Error("automation USB incident update failed", "incident", id, "error", saveErr)
			return
		}
		if err := runtime.publishIncidentAlert(ctx, id); err == nil {
			_ = runtime.incidents.delete(id)
		}
		return
	}
	if record.DeviceHash == "" {
		record.DeviceHash = device.Hash
		record.DeviceID = device.ID
		if err := runtime.incidents.save(record); err != nil {
			runtime.log.Error("automation USB identity journal failed", "incident", id, "error", err)
			return
		}
	}
	if device.Policy == "allow" {
		if err := runtime.usb.set(ctx, "block", device.ID); err != nil {
			runtime.failUSBIncident(record, err)
			return
		}
		devices, err = runtime.usb.list(ctx)
		if err != nil {
			runtime.failUSBIncident(record, err)
			return
		}
		device, found, err = exactUSBDevice(devices, record)
	}
	if err != nil || !found || device.Policy != "block" {
		runtime.failUSBIncident(record, errors.New("USBGuard block could not be verified for recorded device"))
		return
	}
	record.State = automationcontract.StateMitigating
	record.Outcome = fmt.Sprintf("USBGuard permanently blocked device %s on port %s", device.ID, record.DevicePort)
	if err := runtime.incidents.save(record); err != nil {
		runtime.log.Error("automation USB incident update failed", "incident", id, "error", err)
		return
	}
	runtime.setRuleOutcome(record.RuleID, record.Outcome)
	if err := runtime.publishIncidentAlert(ctx, id); err != nil {
		runtime.log.Error("automation USB incident alert pending", "incident", id, "error", err)
	}
}

// failUSBIncident retains the exact recorded target for an operator to review
// and for bounded retries when the device is still within its trigger window.
func (runtime *automationRuntime) failUSBIncident(record incidentRecord, effectErr error) {
	record.State = automationcontract.StateFailed
	record.Outcome = "USBGuard action failed; recorded device requires review"
	if err := runtime.incidents.save(record); err != nil {
		runtime.log.Error("automation USB incident update failed", "incident", record.ID, "error", err)
	}
	runtime.log.Error("automation USB action failed", "incident", record.ID, "error", effectErr)
}

// restoreUSBPolicy restores only the recorded device's prior allow policy and
// requires a matching current hash, port, and vendor/product identity.
func (runtime *automationRuntime) restoreUSBPolicy(ctx context.Context, record incidentRecord) error {
	if record.DeviceHash == "" {
		return nil // No block was applied before a manual rollback request.
	}
	devices, err := runtime.usb.list(ctx)
	if err != nil {
		return err
	}
	device, found, err := exactUSBDevice(devices, record)
	if err != nil || !found {
		return errors.New("recorded USBGuard device is not present for verified restoration")
	}
	if device.Policy == "allow" {
		return nil
	}
	if device.Policy != "block" {
		return errors.New("recorded USBGuard device has an unexpected policy")
	}
	if err := runtime.usb.set(ctx, "allow", device.ID); err != nil {
		return err
	}
	devices, err = runtime.usb.list(ctx)
	if err != nil {
		return err
	}
	device, found, err = exactUSBDevice(devices, record)
	if err != nil || !found || device.Policy != "allow" {
		return errors.New("USBGuard restoration could not be verified")
	}
	return nil
}
