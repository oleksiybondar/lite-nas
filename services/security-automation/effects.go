package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"time"

	authcontract "lite-nas/shared/contracts/auth"
	contract "lite-nas/shared/contracts/securityautomation"
	"lite-nas/shared/workers"
)

const automationExecutable = "/usr/libexec/lite-nas/security-automation"

var (
	verifiedWebAlert  = regexp.MustCompile(`^Security rule triggered: pam_web_api_password_guessing via Web/API for account ([A-Za-z0-9_][A-Za-z0-9_.-]{0,127}) from ([0-9A-Fa-f:.]+)$`)
	verifiedSudoAlert = regexp.MustCompile(`^Security rule triggered: sudo_policy_violation for account ([A-Za-z0-9_][A-Za-z0-9_.-]{0,127}) tty ([A-Za-z0-9_./-]{1,64})$`)
)

// effectCommand is one accepted operator action consumed by the sole effect worker.
type effectCommand struct {
	kind string
	id   string
}

// queueWebRevocation journals a verified, account-and-source keyed Log
// Analysis match before the worker requests one-way Auth revocation.
func (runtime *automationRuntime) queueWebRevocation(intent detectionIntent) error {
	if runtime.incidents == nil || runtime.operatorActions == nil || intent.RuleID != "web_password_guessing" || len(intent.Match.AlertIDs) != 1 || !strings.HasPrefix(intent.Match.AlertIDs[0], "secweb") || intent.Match.Key != intent.Match.AlertIDs[0] {
		return errors.New("web revocation requires one durable matched alert")
	}
	parts := verifiedWebAlert.FindStringSubmatch(intent.Match.LastMessage)
	if len(parts) != 3 || !incidentUsername.MatchString(parts[1]) {
		return errors.New("web alert lacks verified account and source")
	}
	if _, err := validateUFWSource(parts[2]); err != nil {
		return err
	}
	id, err := newIncidentID()
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	record := incidentRecord{
		ID: id, RuleID: intent.RuleID, ActionName: "revoke_web_sessions", Username: parts[1], SourceIP: parts[2],
		StartedAt: now, ExpiresAt: now.Add(5 * time.Minute), State: contract.StatePending, Outcome: "verified web-session revocation queued",
	}
	if err := runtime.incidents.save(record); err != nil {
		return err
	}
	select {
	case runtime.operatorActions <- effectCommand{kind: "web", id: id}:
		return nil
	default:
		return errors.New("effect queue full; durable web revocation will be retried")
	}
}

// queueAutomaticUFW accepts only the dedicated direct-HTTPS Log Analysis
// alert; SSH and generic network alerts lack this ingress provenance.
func (runtime *automationRuntime) queueAutomaticUFW(intent detectionIntent) error {
	if intent.RuleID != "web_https_source_block" || len(intent.Match.AlertIDs) != 1 || !strings.HasPrefix(intent.Match.AlertIDs[0], "secweb") {
		return errors.New("automatic UFW block lacks one direct HTTPS alert")
	}
	parts := verifiedWebAlert.FindStringSubmatch(intent.Match.LastMessage)
	if len(parts) != 3 || intent.Match.Key != parts[2] {
		return errors.New("automatic UFW source differs from verified alert")
	}
	_, err := runtime.queueSourceBlock(context.Background(), intent.RuleID, parts[2], "block_source_ip", "automatic web alert")
	return err
}

// queueAutomaticScanUFW accepts only an ordered pair of Network Security scan
// alerts that the configured sequence correlated to one canonical source IP.
func (runtime *automationRuntime) queueAutomaticScanUFW(intent detectionIntent) error {
	if intent.RuleID != "network_scan_sequence" || len(intent.Match.AlertIDs) != 2 || intent.Match.AlertIDs[0] == intent.Match.AlertIDs[1] ||
		!strings.HasPrefix(intent.Match.AlertIDs[0], "secnet") || !strings.HasPrefix(intent.Match.AlertIDs[1], "secnet") {
		return errors.New("scan block requires two distinct network alerts")
	}
	if _, err := validateUFWSource(intent.Match.Key); err != nil {
		return err
	}
	if !strings.HasPrefix(intent.Match.LastMessage, "Blocked TCP port scan from "+intent.Match.Key+" to ") {
		return errors.New("vertical scan alert differs from correlated source")
	}
	_, err := runtime.queueSourceBlock(context.Background(), intent.RuleID, intent.Match.Key, "block_scan_source_ip", "automatic scan sequence")
	return err
}

