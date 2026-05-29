package processor

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	servicerules "lite-nas/services/resources-monitor/rules"
	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	systemloggingmanagercontract "lite-nas/shared/contracts/systemloggingmanager"
	"lite-nas/shared/eventmanager"
	sharedlogger "lite-nas/shared/logger"
	sharedloggingenum "lite-nas/shared/loggingmanager/enum"
	"lite-nas/shared/messaging"
)

func TestHandleEnvelopeRunsAlertLifecycle(t *testing.T) {
	t.Parallel()

	client := &recordingClient{requestResponse: loggingmanagercontract.OKResponse{OK: true}}
	rule := buildMemoryThresholdRule()

	manager := eventmanager.NewManager(0)
	processor := New([]servicerules.Rule{rule}, manager, client, sharedlogger.NewNop())
	ctx := context.Background()

	highPayload := buildEnvelopePayload(t, 93.0)
	lowPayload := buildEnvelopePayload(t, 80.0)

	assertHandleEnvelopeNoError(t, processor, ctx, highPayload, "first")
	assertPublishedSubjects(t, client.publishSubjects, []string{systemloggingmanagercontract.AlertSubject})
	assertActiveEventExists(t, manager, rule)

	assertHandleEnvelopeNoError(t, processor, ctx, highPayload, "second")
	assertPublishedSubjects(t, client.publishSubjects, []string{
		systemloggingmanagercontract.AlertSubject,
		systemloggingmanagercontract.AlertOccurrenceSubject,
	})

	assertHandleEnvelopeNoError(t, processor, ctx, lowPayload, "third")
	assertNormalizeRequestSubject(t, client.requestSubject)
	assertActiveEventMissing(t, manager, rule)
}

func TestHandleEnvelopeRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	processor := New(nil, eventmanager.NewManager(0), &recordingClient{}, sharedlogger.NewNop())
	err := processor.HandleEnvelope(context.Background(), messaging.Envelope{
		Subject: "system.metrics.events.stats",
		Payload: []byte("{"),
	})
	if err == nil {
		t.Fatal("expected decode error")
	}
}

func TestHandleNewToActiveSkipsCacheWhenPublishFails(t *testing.T) {
	t.Parallel()

	client := &recordingClient{publishErr: errors.New("publish failed")}
	rule := buildMemoryThresholdRule()
	manager := eventmanager.NewManager(0)
	processor := New([]servicerules.Rule{rule}, manager, client, sharedlogger.NewNop())

	err := processor.HandleEnvelope(context.Background(), messaging.Envelope{
		Subject: "system.metrics.events.stats",
		Payload: buildEnvelopePayload(t, 95.0),
	})
	if err != nil {
		t.Fatalf("HandleEnvelope() error = %v", err)
	}

	if _, exists := manager.FindEvent(rule.Event, rule.Field, rule.Condition); exists {
		t.Fatal("event should not be cached when create publish fails")
	}
}

func TestHandleEnvelopeTracksArrayRulePerPoolIndex(t *testing.T) {
	t.Parallel()

	client := &recordingClient{requestResponse: loggingmanagercontract.OKResponse{OK: true}}
	rule := buildPoolHealthRule()
	manager := eventmanager.NewManager(0)
	processor := New([]servicerules.Rule{rule}, manager, client, sharedlogger.NewNop())
	ctx := context.Background()

	degradedPayload := buildPoolEnvelopePayload(t, "ONLINE", "DEGRADED")
	healthyPayload := buildPoolEnvelopePayload(t, "ONLINE", "ONLINE")

	assertHandleEnvelopeNoError(t, processor, ctx, degradedPayload, "activate indexed pool")
	assertPublishedSubjects(t, client.publishSubjects, []string{systemloggingmanagercontract.AlertSubject})
	assertActiveEventExistsForQualifiers(t, manager, rule, "1")
	assertActiveEventMissingForQualifiers(t, manager, rule, "0")

	assertHandleEnvelopeNoError(t, processor, ctx, degradedPayload, "repeat indexed pool")
	assertPublishedSubjects(t, client.publishSubjects, []string{
		systemloggingmanagercontract.AlertSubject,
		systemloggingmanagercontract.AlertOccurrenceSubject,
	})

	assertHandleEnvelopeNoError(t, processor, ctx, healthyPayload, "normalize indexed pool")
	assertNormalizeRequestSubject(t, client.requestSubject)
	assertActiveEventMissingForQualifiers(t, manager, rule, "1")
}

