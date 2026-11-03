package main

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"time"

	"lite-nas/services/security-automation/detection"
	loggingcontract "lite-nas/shared/contracts/loggingmanager"
	contract "lite-nas/shared/contracts/securityautomation"
	securitycontract "lite-nas/shared/contracts/securityloggingmanager"
	"lite-nas/shared/loggingmanager/enum"
	"lite-nas/shared/messaging"
	"lite-nas/shared/servicetoken"
)

// incidentAudit sends a create or normalize operation for one durable incident.
type incidentAudit func(context.Context, incidentRecord, bool) error

// correlationAlert publishes one completed rule match with its contributing alert IDs.
type correlationAlert func(context.Context, automationRuleSpec, detection.Match) error

// correlationEventID is stable across retries of the same rule and source alerts.
func correlationEventID(ruleID string, alertIDs []string) (string, error) {
	if !automationID.MatchString(ruleID) || len(alertIDs) == 0 || len(alertIDs) > 64 {
		return "", errors.New("invalid correlation identity")
	}
	for _, id := range alertIDs {
		if !eventPrefix.MatchString(id) {
			return "", errors.New("invalid contributing alert ID")
		}
	}
	digest := sha256.Sum256([]byte(ruleID + "\x00" + strings.Join(alertIDs, "\x00")))
	value := binary.BigEndian.Uint64(digest[:8])
	return fmt.Sprintf("t%019d", value%10_000_000_000_000_000_000), nil
}

// newCorrelationAlert binds a report action to the existing Security Logging Manager contract.
func newCorrelationAlert(client messaging.Client, service string) (correlationAlert, error) {
	manager, err := servicetoken.NewManager(client, servicetoken.Options{Service: service})
	if err != nil {
		return nil, err
	}
	return func(ctx context.Context, rule automationRuleSpec, match detection.Match) error {
		id, err := correlationEventID(rule.ID, match.AlertIDs)
		if err != nil {
			return err
		}
		message := fmt.Sprintf("Automation correlation %s for %s; source alerts %s", rule.Name, match.Key, strings.Join(match.AlertIDs, ", "))
		if len(message) > 256 {
			message = fmt.Sprintf("Automation correlation %s; %d source alerts, starting with %s", rule.ID, len(match.AlertIDs), match.AlertIDs[0])
		}
		token, err := serviceAccessToken(ctx, manager)
		if err != nil {
			return err
		}
		return client.Publish(ctx, securitycontract.AlertSubject, loggingcontract.AlertPayload{
			AccessToken: token, EventID: id, Category: "automation_correlation", Severity: enum.Severity(rule.Action.Severity),
			CreatedAt: time.Now().UTC().Format(time.RFC3339), Source: "security-automation", Message: message, Email: rule.Action.Email,
			Mitigate: rule.Action.Mitigate,
		})
	}, nil
}

// auditEventID derives a stable Logging Manager ID from an unguessable incident
// ID, allowing retries after service restart without allocating a new alert.
func auditEventID(incidentID string) (string, error) {
	if !ufwIncidentID.MatchString(incidentID) {
		return "", errors.New("invalid audit incident ID")
	}
	digest := sha256.Sum256([]byte(incidentID))
	value := binary.BigEndian.Uint64(digest[:8])
	return fmt.Sprintf("t%019d", value%10_000_000_000_000_000_000), nil
}

// newIncidentAudit binds the existing service-token and Logging Manager
// contracts. Its caller invokes the returned function from one effect worker.
func newIncidentAudit(client messaging.Client, service string) (incidentAudit, error) {
	manager, err := servicetoken.NewManager(client, servicetoken.Options{Service: service})
	if err != nil {
		return nil, err
	}
	return func(ctx context.Context, record incidentRecord, normalize bool) error {
		token, err := serviceAccessToken(ctx, manager)
		if err != nil {
			return err
		}
		id, err := auditEventID(record.ID)
		if err != nil {
			return err
		}
		if normalize {
			message := "Automation restored its direct HTTPS block for " + record.SourceIP
			if record.ActionName == "block_scan_source_ip" {
				message = "Automation restored its scan-source block for " + record.SourceIP
			}
			if record.ActionName == "usb_block_device" {
				message = "Automation restored the recorded USBGuard allow policy for port " + record.DevicePort
			}
			var response loggingcontract.OKResponse
			if err := client.Request(ctx, securitycontract.UpdateAlertStateRPCSubject,
				loggingcontract.UpdateAlertStateInput{AccessToken: token, EventID: id, Status: enum.StatusNormal, Message: &message}, &response); err != nil {
				return err
			}
			if !response.OK {
				return errors.New("security alert normalization was not acknowledged")
			}
			return nil
		}
		message := fmt.Sprintf("Automation rule %s: direct HTTPS block request for %s; %s", record.RuleID, record.SourceIP, record.Outcome)
		if record.ActionName == "block_scan_source_ip" {
			message = fmt.Sprintf("Automation rule %s: scan-source block request for %s; %s", record.RuleID, record.SourceIP, record.Outcome)
		}
		if record.ActionName == "revoke_web_sessions" {
			message = fmt.Sprintf("Automation rule %s: web-session response for account %s from %s; %s", record.RuleID, record.Username, record.SourceIP, record.Outcome)
		}
		if record.ActionName == "login_terminate_session" {
			message = fmt.Sprintf("Automation rule %s: login session response for account %s tty %s; %s", record.RuleID, record.Username, record.TTY, record.Outcome)
		}
		if record.ActionName == "usb_block_device" {
			message = fmt.Sprintf("Automation rule %s: USBGuard response for port %s product %s; %s", record.RuleID, record.DevicePort, record.DeviceProduct, record.Outcome)
		}
		if len(message) > 256 {
			message = message[:256]
		}
		email := false
		severity := enum.SeverityWarning
		if record.State == contract.StateFailed {
			severity = enum.SeverityError
		}
		return client.Publish(ctx, securitycontract.AlertSubject, loggingcontract.AlertPayload{
			AccessToken: token, EventID: id, Category: "automation_mitigation", Severity: severity,
			CreatedAt: record.StartedAt.UTC().Format(time.RFC3339), Source: "security-automation", Message: message, Email: &email,
			Mitigate: "Review the recorded automation outcome and verify the intended effect. Investigate any failed or unexpected action before taking further steps.",
		})
	}, nil
}

// serviceAccessToken refreshes one worker-owned Auth service credential before
// submitting a Security-role RPC or security alert.
func serviceAccessToken(ctx context.Context, manager *servicetoken.Manager) (string, error) {
	token, expires, tokenErr := manager.Token()
	if tokenErr != nil || !expires.After(time.Now().Add(time.Minute)) {
		if tokenErr != nil || manager.Refresh(ctx) != nil {
			if err := manager.Login(ctx); err != nil {
				return "", err
			}
		}
		returnToken, _, err := manager.Token()
		return returnToken, err
	}
	return token, nil
}
