package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"lite-nas/services/security-automation/detection"
	"lite-nas/shared/logger"
	"lite-nas/shared/workers"
)

// startAutomationListener owns the alert-to-observation and timer-to-dispatcher channels.
// The NATS adapter owns alerts and closes it after drain. The timer owns timerTicks.
// This listener alone closes its bounded outputs after accepted alerts have drained.
func startAutomationListener(ctx context.Context, alerts <-chan alertObservation, timerTicks <-chan struct{},
	rules map[string]*compiledAutomationRule, log logger.Logger,
) (<-chan observation, <-chan struct{}, <-chan struct{}) {
	observations := make(chan observation, 128)
	deadlines := make(chan struct{}, 1)
	done := make(chan struct{})
	workers.NewBaseWorker(func(workerCtx context.Context) {
		defer close(done)
		defer close(observations)
		defer close(deadlines)
		shutdown := workerCtx.Done()
		for alerts != nil {
			select {
			case alert, ok := <-alerts:
				if !ok {
					alerts = nil
					continue
				}
				for id, rule := range rules {
					if !rule.enabled.Load() || rule.plugin == nil {
						continue
					}
					for step := range rule.spec.Steps {
						key, matched := rule.matches(step, alert)
						if !matched {
							continue
						}
						at := time.Now()
						if alert.Alert.CreatedAt != "" {
							if parsed, err := time.Parse(time.RFC3339, alert.Alert.CreatedAt); err == nil {
								at = parsed
							}
						}
						item := observation{RuleID: id, Event: detection.Event{ID: alert.Alert.EventID, Key: key, Step: step, At: at, Message: alert.Alert.Message}}
						select {
						case observations <- item:
						case <-workerCtx.Done():
							log.Error("automation observation not queued at shutdown", "rule", id, "event_id", item.Event.ID)
						}
					}
				}
			case _, ok := <-timerTicks:
				if !ok {
					timerTicks = nil
					continue
				}
				select {
				case deadlines <- struct{}{}:
				case <-workerCtx.Done():
					log.Warn("automation deadline tick not queued at shutdown")
				}
			case <-shutdown:
				shutdown = nil
				timerTicks = nil
			}
		}
	}).Start(ctx)
	return observations, deadlines, done
}

// matches applies one precompiled selector and returns its bounded correlation key.
func (rule *compiledAutomationRule) matches(step int, input alertObservation) (string, bool) {
	selector := rule.spec.Steps[step]
	alert := input.Alert
	if selector.Type != input.Kind || !strings.HasPrefix(alert.EventID, selector.IDPrefix) || selector.Category != "" && selector.Category != alert.Category || selector.Source != "" && selector.Source != alert.Source || selector.Severity != "" && selector.Severity != string(alert.Severity) {
		return "", false
	}
	if selector.Message.Equals != "" && selector.Message.Equals != alert.Message || selector.Message.Contains != "" && !strings.Contains(alert.Message, selector.Message.Contains) {
		return "", false
	}
	key := ""
	if pattern := rule.regexes[step]; pattern != nil {
		groups := pattern.FindStringSubmatch(alert.Message)
		if groups == nil {
			return "", false
		}
		if index := pattern.SubexpIndex(rule.spec.CorrelateBy); index >= 0 {
			key = groups[index]
		}
		if rule.spec.CorrelateBy == "account_tty" {
			username := groups[pattern.SubexpIndex("username")]
			tty := groups[pattern.SubexpIndex("tty")]
			if !incidentUsername.MatchString(username) || len(tty) == 0 || len(tty) > 64 {
				return "", false
			}
			key = fmt.Sprintf("%d:%s%s", len(username), username, tty)
		}
		if rule.spec.CorrelateBy == "device_identity" {
			port := groups[pattern.SubexpIndex("device_port")]
			product := groups[pattern.SubexpIndex("device_product")]
			if !usbAlertMessage.MatchString(alert.Message) {
				return "", false
			}
			key = port + "|" + product
		}
	}
	if key == "" {
		switch rule.spec.CorrelateBy {
		case "event_id":
			key = alert.EventID
		case "source":
			key = alert.Source
		case "category":
			key = alert.Category
		}
	}
	if key == "" || len(key) > 256 || strings.HasPrefix(key, "-") || strings.ContainsAny(key, " \t\r\n\x00") {
		return "", false
	}
	return key, true
}
