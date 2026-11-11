package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"

	detection "lite-nas/shared/automationdetection"
	alertcontract "lite-nas/shared/contracts/loggingmanager"
	systemcontract "lite-nas/shared/contracts/systemloggingmanager"
	"lite-nas/shared/loggingmanager"
	"lite-nas/shared/loggingmanager/enum"
	"lite-nas/shared/messaging"
)

// safetyObservation carries one validated create, occurrence, or polled state update.
type safetyObservation struct {
	kind       string
	eventID    string
	alert      alertcontract.AlertPayload
	occurrence alertcontract.AlertOccurrencePayload
	state      string
}

// observedEvent retains fields needed to evaluate conjunctions across NATS payloads.
type observedEvent struct {
	alert     alertcontract.AlertPayload
	valueNum  *float64
	valueText *string
	valueBool *bool
	state     string
	seenAt    time.Time
}

// subscribe validates each incoming payload before it reaches rule matching.
func (runtime *safetyRuntime) subscribe(server messaging.Server, output chan<- safetyObservation) error {
	validate, err := loggingmanager.NewInputValidator()
	if err != nil {
		return err
	}
	if err := server.Subscribe(systemcontract.AlertSubject, func(ctx context.Context, envelope messaging.Envelope) error {
		var alert alertcontract.AlertPayload
		if err := decodeObservation(envelope.Payload, &alert); err != nil {
			return err
		}
		if err := validate.Struct(alert); err != nil {
			return err
		}
		if alert.EventID == "" {
			return errors.New("system alert has no event ID")
		}
		return sendObservation(ctx, output, safetyObservation{kind: "create", eventID: alert.EventID, alert: alert})
	}); err != nil {
		return err
	}
	return server.Subscribe(systemcontract.AlertOccurrenceSubject, func(ctx context.Context, envelope messaging.Envelope) error {
		var occurrence alertcontract.AlertOccurrencePayload
		if err := decodeObservation(envelope.Payload, &occurrence); err != nil {
			return err
		}
		if err := validate.Struct(occurrence); err != nil {
			return err
		}
		if err := validateOccurrenceValue(occurrence); err != nil {
			return err
		}
		return sendObservation(ctx, output, safetyObservation{kind: "occurrence", eventID: occurrence.EventID, occurrence: occurrence})
	})
}

// validateOccurrenceValue checks the typed value before any rule sees it.
func validateOccurrenceValue(value alertcontract.AlertOccurrencePayload) error {
	count := 0
	if value.ValueNum != nil {
		count++
	}
	if value.ValueText != nil {
		count++
		if len(*value.ValueText) > 512 {
			return errors.New("occurrence text exceeds 512 bytes")
		}
	}
	if value.ValueBool != nil {
		count++
	}
	if count == 0 && value.ValueType == "" {
		return nil
	}
	if count != 1 {
		return errors.New("occurrence requires one typed value")
	}
	switch value.ValueType {
	case enum.ValueTypeInt, enum.ValueTypeFloat:
		if value.ValueNum == nil {
			return errors.New("numeric occurrence requires numeric value")
		}
	case enum.ValueTypeText:
		if value.ValueText == nil {
			return errors.New("text occurrence requires text value")
		}
	case enum.ValueTypeBool:
		if value.ValueBool == nil {
			return errors.New("boolean occurrence requires boolean value")
		}
	default:
		return errors.New("unknown occurrence value type")
	}
	return nil
}

// decodeObservation rejects oversized, unknown, and trailing JSON at the NATS boundary.
func decodeObservation(payload []byte, target any) error {
	if len(payload) == 0 || len(payload) > 4096 {
		return errors.New("invalid operational safety observation size")
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return errors.New("trailing observation JSON")
	}
	return nil
}

