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
	"lite-nas/shared/fsm"
	sharedlogger "lite-nas/shared/logger"
	sharedloggingenum "lite-nas/shared/loggingmanager/enum"
	"lite-nas/shared/messaging"
	"lite-nas/shared/ruleevaluator"
)

// Processor orchestrates rule evaluation and alert lifecycle messaging.
type Processor struct {
	rules   []servicerules.Rule
	manager *eventmanager.Manager
	client  messaging.Client
	logger  sharedlogger.Logger
	clock   func() time.Time
}

// lifecycleState is derived from the event manager after each successful
// external operation, so failed sends cannot advance the local lifecycle.
type (
	lifecycleState  string
	lifecycleInput  string
	lifecycleEffect string
)

const (
	lifecycleNormal  lifecycleState  = "normal"
	lifecycleActive  lifecycleState  = "active"
	lifecycleMatched lifecycleInput  = "matched"
	lifecycleCleared lifecycleInput  = "cleared"
	effectCreate     lifecycleEffect = "create"
	effectOccurrence lifecycleEffect = "occurrence"
	effectNormalize  lifecycleEffect = "normalize"
)

// lifecycleMachine makes alert create, occurrence, and normalization paths
// inspectable as one table while the event manager retains correlation data.
var lifecycleMachine = newLifecycleMachine()

func newLifecycleMachine() *fsm.Machine[lifecycleState, lifecycleInput, struct{}, lifecycleEffect] {
	action := func(effect lifecycleEffect) fsm.Action[struct{}, lifecycleInput, lifecycleEffect] {
		return func(_ context.Context, data struct{}, _ lifecycleInput) (struct{}, []lifecycleEffect, error) {
			return data, []lifecycleEffect{effect}, nil
		}
	}
	machine, err := fsm.New(fsm.Definition[lifecycleState, lifecycleInput, struct{}, lifecycleEffect]{
		Initial: lifecycleNormal,
		States:  []lifecycleState{lifecycleNormal, lifecycleActive},
		Inputs:  []lifecycleInput{lifecycleMatched, lifecycleCleared},
		Rows: []fsm.Row[lifecycleState, lifecycleInput]{
			{From: lifecycleNormal, On: lifecycleMatched, To: lifecycleActive, Action: string(effectCreate)},
			{From: lifecycleNormal, On: lifecycleCleared, To: lifecycleNormal},
			{From: lifecycleActive, On: lifecycleMatched, To: lifecycleActive, Action: string(effectOccurrence)},
			{From: lifecycleActive, On: lifecycleCleared, To: lifecycleNormal, Action: string(effectNormalize)},
		},
		Actions: map[string]fsm.Action[struct{}, lifecycleInput, lifecycleEffect]{
			string(effectCreate):     action(effectCreate),
			string(effectOccurrence): action(effectOccurrence),
			string(effectNormalize):  action(effectNormalize),
		},
	})
	if err != nil {
		panic(err)
	} // Static table failure is a programmer error.
	return machine
}

// ActiveEvent stores in-memory state for one active alert rule key.
type ActiveEvent struct {
	EventID    string
	Rule       servicerules.Rule
	FieldPath  string
	Qualifiers []string
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
	extractedValues, ok := p.evaluateRule(rule, payload)
	if !ok {
		return
	}

	for _, extractedValue := range extractedValues {
		activeEvent, isActive := p.findActiveEvent(rule, extractedValue.Qualifiers)
		p.logRuleEvaluation(rule, extractedValue, isActive)
		p.handleRuleTransition(ctx, rule, extractedValue, activeEvent, isActive)
	}
}

// evaluateRule extracts and evaluates one rule against payload.
func (p *Processor) evaluateRule(rule servicerules.Rule, payload map[string]any) ([]evaluatedValue, bool) {
	extractedValues, exists := ruleevaluator.ExtractValuesByPath(payload, rule.Field)
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
		return nil, false
	}

	evaluatedValues := make([]evaluatedValue, 0, len(extractedValues))
	for _, extractedValue := range extractedValues {
		evaluatedValues = append(evaluatedValues, evaluatedValue{
			FieldPath:  extractedValue.FieldPath,
			Qualifiers: ruleevaluator.FormatIndexQualifiers(extractedValue.Indexes),
			Value:      extractedValue.Value,
			IsMatch:    ruleevaluator.EvaluateCondition(extractedValue.Value, rule.Condition, rule.Values),
		})
	}

	return evaluatedValues, true
}