// queueSessionTermination journals a count of three sudo denials from the
// same account and terminal before resolving a unique loginctl session.
func (runtime *automationRuntime) queueSessionTermination(intent detectionIntent) error {
	if runtime.incidents == nil || runtime.operatorActions == nil || intent.RuleID != "repeated_sudo_denial" || len(intent.Match.AlertIDs) != 3 {
		return errors.New("session termination requires three matched sudo alerts")
	}
	parts := verifiedSudoAlert.FindStringSubmatch(intent.Match.LastMessage)
	if len(parts) != 3 || !incidentUsername.MatchString(parts[1]) || !verifiedTTY.MatchString(parts[2]) || intent.Match.Key != fmt.Sprintf("%d:%s%s", len(parts[1]), parts[1], parts[2]) {
		return errors.New("sudo alerts lack one verified account and terminal")
	}
	for _, alertID := range intent.Match.AlertIDs {
		if !strings.HasPrefix(alertID, "secpriv") {
			return errors.New("sudo alert has unexpected event ID")
		}
	}
	id, err := newIncidentID()
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	record := incidentRecord{
		ID: id, RuleID: intent.RuleID, ActionName: "login_terminate_session", Username: parts[1], TTY: parts[2],
		StartedAt: now, ExpiresAt: now.Add(5 * time.Minute), State: contract.StatePending, Outcome: "verified login-session termination queued",
	}
	if err := runtime.incidents.save(record); err != nil {
		return err
	}
	select {
	case runtime.operatorActions <- effectCommand{kind: "session", id: id}:
		return nil
	default:
		return errors.New("effect queue full; durable session termination will be retried")
	}
}

// queueUSBBlock records two distinct kernel device alerts for the same port
// and vendor/product before USBGuard inventory resolves the current device.
func (runtime *automationRuntime) queueUSBBlock(intent detectionIntent) error {
	if runtime.incidents == nil || runtime.operatorActions == nil || intent.RuleID != "usb_suspicious_activity" || len(intent.Match.AlertIDs) != 2 {
		return errors.New("USB blocking requires two matched connection alerts")
	}
	parts := usbAlertMessage.FindStringSubmatch(intent.Match.LastMessage)
	if len(parts) != 3 || intent.Match.Key != parts[1]+"|"+parts[2] || intent.Match.AlertIDs[0] == intent.Match.AlertIDs[1] {
		return errors.New("USB alerts lack one verified device identity")
	}
	for _, alertID := range intent.Match.AlertIDs {
		if !strings.HasPrefix(alertID, "secusb") {
			return errors.New("USB alert has unexpected event ID")
		}
	}
	id, err := newIncidentID()
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	record := incidentRecord{
		ID: id, RuleID: intent.RuleID, ActionName: "usb_block_device", DevicePort: parts[1], DeviceProduct: parts[2],
		StartedAt: now, ExpiresAt: now.AddDate(10, 0, 0), State: contract.StatePending, Outcome: "USBGuard device resolution queued",
	}
	if err := runtime.incidents.save(record); err != nil {
		return err
	}
	select {
	case runtime.operatorActions <- effectCommand{kind: "usb", id: id}:
		return nil
	default:
		return errors.New("effect queue full; durable USB block will be retried")
	}
}

// revokeMatchedWebSessions asks Auth to revoke only active Admin Panel
// sessions with the exact account and direct HTTPS client address.
func (runtime *automationRuntime) revokeMatchedWebSessions(ctx context.Context, username, source string) (int, error) {
	requestCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	var response authcontract.RevokeMatchedSessionsResponse
	if err := runtime.client.Request(requestCtx, authcontract.RevokeMatchedSessionsRPCSubject,
		authcontract.RevokeMatchedSessionsRequest{Login: username, ClientIP: source}, &response); err != nil {
		return 0, err
	}
	if response.Revoked < 0 || response.Revoked > 4096 {
		return 0, errors.New("invalid Auth revocation count")
	}
	return response.Revoked, nil
}

