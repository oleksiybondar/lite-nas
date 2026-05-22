package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	servicerules "lite-nas/services/resources-monitor/rules"
	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	systemloggingmanagercontract "lite-nas/shared/contracts/systemloggingmanager"
	"lite-nas/shared/eventmanager"
	sharedlogger "lite-nas/shared/logger"
	sharedloggingenum "lite-nas/shared/loggingmanager/enum"
	"lite-nas/shared/messaging"
	"lite-nas/shared/ruleevaluator"
)

const maxEventCounter = uint64(99_999_999)

// Processor orchestrates rule evaluation and alert lifecycle messaging.
type Processor struct {
	rules   []servicerules.Rule
	manager *eventmanager.Manager
	client  messaging.Client
	logger  sharedlogger.Logger
	clock   func() time.Time
}

// ActiveEvent stores in-memory state for one active alert rule key.
type ActiveEvent struct {
	EventID string
	Rule    servicerules.Rule
}

// New creates a Processor with rule set, event state manager, and messaging
// dependencies.
func New(
	rules []servicerules.Rule,
	manager *eventmanager.Manager,
	client messaging.Client,
	log sharedlogger.Logger,
) *Processor {
	return &Processor{
		rules:   rules,
		manager: manager,
		client:  client,
		logger:  log,
		clock:   time.Now,
	}
}

// HandleEnvelope processes one inbound messaging envelope.
func (p *Processor) HandleEnvelope(ctx context.Context, envelope messaging.Envelope) error {
	var payload map[string]any
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return err
	}

	for _, rule := range p.rules {
		if rule.Event != envelope.Subject {
			continue
		}

		p.processRule(ctx, rule, payload)
	}

	return nil
}

// processRule executes lifecycle transitions for one rule and one payload.
func (p *Processor) processRule(ctx context.Context, rule servicerules.Rule, payload map[string]any) {
	extractedValue, isMatch, ok := p.evaluateRule(rule, payload)
	if !ok {
		return
	}

	activeEvent, isActive := p.findActiveEvent(rule)
	p.logRuleEvaluation(rule, extractedValue, isMatch, isActive)
	p.handleRuleTransition(ctx, rule, extractedValue, isMatch, activeEvent, isActive)
}

// evaluateRule extracts and evaluates one rule against payload.
func (p *Processor) evaluateRule(rule servicerules.Rule, payload map[string]any) (any, bool, bool) {
	extractedValue, exists := ruleevaluator.ExtractValueByPath(payload, rule.Field)
	if !exists {
		p.logger.Debug(
			"rule field path not found in payload",
			"event",
			rule.Event,
			"field",
			rule.Field,
			"condition",
			rule.Condition,
		)
		return nil, false, false
	}

	isMatch := ruleevaluator.EvaluateCondition(extractedValue, rule.Condition, rule.Values)
	return extractedValue, isMatch, true
}

// logRuleEvaluation logs one completed rule evaluation.
func (p *Processor) logRuleEvaluation(rule servicerules.Rule, extractedValue any, isMatch bool, isActive bool) {
	p.logger.Debug(
		"rule evaluated",
		"event",
		rule.Event,
		"field",
		rule.Field,
		"condition",
		rule.Condition,
		"rule_value",
		rule.Values,
		"extracted_value",
		extractedValue,
		"match",
		isMatch,
		"active",
		isActive,
	)
}

// handleRuleTransition executes lifecycle transition based on match and active state.
func (p *Processor) handleRuleTransition(
	ctx context.Context,
	rule servicerules.Rule,
	extractedValue any,
	isMatch bool,
	activeEvent ActiveEvent,
	isActive bool,
) {
	if !isActive {
		if isMatch {
			p.handleNewToActive(ctx, rule)
		}
		return
	}

	if isMatch {
		p.handleActiveToActive(ctx, activeEvent, extractedValue)
		return
	}

	p.handleActiveToNormal(ctx, activeEvent)
}

// findActiveEvent looks up active event state for one rule key.
func (p *Processor) findActiveEvent(rule servicerules.Rule) (ActiveEvent, bool) {
	cached, exists := p.manager.FindEvent(rule.Event, rule.Field, rule.Condition)
	if !exists {
		return ActiveEvent{}, false
	}

	activeEvent, ok := cached.Payload.(ActiveEvent)
	if !ok {
		return ActiveEvent{}, false
	}

	return activeEvent, true
}

