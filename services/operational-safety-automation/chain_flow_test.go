package main

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	alertcontract "lite-nas/shared/contracts/loggingmanager"
	safetycontract "lite-nas/shared/contracts/operationalsafety"
	"lite-nas/shared/logger"
)

// chainAlert identifies one simulated NATS publication from the real effect worker.
type chainAlert struct {
	ruleID string
	id     string
}

// chainFlow drives configured Raspberry Pi rules with a virtual clock and fake host effects.
type chainFlow struct {
	runtime    *safetyRuntime
	now        time.Time
	published  chan chainAlert
	normalized chan string
	mu         sync.Mutex
	caps       []uint64
	sequence   map[string]int
}

// mustChainFlow compiles the documented rules and runs the service's outbound worker.
func mustChainFlow(t *testing.T) *chainFlow {
	t.Helper()
	flow := &chainFlow{now: time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC), published: make(chan chainAlert, 16), normalized: make(chan string, 16), sequence: make(map[string]int)}
	rules, order, err := loadRules("../../docs/examples/operational-safety-rpi-cpu-rules.json")
	if err != nil {
		t.Fatal(err)
	}
	for id, rule := range rules {
		if rule.detector != nil {
			rules[id], err = compileRuleWithClock(rule.spec, func() time.Time { return flow.now })
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	effects := make(chan effectIntent, 16)
	flow.runtime = &safetyRuntime{rules: rules, order: order, tasks: make(map[string]*safetyTask), stages: make(map[string]*stageTask), events: make(map[string]*observedEvent), intents: effects, now: func() time.Time { return flow.now }, log: logger.NopLogger{}}
	flow.runtime.apply = func(_ context.Context, capKHz, _ uint64) error {
		flow.mu.Lock()
		flow.caps = append(flow.caps, capKHz)
		flow.mu.Unlock()
		return nil
	}
	flow.runtime.publishOwned = func(_ context.Context, ruleID, _ string) (string, error) {
		prefix := flow.runtime.rules[ruleID].spec.Output.IDPrefix
		flow.sequence[prefix]++
		id := fmt.Sprintf("%s_%d", prefix, flow.sequence[prefix])
		flow.published <- chainAlert{ruleID: ruleID, id: id}
		return id, nil
	}
	flow.runtime.normalizeOwned = func(_ context.Context, id string) error {
		flow.normalized <- id
		return nil
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for intent := range effects {
			flow.runtime.applyIntent(context.Background(), intent)
		}
	}()
	t.Cleanup(func() { close(effects); <-done })
	return flow
}

// mustObserve sends one validated-in-production observation through the dispatcher.
func (flow *chainFlow) mustObserve(t *testing.T, item safetyObservation) {
	t.Helper()
	if err := flow.runtime.processObservation(context.Background(), item); err != nil {
		t.Fatal(err)
	}
}

// mustAlert receives the next rule-owned event from the fake NATS publication boundary.
func (flow *chainFlow) mustAlert(t *testing.T, ruleID string) string {
	t.Helper()
	select {
	case alert := <-flow.published:
		if alert.ruleID != ruleID {
			t.Fatalf("published rule = %q, want %q", alert.ruleID, ruleID)
		}
		return alert.id
	case <-time.After(time.Second):
		t.Fatalf("rule %q did not publish", ruleID)
		return ""
	}
}

// mustNormalized confirms the outbound worker closed exactly the expected stage event.
func (flow *chainFlow) mustNormalized(t *testing.T, id string) {
	t.Helper()
	select {
	case got := <-flow.normalized:
		if got != id {
			t.Fatalf("normalized event = %q, want %q", got, id)
		}
	case <-time.After(time.Second):
		t.Fatalf("event %q was not normalized", id)
	}
}

// mustNormalRetryState waits for the outbound worker's pending-record bookkeeping.
func (flow *chainFlow) mustNormalRetryState(t *testing.T, id string, wantPending bool) {
	t.Helper()
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	deadline := time.After(time.Second)
	for {
		flow.runtime.mu.Lock()
		pending, exists := flow.runtime.pendingNormal[id]
		flow.runtime.mu.Unlock()
		if wantPending && exists && !pending || !wantPending && !exists {
			return
		}
		select {
		case <-ticker.C:
		case <-deadline:
			t.Fatalf("normal-event pending state did not become %t", wantPending)
		}
	}
}

// mustEcho models the service receiving its own published alert through NATS.
func (flow *chainFlow) mustEcho(t *testing.T, id string) {
	t.Helper()
	flow.mustObserve(t, safetyObservation{kind: "create", eventID: id, alert: alertcontract.AlertPayload{EventID: id, Source: "operational-safety-automation", Category: "operational_safety.cpu"}})
}

// mustCriticalChain makes a direct critical sensor alert descend through all three caps.
func (flow *chainFlow) mustCriticalChain(t *testing.T) (string, string, string) {
	t.Helper()
	flow.mustObserve(t, safetyObservation{kind: "create", eventID: "snstcrit_1", alert: alertcontract.AlertPayload{EventID: "snstcrit_1", Source: "sensor-monitor", Category: "sensor.temperature.cpu"}})
	warning := flow.mustAlert(t, "cpu_warning")
	flow.mustEcho(t, warning)
	errorID := flow.mustAlert(t, "cpu_error")
	flow.mustEcho(t, errorID)
	critical := flow.mustAlert(t, "cpu_critical")
	flow.mustEcho(t, critical)
	return warning, errorID, critical
}

// Requirements: operational-safety-automation/FR-CHAIN-006.
func TestCriticalSensorJumpDescendsThroughPublishedStages(t *testing.T) {
	flow := mustChainFlow(t)
	flow.mustCriticalChain(t)
	flow.mu.Lock()
	got := append([]uint64(nil), flow.caps...)
	flow.mu.Unlock()
	if !reflect.DeepEqual(got, []uint64{1800000, 1600000, 1500000}) {
		t.Fatalf("critical jump caps = %v", got)
	}
}

// Requirements: operational-safety-automation/FR-CHAIN-007.
func TestChainUnwindsOneEventAfterEachVerifiedCooldown(t *testing.T) {
	flow := mustChainFlow(t)
	warning, errorID, critical := flow.mustCriticalChain(t)
	flow.mustObserve(t, safetyObservation{kind: "state", eventID: "snstcrit_1", state: "normal"})
	flow.now = flow.now.Add(4*time.Minute + 59*time.Second)
	flow.runtime.advanceCooldowns(context.Background())
	flow.now = flow.now.Add(time.Second)
	flow.runtime.advanceCooldowns(context.Background())
	flow.mustNormalized(t, critical)
	flow.mustObserve(t, safetyObservation{kind: "state", eventID: critical, state: "normal"})
	flow.now = flow.now.Add(5 * time.Minute)
	flow.runtime.advanceCooldowns(context.Background())
	flow.mustNormalized(t, errorID)
	flow.mustObserve(t, safetyObservation{kind: "state", eventID: errorID, state: "normal"})
	flow.now = flow.now.Add(5 * time.Minute)
	flow.runtime.advanceCooldowns(context.Background())
	flow.mustNormalized(t, warning)
	flow.mustObserve(t, safetyObservation{kind: "state", eventID: warning, state: "normal"})
	normal := flow.mustAlert(t, "cpu_normal")
	flow.mustNormalized(t, normal)
	flow.mu.Lock()
	got := append([]uint64(nil), flow.caps...)
	flow.mu.Unlock()
	if !reflect.DeepEqual(got, []uint64{1800000, 1600000, 1500000, 1600000, 1800000, 1900000}) {
		t.Fatalf("chain cap progression = %v", got)
	}
}

// Requirements: operational-safety-automation/FR-CHAIN-008.
func TestBadValueDuringCooldownRestartsStageWait(t *testing.T) {
	flow := mustChainFlow(t)
	_, _, critical := flow.mustCriticalChain(t)
	flow.mustObserve(t, safetyObservation{kind: "state", eventID: "snstcrit_1", state: "normal"})
	flow.now = flow.now.Add(2 * time.Minute)
	flow.mustObserve(t, safetyObservation{kind: "create", eventID: "snstcrit_2", alert: alertcontract.AlertPayload{EventID: "snstcrit_2", Source: "sensor-monitor", Category: "sensor.temperature.cpu"}})
	flow.now = flow.now.Add(4 * time.Minute)
	flow.runtime.advanceCooldowns(context.Background())
	flow.mustObserve(t, safetyObservation{kind: "state", eventID: "snstcrit_2", state: "normal"})
	flow.now = flow.now.Add(4*time.Minute + 59*time.Second)
	flow.runtime.advanceCooldowns(context.Background())
	if got := len(flow.normalized); got != 0 {
		t.Fatalf("events normalized before renewed cooldown = %d", got)
	}
	flow.now = flow.now.Add(time.Second)
	flow.runtime.advanceCooldowns(context.Background())
	flow.mustNormalized(t, critical)
}

// Requirements: operational-safety-automation/FR-CHAIN-009.
func TestFailedRestoreKeepsStageEventActive(t *testing.T) {
	flow := mustChainFlow(t)
	_, _, critical := flow.mustCriticalChain(t)
	flow.runtime.apply = func(_ context.Context, capKHz, _ uint64) error {
		if capKHz == 1600000 {
			return errors.New("simulated policy write failure")
		}
		return nil
	}
	flow.mustObserve(t, safetyObservation{kind: "state", eventID: "snstcrit_1", state: "normal"})
	flow.now = flow.now.Add(5 * time.Minute)
	flow.runtime.advanceCooldowns(context.Background())
	got := struct {
		phase      string
		ownEventID string
		normalized int
	}{phase: string(flow.runtime.stages["cpu_critical"].instance.State), ownEventID: flow.runtime.stages["cpu_critical"].eventID, normalized: len(flow.normalized)}
	want := struct {
		phase      string
		ownEventID string
		normalized int
	}{phase: "cooling", ownEventID: critical}
	if got != want {
		t.Fatalf("failed restore state = %#v, want %#v", got, want)
	}
}

// Requirements: operational-safety-automation/FR-CHAIN-010.
func TestNormalRecordRetriesFailedNormalization(t *testing.T) {
	flow := mustChainFlow(t)
	attempts := 0
	flow.runtime.normalizeOwned = func(_ context.Context, id string) error {
		attempts++
		if attempts == 1 {
			return errors.New("temporary Logging Manager failure")
		}
		flow.normalized <- id
		return nil
	}
	flow.mustEcho(t, "opswarn_99")
	flow.mustObserve(t, safetyObservation{kind: "state", eventID: "opswarn_99", state: "normal"})
	normal := flow.mustAlert(t, "cpu_normal")
	flow.mustNormalRetryState(t, normal, true)
	flow.runtime.retryNormalRecords(context.Background())
	flow.mustNormalized(t, normal)
	flow.mustNormalRetryState(t, normal, false)
}

// Requirements: operational-safety-automation/FR-CHAIN-011.
func TestLinkedStageTaskForceStopRespectsActiveChild(t *testing.T) {
	flow := mustChainFlow(t)
	flow.mustCriticalChain(t)
	operator, token := mustSafetyOperator(t, []string{"lite-nas-security"})
	flow.runtime.verifier = operator.verifier
	flow.runtime.client = operator.client
	id := flow.runtime.stages["cpu_warning"].id
	response, err := flow.runtime.forceStop(context.Background(), mustSafetyEnvelope(t, safetycontract.OperationalSafetyTaskForceStopRequest{AccessToken: token, TaskID: id}))
	if err != nil {
		t.Fatal(err)
	}
	if got := response.(safetycontract.OperationalSafetyOperationResponse).State; got != safetycontract.OutcomeBlocked {
		t.Fatalf("linked stage force stop state = %q", got)
	}
}

// Requirements: operational-safety-automation/FR-CHAIN-011.
func TestLinkedStageTaskForceSetUsesDeclaredCap(t *testing.T) {
	flow := mustChainFlow(t)
	flow.mustObserve(t, safetyObservation{kind: "create", eventID: "snsvwrn_1", alert: alertcontract.AlertPayload{EventID: "snsvwrn_1", Source: "sensor-monitor", Category: "sensor.power.input"}})
	warning := flow.mustAlert(t, "cpu_warning")
	flow.mustEcho(t, warning)
	flow.mustObserve(t, safetyObservation{kind: "state", eventID: "snsvwrn_1", state: "normal"})
	flow.now = flow.now.Add(5 * time.Minute)
	flow.runtime.advanceCooldowns(context.Background())
	flow.mustNormalized(t, warning)
	response, err := flow.runtime.forceStageSet(context.Background(), "cpu_warning", safetycontract.LevelWarning)
	if err != nil {
		t.Fatal(err)
	}
	if response.State != safetycontract.OutcomeCompleted {
		t.Fatalf("linked stage force set state = %q", response.State)
	}
}

// Requirements: operational-safety-automation/FR-CHAIN-013.
func TestWarningWaitsForBothPowerAndTemperatureToClear(t *testing.T) {
	flow := mustChainFlow(t)
	flow.mustObserve(t, safetyObservation{kind: "create", eventID: "snsvwrn_1", alert: alertcontract.AlertPayload{EventID: "snsvwrn_1", Source: "sensor-monitor", Category: "sensor.power.input"}})
	warning := flow.mustAlert(t, "cpu_warning")
	flow.mustEcho(t, warning)
	flow.mustObserve(t, safetyObservation{kind: "create", eventID: "snstwrn_1", alert: alertcontract.AlertPayload{EventID: "snstwrn_1", Source: "sensor-monitor", Category: "sensor.temperature.cpu"}})
	flow.mustObserve(t, safetyObservation{kind: "state", eventID: "snsvwrn_1", state: "normal"})
	flow.now = flow.now.Add(6 * time.Minute)
	flow.runtime.advanceCooldowns(context.Background())
	got := struct {
		phase      string
		normalized int
	}{phase: string(flow.runtime.stages["cpu_warning"].instance.State), normalized: len(flow.normalized)}
	want := struct {
		phase      string
		normalized int
	}{phase: "active"}
	if got != want {
		t.Fatalf("warning with temperature still active = %#v, want %#v", got, want)
	}
}