func TestHandleActiveToNormalKeepsCacheWhenNormalizationFails(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		client *recordingClient
	}{
		{
			name: "request failure",
			client: &recordingClient{
				requestResponse: loggingmanagercontract.OKResponse{OK: true},
				requestErr:      errors.New("request failed"),
			},
		},
		{
			name:   "rejected state update",
			client: &recordingClient{requestResponse: loggingmanagercontract.OKResponse{OK: false}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			rule := buildMemoryThresholdRule()
			manager := eventmanager.NewManager(0)
			processor := New([]servicerules.Rule{rule}, manager, testCase.client, sharedlogger.NewNop())

			assertHandleEnvelopeNoError(t, processor, context.Background(), buildEnvelopePayload(t, 95.0), "activate")
			assertHandleEnvelopeNoError(t, processor, context.Background(), buildEnvelopePayload(t, 50.0), "normalize")

			assertActiveEventExists(t, manager, rule)
		})
	}
}

func TestAssignOccurrenceValueSetsBoolTypeForBoolInput(t *testing.T) {
	t.Parallel()

	occurrence := loggingmanagercontract.AlertOccurrencePayload{}
	assignOccurrenceValue(&occurrence, true)

	if occurrence.ValueType != sharedloggingenum.ValueTypeBool {
		t.Fatalf("occurrence.ValueType = %q, want %q", occurrence.ValueType, sharedloggingenum.ValueTypeBool)
	}
}

func TestAssignOccurrenceValueSetsBoolValueForBoolInput(t *testing.T) {
	t.Parallel()

	occurrence := loggingmanagercontract.AlertOccurrencePayload{}
	assignOccurrenceValue(&occurrence, true)

	if occurrence.ValueBool == nil || !*occurrence.ValueBool {
		t.Fatal("occurrence.ValueBool should be true")
	}
}

func TestAssignOccurrenceValueSetsTextTypeForTextInput(t *testing.T) {
	t.Parallel()

	occurrence := loggingmanagercontract.AlertOccurrencePayload{}
	assignOccurrenceValue(&occurrence, "text")

	if occurrence.ValueType != sharedloggingenum.ValueTypeText {
		t.Fatalf("occurrence.ValueType = %q, want %q", occurrence.ValueType, sharedloggingenum.ValueTypeText)
	}
}

func TestAssignOccurrenceValueSetsTextValueForTextInput(t *testing.T) {
	t.Parallel()

	occurrence := loggingmanagercontract.AlertOccurrencePayload{}
	assignOccurrenceValue(&occurrence, "text")

	if occurrence.ValueText == nil || *occurrence.ValueText != "text" {
		t.Fatal("occurrence.ValueText should be \"text\"")
	}
}

// recordingClient captures publish and request calls for processor tests.
type recordingClient struct {
	publishSubjects []string
	requestSubject  string
	requestResponse loggingmanagercontract.OKResponse
	publishErr      error
	requestErr      error
}

// Publish records subject usage.
func (client *recordingClient) Publish(_ context.Context, subject string, _ any) error {
	client.publishSubjects = append(client.publishSubjects, subject)
	return client.publishErr
}

// Request records request subject usage and returns configured response.
func (client *recordingClient) Request(_ context.Context, subject string, _ any, response any) error {
	client.requestSubject = subject
	typedResponse, ok := response.(*loggingmanagercontract.OKResponse)
	if ok {
		*typedResponse = client.requestResponse
	}
	return client.requestErr
}

// Drain is a no-op for recording client.
func (client *recordingClient) Drain() error {
	return nil
}

// Close is a no-op for recording client.
func (client *recordingClient) Close() {}

// mustMarshalEnvelopePayload marshals map payload into JSON bytes.
func mustMarshalEnvelopePayload(t *testing.T, payload map[string]any) []byte {
	t.Helper()
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	return data
}