// logRuleEvaluation logs one completed rule evaluation.
func (p *Processor) logRuleEvaluation(rule servicerules.Rule, extractedValue evaluatedValue, isActive bool) {
	p.logger.Debug(
		"rule evaluated",
		"event",
		rule.Event,
		"field",
		rule.Field,
		"indexed_field",
		extractedValue.FieldPath,
		"condition",
		rule.Condition,
		"rule_value",
		rule.Values,
		"extracted_value",
		extractedValue.Value,
		"match",
		extractedValue.IsMatch,
		"active",
		isActive,
		"qualifiers",
		extractedValue.Qualifiers,
	)
}

// handleRuleTransition executes lifecycle transition based on match and active state.
func (p *Processor) handleRuleTransition(
	ctx context.Context,
	rule servicerules.Rule,
	extractedValue evaluatedValue,
	activeEvent ActiveEvent,
	isActive bool,
) {
	state := lifecycleNormal
	if isActive {
		state = lifecycleActive
	}
	input := lifecycleCleared
	if extractedValue.IsMatch {
		input = lifecycleMatched
	}
	transition, err := lifecycleMachine.Step(ctx, fsm.Instance[lifecycleState, struct{}]{State: state}, input)
	if err != nil {
		p.logger.Error("alert lifecycle transition failed", "state", state, "input", input, "error", err)
		return
	}
	for _, effect := range transition.Effects {
		switch effect {
		case effectCreate:
			p.handleNewToActive(ctx, rule, extractedValue)
		case effectOccurrence:
			p.handleActiveToActive(ctx, activeEvent, extractedValue.Value)
		case effectNormalize:
			p.handleActiveToNormal(ctx, activeEvent)
		}
	}
}

// findActiveEvent looks up active event state for one rule key.
func (p *Processor) findActiveEvent(rule servicerules.Rule, qualifiers []string) (ActiveEvent, bool) {
	cached, exists := p.manager.FindEvent(rule.Event, rule.Field, rule.Condition, qualifiers...)
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
func (p *Processor) handleNewToActive(ctx context.Context, rule servicerules.Rule, extractedValue evaluatedValue) {
	eventID := p.nextEventID(rule.EventPrefix)
	createInput := p.buildAlertCreateInput(rule, eventID, extractedValue)

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
		"indexed_field",
		extractedValue.FieldPath,
		"condition",
		rule.Condition,
	)

	if err := p.manager.CreateEvent(
		rule.Event,
		rule.Field,
		rule.Condition,
		buildActiveEvent(rule, eventID, extractedValue),
		extractedValue.Qualifiers...,
	); err != nil {
		p.logger.Warn("failed to cache active event", "event_id", eventID, "error", err)
	}
}

// buildAlertCreateInput constructs the logging-manager alert create payload for
// one newly matched rule.
func (p *Processor) buildAlertCreateInput(
	rule servicerules.Rule,
	eventID string,
	extractedValue evaluatedValue,
) loggingmanagercontract.AlertPayload {
	now := p.clock().UTC().Format(time.RFC3339)
	priority := rule.Priority

	return loggingmanagercontract.AlertPayload{
		EventID:      eventID,
		Category:     rule.Category,
		Severity:     rule.Severity,
		Priority:     &priority,
		CreatedAt:    now,
		Source:       rule.Source,
		Message:      rule.Message,
		TriggerValue: fmt.Sprintf("%v", extractedValue.Value),
	}
}

// buildActiveEvent constructs the in-memory active-event cache payload for one
// matched rule key.
func buildActiveEvent(
	rule servicerules.Rule,
	eventID string,
	extractedValue evaluatedValue,
) ActiveEvent {
	return ActiveEvent{
		EventID:    eventID,
		Rule:       rule,
		FieldPath:  extractedValue.FieldPath,
		Qualifiers: append([]string(nil), extractedValue.Qualifiers...),
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
		"indexed_field",
		activeEvent.FieldPath,
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

	p.manager.DeleteEvent(
		activeEvent.Rule.Event,
		activeEvent.Rule.Field,
		activeEvent.Rule.Condition,
		activeEvent.Qualifiers...,
	)
}

// evaluatedValue stores one resolved rule field value together with its
// indexed identity and condition result.
type evaluatedValue struct {
	FieldPath  string
	Qualifiers []string
	Value      any
	IsMatch    bool
}

// nextEventID generates the next event ID from prefix and in-memory counter.
func (p *Processor) nextEventID(prefix string) string {
	return p.manager.NextEventID(prefix)
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