// applyWebIncident retries an idempotent, one-way Auth request and retains a
// failed record until its alert can be published for operator inspection.
func (runtime *automationRuntime) applyOneWayIncident(ctx context.Context, id string) {
	record, ok := runtime.incidents.get(id)
	if !ok || record.ActionName != "revoke_web_sessions" && record.ActionName != "login_terminate_session" {
		return
	}
	if record.State == contract.StatePending || record.State == contract.StateFailed {
		outcome := ""
		var err error
		if !time.Now().Before(record.ExpiresAt) {
			outcome = "verification window expired before session action; no effect"
		} else if record.ActionName == "revoke_web_sessions" {
			var count int
			count, err = runtime.runWebRevocation(ctx, record.Username, record.SourceIP)
			outcome = fmt.Sprintf("Auth revoked %d matching web sessions", count)
		} else {
			var sessionID string
			var matched bool
			sessionID, matched, err = runtime.terminateSession(ctx, record.Username, record.TTY)
			outcome = "no unique active account-and-TTY session; no termination"
			if matched {
				outcome = "loginctl terminated verified session " + sessionID
			}
		}
		if err != nil {
			record.State = contract.StateFailed
			record.Outcome = "one-way session action failed; retry pending"
			if saveErr := runtime.incidents.save(record); saveErr != nil {
				runtime.log.Error("automation web incident update failed", "incident", id, "error", saveErr)
			}
			runtime.log.Error("automation session action failed", "incident", id, "error", err)
			return
		}
		record.State = contract.StateMitigating
		record.Outcome = outcome
		if err := runtime.incidents.save(record); err != nil {
			runtime.log.Error("automation web incident update failed", "incident", id, "error", err)
			return
		}
		runtime.setRuleOutcome(record.RuleID, record.Outcome)
	}
	if err := runtime.publishIncidentAlert(ctx, id); err != nil {
		runtime.log.Error("automation web incident alert pending", "incident", id, "error", err)
		return
	}
	if err := runtime.incidents.delete(id); err != nil {
		runtime.log.Error("automation web incident cleanup failed", "incident", id, "error", err)
	}
}

// discoverPhysicalIngressInterfaces returns sysfs-backed physical interfaces,
// including currently down devices so their UFW rules remain effective if up.
// The privileged helper independently rechecks every interface before UFW.
func discoverPhysicalIngressInterfaces() ([]string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var names []string
	for _, iface := range interfaces {
		if physicalIngressInterface(iface.Name) == nil {
			names = append(names, iface.Name)
		}
	}
	if len(names) == 0 || len(names) > 16 {
		return nil, errors.New("no bounded, verified physical ingress interface set")
	}
	sort.Strings(names)
	return names, nil
}

// invokePrivilegedUFW asks sudo to run only the root-validated helper form.
func invokePrivilegedUFW(ctx context.Context, operation, incidentID, iface, source, actionName string) error {
	if _, err := ufwHelperArguments(operation, incidentID, iface, source, actionName, physicalIngressInterface); err != nil {
		return err
	}
	commandCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	output, err := exec.CommandContext(commandCtx, "/usr/bin/sudo", "-n", automationExecutable,
		"ufw-helper", operation, incidentID, iface, source, actionName).CombinedOutput()
	if err != nil {
		return fmt.Errorf("privileged UFW helper failed: %w: %s", err, output)
	}
	return nil
}

// applyIncident verifies each physical-interface rule before reporting mitigation.
// A failure leaves the durable record visible for explicit or timed restoration.
func (runtime *automationRuntime) applyIncident(ctx context.Context, id string) {
	record, ok := runtime.incidents.get(id)
	if !ok || record.State != contract.StatePending {
		return
	}
	for _, iface := range record.Interfaces {
		if current, exists := runtime.incidents.get(id); !exists || current.State == contract.StateRestoring {
			return // A queued rollback owns any already applied rule.
		}
		if err := runtime.runUFW(ctx, "apply", record.ID, iface, record.SourceIP, record.ActionName); err != nil {
			record.State = contract.StateFailed
			record.Outcome = fmt.Sprintf("UFW apply failed on %s; earlier interface blocks may remain until restore", iface)
			updated, saveErr := runtime.incidents.transition(record, contract.StatePending)
			if saveErr != nil {
				runtime.log.Error("automation incident update failed", "incident", id, "error", saveErr)
			}
			if updated {
				runtime.setRuleOutcome(record.RuleID, record.Outcome)
			}
			runtime.log.Error("automation mitigation failed", "incident", id, "error", err)
			if auditErr := runtime.publishIncidentAlert(ctx, id); auditErr != nil {
				runtime.log.Error("automation incident alert pending", "incident", id, "error", auditErr)
			}
			return
		}
	}
	record.State = contract.StateMitigating
	record.Outcome = "direct HTTPS source block applied and verified"
	if record.ActionName == "block_scan_source_ip" {
		record.Outcome = "scan source blocked on physical interfaces and verified"
	}
	updated, err := runtime.incidents.transition(record, contract.StatePending)
	if err != nil {
		runtime.log.Error("automation mitigation state update failed", "incident", id, "error", err)
		return
	}
	if !updated {
		return // An operator restore has taken ownership of this incident.
	}
	runtime.setRuleOutcome(record.RuleID, record.Outcome)
	runtime.log.Warn("automation mitigation applied", "incident", id, "rule", record.RuleID, "source", record.SourceIP, "expires", record.ExpiresAt)
	if err := runtime.publishIncidentAlert(ctx, id); err != nil {
		runtime.log.Error("automation incident alert pending", "incident", id, "error", err)
	}
}