func buildMemoryThresholdRule() servicerules.Rule {
	return servicerules.Rule{
		Event:       "system.metrics.events.stats",
		EventPrefix: "sysram",
		Field:       "snapshot.mem.usedPct",
		Condition:   ">=",
		Values:      90.0,
		Message:     "RAM usage is above threshold",
		Category:    "system.metrics.mem.used",
		Severity:    sharedloggingenum.SeverityWarning,
		Priority:    2,
		Source:      "system-metrics",
	}
}

func buildPoolHealthRule() servicerules.Rule {
	return servicerules.Rule{
		Event:         "zfs.metrics.events.snapshot",
		EventPrefix:   "zfspool",
		Field:         "snapshot.Pools[].Health",
		Condition:     "==",
		Values:        "DEGRADED",
		Message:       "Pool health is degraded",
		NormalMessage: "Pool health returned to normal",
		Category:      "zfs.metrics.pool.health",
		Severity:      sharedloggingenum.SeverityWarning,
		Priority:      2,
		Source:        "zfs-metrics",
	}
}

func buildEnvelopePayload(t *testing.T, usedPct float64) []byte {
	t.Helper()

	return mustMarshalEnvelopePayload(t, map[string]any{
		"snapshot": map[string]any{
			"mem": map[string]any{"usedPct": usedPct},
		},
	})
}

func buildPoolEnvelopePayload(t *testing.T, healthValues ...string) []byte {
	t.Helper()

	pools := make([]map[string]any, 0, len(healthValues))
	for index, healthValue := range healthValues {
		pools = append(pools, map[string]any{
			"Name":   "pool",
			"Health": healthValue,
			"Index":  index,
		})
	}

	return mustMarshalEnvelopePayload(t, map[string]any{
		"snapshot": map[string]any{
			"Pools": pools,
		},
	})
}

func assertHandleEnvelopeNoError(
	t *testing.T,
	processor *Processor,
	ctx context.Context,
	payload []byte,
	stage string,
) {
	t.Helper()

	if err := processor.HandleEnvelope(ctx, messaging.Envelope{
		Subject: processor.rules[0].Event,
		Payload: payload,
	}); err != nil {
		t.Fatalf("HandleEnvelope() %s error = %v", stage, err)
	}
}

func assertPublishedSubjects(t *testing.T, got []string, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("len(publishSubjects) = %d, want %d; got=%v", len(got), len(want), got)
	}

	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("publishSubjects[%d] = %q, want %q", index, got[index], want[index])
		}
	}
}

func assertNormalizeRequestSubject(t *testing.T, got string) {
	t.Helper()

	if got != systemloggingmanagercontract.UpdateAlertStateRPCSubject {
		t.Fatalf("request subject = %q, want %q", got, systemloggingmanagercontract.UpdateAlertStateRPCSubject)
	}
}

func assertActiveEventExists(t *testing.T, manager *eventmanager.Manager, rule servicerules.Rule) {
	t.Helper()

	if _, exists := manager.FindEvent(rule.Event, rule.Field, rule.Condition); !exists {
		t.Fatal("active event missing")
	}
}

func assertActiveEventMissing(t *testing.T, manager *eventmanager.Manager, rule servicerules.Rule) {
	t.Helper()

	if _, exists := manager.FindEvent(rule.Event, rule.Field, rule.Condition); exists {
		t.Fatal("active event still exists")
	}
}

func assertActiveEventExistsForQualifiers(
	t *testing.T,
	manager *eventmanager.Manager,
	rule servicerules.Rule,
	qualifiers ...string,
) {
	t.Helper()

	if _, exists := manager.FindEvent(rule.Event, rule.Field, rule.Condition, qualifiers...); !exists {
		t.Fatalf("active event missing for qualifiers %v", qualifiers)
	}
}

func assertActiveEventMissingForQualifiers(
	t *testing.T,
	manager *eventmanager.Manager,
	rule servicerules.Rule,
	qualifiers ...string,
) {
	t.Helper()

	if _, exists := manager.FindEvent(rule.Event, rule.Field, rule.Condition, qualifiers...); exists {
		t.Fatalf("active event still exists for qualifiers %v", qualifiers)
	}
}