// handleNewToActive publishes a new alert and caches active state.
func (p *Processor) handleNewToActive(ctx context.Context, rule servicerules.Rule) {
	eventID := p.nextEventID(rule.EventPrefix)
	now := p.clock().UTC().Format(time.RFC3339)
	priority := rule.Priority
	createInput := loggingmanagercontract.AlertPayload{
		EventID:   eventID,
		Category:  rule.Category,
		Severity:  rule.Severity,
		Priority:  &priority,
		CreatedAt: now,
		Source:    rule.Source,
	}

	if err := p.client.Publish(ctx, systemloggingmanagercontract.AlertSubject, createInput); err != nil {
		p.logger.Warn("failed to publish alert create", "subject", systemloggingmanagercontract.AlertSubject, "error", err)
		return
	}

	p.logger.Info(
		"published alert create",
		"subject",
		systemloggingmanagercontract.AlertSubject,
		"event_id",
		eventID,
		"event",
		rule.Event,
		"field",
		rule.Field,
		"condition",
		rule.Condition,
	)

	if err := p.manager.CreateEvent(
		rule.Event,
		rule.Field,
		rule.Condition,
		ActiveEvent{EventID: eventID, Rule: rule},
	); err != nil {
		p.logger.Warn("failed to cache active event", "event_id", eventID, "error", err)
	}
}

// handleActiveToActive publishes one occurrence for an active alert.
func (p *Processor) handleActiveToActive(ctx context.Context, activeEvent ActiveEvent, extractedValue any) {
	occurrence := loggingmanagercontract.AlertOccurrencePayload{
		EventID:   activeEvent.EventID,
		Timestamp: p.clock().UTC().Format(time.RFC3339),
	}

	assignOccurrenceValue(&occurrence, extractedValue)

	if err := p.client.Publish(ctx, systemloggingmanagercontract.AlertOccurrenceSubject, occurrence); err != nil {
		p.logger.Warn("failed to publish alert occurrence", "subject", systemloggingmanagercontract.AlertOccurrenceSubject, "event_id", activeEvent.EventID, "error", err)
		return
	}

	p.logger.Info(
		"published alert occurrence",
		"subject",
		systemloggingmanagercontract.AlertOccurrenceSubject,
		"event_id",
		activeEvent.EventID,
		"event",
		activeEvent.Rule.Event,
		"field",
		activeEvent.Rule.Field,
		"condition",
		activeEvent.Rule.Condition,
		"value_type",
		occurrence.ValueType,
	)
}

// handleActiveToNormal updates alert state to normal and clears cached active
// event when the state update succeeds.
func (p *Processor) handleActiveToNormal(ctx context.Context, activeEvent ActiveEvent) {
	var message *string
	if activeEvent.Rule.NormalMessage != "" {
		normalizedMessage := activeEvent.Rule.NormalMessage
		message = &normalizedMessage
	}

	request := loggingmanagercontract.UpdateAlertStateInput{
		EventID: activeEvent.EventID,
		Status:  sharedloggingenum.StatusNormal,
		Message: message,
	}

	var response loggingmanagercontract.OKResponse
	if err := p.client.Request(ctx, systemloggingmanagercontract.UpdateAlertStateRPCSubject, request, &response); err != nil {
		p.logger.Warn("failed to normalize alert state", "subject", systemloggingmanagercontract.UpdateAlertStateRPCSubject, "event_id", activeEvent.EventID, "error", err)
		return
	}

	if !response.OK {
		p.logger.Warn("logging manager rejected normalize request", "event_id", activeEvent.EventID)
		return
	}

	p.manager.DeleteEvent(activeEvent.Rule.Event, activeEvent.Rule.Field, activeEvent.Rule.Condition)
}

// nextEventID generates the next event ID from prefix and in-memory counter.
func (p *Processor) nextEventID(prefix string) string {
	next := p.manager.NextCounter()
	if next > maxEventCounter {
		p.manager.SetCounter(1)
		next = 1
	}

	return fmt.Sprintf("%s_%08d", prefix, next)
}

// assignOccurrenceValue maps a dynamic value into typed occurrence fields.
func assignOccurrenceValue(occurrence *loggingmanagercontract.AlertOccurrencePayload, value any) {
	if numericValue, valueType, ok := toOccurrenceNumber(value); ok {
		occurrence.ValueType = valueType
		occurrence.ValueNum = &numericValue
		return
	}

	if booleanValue, ok := value.(bool); ok {
		occurrence.ValueType = sharedloggingenum.ValueTypeBool
		occurrence.ValueBool = &booleanValue
		return
	}

	textValue := fmt.Sprintf("%v", value)
	occurrence.ValueType = sharedloggingenum.ValueTypeText
	occurrence.ValueText = &textValue
}

// toOccurrenceNumber converts numeric values into occurrence numeric payload
// shape and associated value type.
func toOccurrenceNumber(value any) (float64, sharedloggingenum.ValueType, bool) {
	reflectedValue := reflect.ValueOf(value)
	if !reflectedValue.IsValid() {
		return 0, "", false
	}

	switch reflectedValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(reflectedValue.Int()), sharedloggingenum.ValueTypeInt, true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(reflectedValue.Uint()), sharedloggingenum.ValueTypeInt, true
	case reflect.Float32, reflect.Float64:
		return reflectedValue.Float(), sharedloggingenum.ValueTypeFloat, true
	default:
		return 0, "", false
	}
}