// restoreIncident removes only the recorded UFW rules, including partial work
// left by a failed or interrupted apply. It leaves failures visible for retry.
func (runtime *automationRuntime) restoreIncident(ctx context.Context, id string) error {
	record, ok := runtime.incidents.get(id)
	if !ok {
		return errors.New("unknown automation incident")
	}
	if record.ActionName == "revoke_web_sessions" || record.ActionName == "login_terminate_session" {
		return errors.New("session action is one-way and has no restore operation")
	}
	record.State = contract.StateRestoring
	record.Outcome = "restoring owned UFW rules"
	if record.ActionName == "usb_block_device" {
		record.Outcome = "restoring recorded USBGuard allow policy"
	}
	if err := runtime.incidents.save(record); err != nil {
		return err
	}
	if record.ActionName == "usb_block_device" {
		if err := runtime.restoreUSBPolicy(ctx, record); err != nil {
			record.State = contract.StateRestoring
			record.Outcome = "USBGuard manual restoration failed; recorded block may remain"
			_ = runtime.incidents.save(record)
			return err
		}
	} else {
		for _, iface := range record.Interfaces {
			if err := runtime.runUFW(ctx, "restore", record.ID, iface, record.SourceIP, record.ActionName); err != nil {
				record.State = contract.StateFailed
				record.Outcome = fmt.Sprintf("UFW restore failed on %s; an owned block may remain", iface)
				runtime.setRuleOutcome(record.RuleID, record.Outcome)
				if saveErr := runtime.incidents.save(record); saveErr != nil {
					runtime.log.Error("automation incident update failed", "incident", id, "error", saveErr)
				}
				return err
			}
		}
	}
	if err := runtime.publishIncidentAlert(ctx, id); err != nil {
		return err
	}
	if runtime.audit != nil {
		if err := runtime.audit(ctx, record, true); err != nil {
			if current, exists := runtime.incidents.get(id); exists {
				current.AlertPublished = false // Retry create if normalization raced delivery.
				current.Outcome = "UFW restored; security alert normalization pending"
				if current.ActionName == "usb_block_device" {
					current.Outcome = "USBGuard allow policy restored; security alert normalization pending"
				}
				if _, saveErr := runtime.incidents.transition(current, current.State); saveErr != nil {
					runtime.log.Error("automation alert retry state update failed", "incident", id, "error", saveErr)
				}
			}
			return err
		}
	}
	if err := runtime.incidents.delete(id); err != nil {
		return err
	}
	runtime.log.Warn("automation mitigation restored", "incident", id, "rule", record.RuleID, "source", record.SourceIP)
	if record.ActionName == "usb_block_device" {
		runtime.setRuleOutcome(record.RuleID, "recorded USBGuard allow policy restored and verified")
	} else {
		outcome := "direct HTTPS source block restored and verified"
		if record.ActionName == "block_scan_source_ip" {
			outcome = "scan source block restored and verified"
		}
		runtime.setRuleOutcome(record.RuleID, outcome)
	}
	return nil
}

// publishIncidentAlert retries a stable event ID until the Logging Manager
// create message has been accepted and that fact is journaled.
func (runtime *automationRuntime) publishIncidentAlert(ctx context.Context, id string) error {
	if runtime.audit == nil {
		return nil
	}
	record, ok := runtime.incidents.get(id)
	if !ok || record.AlertPublished {
		return nil
	}
	if err := runtime.audit(ctx, record, false); err != nil {
		if record.ActionName == "" || record.ActionName == "block_source_ip" || record.ActionName == "block_scan_source_ip" {
			if record.State == contract.StateMitigating || record.State == contract.StateRestoring {
				if record.State == contract.StateMitigating {
					record.Outcome = "direct HTTPS block applied; security alert publication pending"
					if record.ActionName == "block_scan_source_ip" {
						record.Outcome = "scan source block applied; security alert publication pending"
					}
				} else {
					record.Outcome = "UFW restored; security alert publication pending"
				}
				_, _ = runtime.incidents.transition(record, record.State)
			}
		}
		return err
	}
	for {
		current, exists := runtime.incidents.get(id)
		if !exists || current.AlertPublished {
			return nil
		}
		current.AlertPublished = true
		if current.State == contract.StateMitigating && (current.ActionName == "" || current.ActionName == "block_source_ip" || current.ActionName == "block_scan_source_ip") {
			current.Outcome = "direct HTTPS source block applied and verified"
			if current.ActionName == "block_scan_source_ip" {
				current.Outcome = "scan source blocked on physical interfaces and verified"
			}
		}
		updated, err := runtime.incidents.transition(current, current.State)
		if err != nil {
			return err
		}
		if updated {
			return nil
		}
	}
}