// sendObservation applies bounded-channel backpressure until callback cancellation.
func sendObservation(ctx context.Context, output chan<- safetyObservation, observation safetyObservation) error {
	select {
	case output <- observation:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// processObservation joins alert fields, runs matching FSMs, and queues every effect.
func (runtime *safetyRuntime) processObservation(ctx context.Context, item safetyObservation) error {
	runtime.mu.Lock()
	if !runtime.relevantEventLocked(item.eventID) {
		runtime.mu.Unlock()
		return nil
	}
	event := runtime.events[item.eventID]
	if event == nil {
		if len(runtime.events) >= 4096 {
			runtime.pruneEventsLocked()
			if len(runtime.events) >= 4096 {
				runtime.mu.Unlock()
				return errors.New("operational safety event cache is full")
			}
		}
		event = &observedEvent{}
		runtime.events[item.eventID] = event
	}
	switch item.kind {
	case "create":
		item.alert.AccessToken = ""
		event.alert = item.alert
	case "occurrence":
		event.valueNum = item.occurrence.ValueNum
		event.valueText = item.occurrence.ValueText
		event.valueBool = item.occurrence.ValueBool
	case "state":
		event.state = item.state
	default:
		runtime.mu.Unlock()
		return errors.New("unknown observation kind")
	}
	event.seenAt = runtime.now()
	snapshot := *event
	runtime.mu.Unlock()
	if item.kind == "state" && item.state == "normal" {
		if err := runtime.normalizeTaskAlert(ctx, item.eventID); err != nil {
			return err
		}
	}
	for _, id := range runtime.order {
		runtime.mu.Lock()
		rule := runtime.rules[id]
		enabled := rule.enabled
		runtime.mu.Unlock()
		if rule.stage != nil {
			if err := runtime.evaluateStage(ctx, id); err != nil {
				return err
			}
			continue
		}
		if !enabled {
			continue
		}
		for step := range rule.spec.Steps {
			key, matched := rule.matches(step, item.kind, item.eventID, snapshot)
			if !matched {
				continue
			}
			if rule.spec.Action.Kind == "cpu_frequency" {
				if snapshot.state == "normal" {
					continue
				}
				if err := runtime.activateTask(ctx, id, key, item.eventID, rule.spec.Steps[step].Level); err != nil {
					return err
				}
				continue
			}
			matches, err := rule.detector.Observe(ctx, detection.Event{ID: item.eventID, Key: key, Step: step, At: runtime.now(), Message: snapshot.alert.Message})
			if err != nil {
				return err
			}
			for _, match := range matches {
				kind := "emit"
				if rule.spec.Action.Kind == "emit_normal" {
					kind = "emit_normal"
				}
				if err := runtime.enqueueEffect(ctx, effectIntent{kind: kind, ruleID: id, key: match.Key}); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// relevantEventLocked rejects unrelated alerts before occupying event cache space.
func (runtime *safetyRuntime) relevantEventLocked(eventID string) bool {
	for id, rule := range runtime.rules {
		if !rule.enabled && runtime.stages[id] == nil {
			continue
		}
		if rule.stage != nil && strings.HasPrefix(eventID, rule.spec.Output.IDPrefix) {
			return true
		}
		for _, step := range rule.spec.Steps {
			if strings.HasPrefix(eventID, step.IDPrefix) {
				return true
			}
		}
	}
	for _, task := range runtime.tasks {
		if _, tracked := task.instance.Data.Active[eventID]; tracked {
			return true
		}
	}
	return false
}

// pruneEventsLocked removes old untracked records while retaining active recovery inputs.
func (runtime *safetyRuntime) pruneEventsLocked() {
	cutoff := time.Now().Add(-time.Hour)
	for id, event := range runtime.events {
		if event.seenAt.After(cutoff) {
			continue
		}
		tracked := false
		for _, task := range runtime.tasks {
			if _, active := task.instance.Data.Active[id]; active {
				tracked = true
				break
			}
		}
		if !tracked {
			delete(runtime.events, id)
		}
	}
	if len(runtime.events) < 4096 {
		return
	}
	oldestID := ""
	var oldestAt time.Time
	for id, event := range runtime.events {
		tracked := false
		for _, task := range runtime.tasks {
			if _, active := task.instance.Data.Active[id]; active {
				tracked = true
				break
			}
		}
		if !tracked && (oldestID == "" || event.seenAt.Before(oldestAt)) {
			oldestID, oldestAt = id, event.seenAt
		}
	}
	if oldestID != "" {
		delete(runtime.events, oldestID)
		runtime.log.Warn("operational safety event cache evicted untracked alert", "event_id", oldestID)
	}
}

// matches applies one rule selector to the assembled alert, occurrence, and state fields.
func (rule *compiledRule) matches(index int, kind, eventID string, event observedEvent) (string, bool) {
	step := rule.spec.Steps[index]
	if !strings.HasPrefix(eventID, step.IDPrefix) || event.alert.EventID == "" {
		return "", false
	}
	if step.Source != "" && step.Source != event.alert.Source || step.Category != "" && step.Category != event.alert.Category || step.State != "" && step.State != event.state {
		return "", false
	}
	message := event.alert.Message
	if step.Message.Equals != "" && step.Message.Equals != message || step.Message.Contains != "" && !strings.Contains(message, step.Message.Contains) {
		return "", false
	}
	var groups []string
	if pattern := rule.patterns[index]; pattern != nil {
		groups = pattern.FindStringSubmatch(message)
		if groups == nil {
			return "", false
		}
	}
	if step.Value.NumGTE != nil && (event.valueNum == nil || *event.valueNum < *step.Value.NumGTE) || step.Value.NumLTE != nil && (event.valueNum == nil || *event.valueNum > *step.Value.NumLTE) || step.Value.TextEquals != "" && (event.valueText == nil || *event.valueText != step.Value.TextEquals) || step.Value.BoolEquals != nil && (event.valueBool == nil || *event.valueBool != *step.Value.BoolEquals) {
		return "", false
	}
	key := ""
	switch rule.spec.CorrelateBy {
	case "event_id":
		key = eventID
	case "source":
		key = event.alert.Source
	case "category":
		key = event.alert.Category
	default:
		if len(groups) > 0 {
			key = groups[rule.patterns[index].SubexpIndex(rule.spec.CorrelateBy)]
		}
	}
	if key == "" || len(key) > 256 || strings.ContainsAny(key, " \t\r\n\x00") {
		return "", false
	}
	return key, true
}

// pollAlertStates reads current lifecycle state through the existing system logging RPC.
func (runtime *safetyRuntime) pollAlertStates(ctx context.Context, output chan<- safetyObservation) {
	runtime.mu.Lock()
	ids := make([]string, 0, len(runtime.events))
	for id := range runtime.events {
		ids = append(ids, id)
	}
	runtime.mu.Unlock()
	if len(ids) == 0 {
		return
	}
	token, err := runtime.serviceToken(ctx)
	if err != nil {
		runtime.log.Error("system state polling token unavailable", "error", err)
		return
	}
	for _, id := range ids {
		requestCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		var response alertcontract.GetAlertResponse
		err := runtime.client.Request(requestCtx, systemcontract.GetAlertRPCSubject, alertcontract.GetAlertInput{AccessToken: token, EventID: id}, &response)
		cancel()
		if err != nil {
			runtime.log.Error("system alert state poll failed", "event_id", id, "error", err)
			continue
		}
		if response.Item == nil || response.Item.EventID != id {
			runtime.log.Error("system alert state poll returned no matching alert", "event_id", id)
			continue
		}
		state := string(response.Item.Status)
		if !validState(state) || state == "" {
			runtime.log.Error("system alert state poll returned invalid state", "event_id", id, "state", state)
			continue
		}
		runtime.mu.Lock()
		cached := runtime.events[id]
		if cached == nil {
			runtime.mu.Unlock()
			continue
		}
		previous := cached.state
		runtime.mu.Unlock()
		if state != previous {
			if err := sendObservation(ctx, output, safetyObservation{kind: "state", eventID: id, state: state}); err != nil {
				runtime.log.Error("system state observation not queued", "event_id", id, "error", err)
				return
			}
		}
	}
}