// setRuleOutcome updates the operator view after one owned effect transition.
func (runtime *automationRuntime) setRuleOutcome(ruleID, outcome string) {
	runtime.mutex.Lock()
	defer runtime.mutex.Unlock()
	if rule := runtime.rules[ruleID]; rule != nil {
		rule.view.LastOutcome = outcome
	}
}

// startOperatorEffectsWorker serializes accepted operator effects and timed
// restoration in the service process. The RPC adapter owns and closes its
// bounded input only after NATS Drain has stopped callbacks.
func startOperatorEffectsWorker(ctx context.Context, runtime *automationRuntime) <-chan struct{} {
	done := make(chan struct{})
	workers.NewBaseWorker(func(workerCtx context.Context) {
		defer close(done)
		for _, record := range runtime.incidents.list() {
			if record.ActionName == "revoke_web_sessions" || record.ActionName == "login_terminate_session" {
				runtime.applyOneWayIncident(context.Background(), record.ID)
				continue
			}
			if record.ActionName == "usb_block_device" {
				if record.State == contract.StatePending || record.State == contract.StateFailed {
					runtime.applyUSBIncident(context.Background(), record.ID)
				} else if record.State == contract.StateRestoring {
					if err := runtime.restoreIncident(context.Background(), record.ID); err != nil {
						runtime.log.Error("automation USB restoration failed", "incident", record.ID, "error", err)
					}
				}
				continue
			}
			if record.State == contract.StatePending || record.State == contract.StateRestoring || !time.Now().Before(record.ExpiresAt) {
				if err := runtime.restoreIncident(context.Background(), record.ID); err != nil {
					runtime.log.Error("automation startup restoration failed", "incident", record.ID, "error", err)
				}
			}
		}
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		shutdown := workerCtx.Done()
		input := runtime.operatorActions
		for input != nil {
			select {
			case action, ok := <-input:
				if !ok {
					input = nil
					continue
				}
				switch action.kind {
				case "restore":
					if err := runtime.restoreIncident(context.Background(), action.id); err != nil {
						runtime.log.Error("automation operator restoration failed", "incident", action.id, "error", err)
					}
				case "web", "session":
					runtime.applyOneWayIncident(context.Background(), action.id)
				case "usb":
					runtime.applyUSBIncident(context.Background(), action.id)
				case "apply":
					runtime.applyIncident(context.Background(), action.id)
				default:
					runtime.log.Error("unknown automation effect command", "kind", action.kind, "incident", action.id)
				}
			case <-ticker.C:
				for _, record := range runtime.incidents.list() {
					if record.ActionName == "revoke_web_sessions" || record.ActionName == "login_terminate_session" {
						runtime.applyOneWayIncident(context.Background(), record.ID)
						continue
					}
					if record.ActionName == "usb_block_device" {
						if record.State == contract.StatePending || record.State == contract.StateFailed && time.Since(record.StartedAt) < 10*time.Minute {
							runtime.applyUSBIncident(context.Background(), record.ID)
						} else if record.State == contract.StateRestoring {
							if err := runtime.restoreIncident(context.Background(), record.ID); err != nil {
								runtime.log.Error("automation USB restoration retry failed", "incident", record.ID, "error", err)
							}
						} else if !record.AlertPublished {
							_ = runtime.publishIncidentAlert(context.Background(), record.ID)
						}
						continue
					}
					if record.State == contract.StateRestoring || !time.Now().Before(record.ExpiresAt) {
						if err := runtime.restoreIncident(context.Background(), record.ID); err != nil {
							runtime.log.Error("automation timed restoration failed", "incident", record.ID, "error", err)
						}
					} else if !record.AlertPublished {
						if err := runtime.publishIncidentAlert(context.Background(), record.ID); err != nil {
							runtime.log.Error("automation incident alert retry failed", "incident", record.ID, "error", err)
						}
					}
				}
			case <-shutdown:
				shutdown = nil
				ticker.Stop()
			}
		}
	}).Start(ctx)
	return done
}
