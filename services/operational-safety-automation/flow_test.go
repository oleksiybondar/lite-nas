package main

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	detection "lite-nas/shared/automationdetection"
	alertcontract "lite-nas/shared/contracts/loggingmanager"
	contract "lite-nas/shared/contracts/operationalsafety"
	"lite-nas/shared/logger"
)

// safetyFlow is a deterministic observation, publication, and CPU effect fixture.
type safetyFlow struct {
	runtime   *safetyRuntime
	now       time.Time
	published chan effectIntent
	mu        sync.Mutex
	caps      []uint64
}

// mustSafetyFlow compiles real rule FSMs and runs the normal outbound effect worker.
func mustSafetyFlow(t *testing.T) *safetyFlow {
	t.Helper()
	flow := &safetyFlow{now: time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC), published: make(chan effectIntent, 16)}
	enabled := true
	emit := ruleSpec{ID: "power_chain", Name: "Power chain", Enabled: &enabled, CorrelateBy: "source", Steps: []selector{{IDPrefix: "power"}, {IDPrefix: "thermal"}}}
	emit.Action.Kind = "emit"
	emit.FSM.Type = detection.Sequence
	emit.FSM.Window = "1m"
	emit.FSM.MaxKeys = 16
	emit.FSM.MaxIdle = "10m"
	emit.Output.IDPrefix = "opsafe"
	emit.Output.Category = "system"
	emit.Output.Severity = "warning"
	emit.Output.Message = "Power and thermal condition"
	cpu := ruleSpec{ID: "cpu_limit", Name: "CPU limit", Enabled: &enabled, CorrelateBy: "source", Steps: []selector{{IDPrefix: "opsafe", Level: contract.LevelCritical}}}
	cpu.Action.Kind = "cpu_frequency"
	cpu.Action.Cooldown = "1m"
	cpu.Action.CapsKHz = map[contract.OperationalSafetyLevel]uint64{contract.LevelNormal: 1900000, contract.LevelWarning: 1500000, contract.LevelError: 1200000, contract.LevelCritical: 1000000}
	cpu.Output.IDPrefix = "cpusafe"
	cpu.Output.Category = "system"
	cpu.Output.Severity = "warning"
	cpu.Output.Message = "CPU cap changed"
	emitRule, err := compileRuleWithClock(emit, func() time.Time { return flow.now })
	if err != nil {
		t.Fatal(err)
	}
	cpuRule, err := compileRuleWithClock(cpu, func() time.Time { return flow.now })
	if err != nil {
		t.Fatal(err)
	}
	if err := validateRuleGraph(map[string]*compiledRule{emit.ID: emitRule, cpu.ID: cpuRule}); err != nil {
		t.Fatal(err)
	}
	effects := make(chan effectIntent, 16)
	flow.runtime = &safetyRuntime{rules: map[string]*compiledRule{emit.ID: emitRule, cpu.ID: cpuRule}, order: []string{emit.ID, cpu.ID}, tasks: map[string]*safetyTask{}, events: map[string]*observedEvent{}, intents: effects, now: func() time.Time { return flow.now }, log: logger.NopLogger{}}
	flow.runtime.apply = func(_ context.Context, capKHz, _ uint64) error {
		flow.mu.Lock()
		flow.caps = append(flow.caps, capKHz)
		flow.mu.Unlock()
		return nil
	}
	flow.runtime.publish = func(_ context.Context, intent effectIntent) error {
		flow.published <- intent
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

// mustObserve advances the real service dispatcher with one already validated alert observation.
func (flow *safetyFlow) mustObserve(t *testing.T, item safetyObservation) {
	t.Helper()
	if err := flow.runtime.processObservation(context.Background(), item); err != nil {
		t.Fatal(err)
	}
}

// mustFailObservation checks the expected error branch of a simulated effect failure.
func (flow *safetyFlow) mustFailObservation(t *testing.T, item safetyObservation) {
	t.Helper()
	if err := flow.runtime.processObservation(context.Background(), item); err == nil {
		t.Fatal("failed CPU effect was reported as successful")
	}
}

// mustPublication waits only for outbound worker completion; FSM time remains controlled by the fixture.
func (flow *safetyFlow) mustPublication(t *testing.T) effectIntent {
	t.Helper()
	select {
	case intent := <-flow.published:
		return intent
	case <-time.After(time.Second):
		t.Fatal("outbound publication did not complete")
		return effectIntent{}
	}
}

// publicationCountAfterFlush waits for a marker behind all earlier outbound intents.
func (flow *safetyFlow) publicationCountAfterFlush(t *testing.T) int {
	t.Helper()
	done := make(chan error, 1)
	if err := flow.runtime.enqueueEffect(context.Background(), effectIntent{kind: "emit", ruleID: "power_chain", key: "test_barrier", done: done}); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	count := 0
	for {
		intent := flow.mustPublication(t)
		if intent.key == "test_barrier" {
			return count
		}
		count++
	}
}

// mustStartCriticalTask feeds the source sequence, loops its emitted alert back, and applies a CPU cap.
func (flow *safetyFlow) mustStartCriticalTask(t *testing.T) {
	t.Helper()
	flow.mustObserve(t, safetyObservation{kind: "create", eventID: "power_1", alert: alertcontract.AlertPayload{EventID: "power_1", Source: "host_1", Message: "Low power"}})
	flow.mustObserve(t, safetyObservation{kind: "create", eventID: "thermal_1", alert: alertcontract.AlertPayload{EventID: "thermal_1", Source: "host_1", Message: "Hot CPU"}})
	if got := flow.mustPublication(t); got.kind != "emit" {
		t.Fatalf("first publication kind = %q", got.kind)
	}
	flow.mustObserve(t, safetyObservation{kind: "create", eventID: "opsafe_1", alert: alertcontract.AlertPayload{EventID: "opsafe_1", Source: "operational-safety-automation", Message: "Power and thermal condition"}})
	flow.mustPublication(t)
}

// Requirements: operational-safety-automation/operator-surface.
func TestOperationalSafetySequenceEmitsThenAppliesCPUCap(t *testing.T) {
	flow := mustSafetyFlow(t)
	flow.mustStartCriticalTask(t)
	flow.mu.Lock()
	got := append([]uint64(nil), flow.caps...)
	flow.mu.Unlock()
	if !reflect.DeepEqual(got, []uint64{1000000}) {
		t.Fatalf("CPU cap requests = %v", got)
	}
}

// Requirements: operational-safety-automation/operator-surface.
func TestOperationalSafetyRecoversOneVerifiedLevelPerCooldown(t *testing.T) {
	flow := mustSafetyFlow(t)
	flow.mustStartCriticalTask(t)
	flow.mustObserve(t, safetyObservation{kind: "state", eventID: "opsafe_1", state: "normal"})
	flow.now = flow.now.Add(59 * time.Second)
	flow.runtime.advanceCooldowns(context.Background())
	for range 3 {
		flow.now = flow.now.Add(time.Second)
		flow.runtime.advanceCooldowns(context.Background())
		flow.mustPublication(t)
		flow.now = flow.now.Add(59 * time.Second)
	}
	flow.mu.Lock()
	got := append([]uint64(nil), flow.caps...)
	flow.mu.Unlock()
	if !reflect.DeepEqual(got, []uint64{1000000, 1200000, 1500000, 1900000}) {
		t.Fatalf("verified cap progression = %v", got)
	}
}

// Requirements: operational-safety-automation/operator-surface.
func TestOperationalSafetyActiveAlertBlocksRecovery(t *testing.T) {
	flow := mustSafetyFlow(t)
	flow.mustStartCriticalTask(t)
	flow.mustObserve(t, safetyObservation{kind: "create", eventID: "opsafe_2", alert: alertcontract.AlertPayload{EventID: "opsafe_2", Source: "operational-safety-automation", Message: "Second condition"}})
	flow.mustObserve(t, safetyObservation{kind: "state", eventID: "opsafe_1", state: "normal"})
	flow.now = flow.now.Add(2 * time.Minute)
	flow.runtime.advanceCooldowns(context.Background())
	flow.mu.Lock()
	got := append([]uint64(nil), flow.caps...)
	flow.mu.Unlock()
	if !reflect.DeepEqual(got, []uint64{1000000}) {
		t.Fatalf("CPU cap requests with active alert = %v", got)
	}
}

// Requirements: operational-safety-automation/operator-surface.
func TestOperationalSafetyOutOfOrderSequenceDoesNotEmit(t *testing.T) {
	flow := mustSafetyFlow(t)
	flow.mustObserve(t, safetyObservation{kind: "create", eventID: "thermal_1", alert: alertcontract.AlertPayload{EventID: "thermal_1", Source: "host_1"}})
	flow.mustObserve(t, safetyObservation{kind: "create", eventID: "power_1", alert: alertcontract.AlertPayload{EventID: "power_1", Source: "host_1"}})
	if got := flow.publicationCountAfterFlush(t); got != 0 {
		t.Fatalf("out-of-order publications = %d", got)
	}
}

// Requirements: operational-safety-automation/operator-surface.
func TestOperationalSafetyExpiredSequenceDoesNotEmit(t *testing.T) {
	flow := mustSafetyFlow(t)
	flow.mustObserve(t, safetyObservation{kind: "create", eventID: "power_1", alert: alertcontract.AlertPayload{EventID: "power_1", Source: "host_1"}})
	flow.now = flow.now.Add(61 * time.Second)
	flow.mustObserve(t, safetyObservation{kind: "create", eventID: "thermal_1", alert: alertcontract.AlertPayload{EventID: "thermal_1", Source: "host_1"}})
	if got := flow.publicationCountAfterFlush(t); got != 0 {
		t.Fatalf("expired sequence publications = %d", got)
	}
}

// Requirements: operational-safety-automation/operator-surface.
func TestOperationalSafetyFailedCapDoesNotCommitFSM(t *testing.T) {
	flow := mustSafetyFlow(t)
	flow.runtime.apply = func(context.Context, uint64, uint64) error { return errors.New("simulated sysfs failure") }
	flow.mustObserve(t, safetyObservation{kind: "create", eventID: "power_1", alert: alertcontract.AlertPayload{EventID: "power_1", Source: "host_1"}})
	flow.mustObserve(t, safetyObservation{kind: "create", eventID: "thermal_1", alert: alertcontract.AlertPayload{EventID: "thermal_1", Source: "host_1"}})
	flow.mustPublication(t)
	flow.mustFailObservation(t, safetyObservation{kind: "create", eventID: "opsafe_1", alert: alertcontract.AlertPayload{EventID: "opsafe_1", Source: "operational-safety-automation"}})
	got := struct {
		count   int
		level   contract.OperationalSafetyLevel
		pending bool
		lastCap uint64
	}{count: len(flow.runtime.tasks), lastCap: flow.runtime.lastCap}
	for _, task := range flow.runtime.tasks {
		got.level, got.pending = task.level, task.pending
	}
	want := struct {
		count   int
		level   contract.OperationalSafetyLevel
		pending bool
		lastCap uint64
	}{count: 1, level: contract.LevelNormal}
	if got != want {
		t.Fatalf("failed effect state = %#v, want %#v", got, want)
	}
}

// Requirements: operational-safety-automation/operator-surface.
func TestOperationalSafetyRenewedTriggerCancelsCooldown(t *testing.T) {
	flow := mustSafetyFlow(t)
	flow.mustStartCriticalTask(t)
	flow.mustObserve(t, safetyObservation{kind: "state", eventID: "opsafe_1", state: "normal"})
	flow.now = flow.now.Add(30 * time.Second)
	flow.mustObserve(t, safetyObservation{kind: "create", eventID: "opsafe_2", alert: alertcontract.AlertPayload{EventID: "opsafe_2", Source: "operational-safety-automation"}})
	flow.now = flow.now.Add(2 * time.Minute)
	flow.runtime.advanceCooldowns(context.Background())
	flow.mu.Lock()
	got := append([]uint64(nil), flow.caps...)
	flow.mu.Unlock()
	if !reflect.DeepEqual(got, []uint64{1000000}) {
		t.Fatalf("CPU cap requests after renewed trigger = %v", got)
	}
}

// Requirements: operational-safety-automation/operator-surface.
func TestOperationalSafetyCombinedConditionsEmitAfterLastFieldArrives(t *testing.T) {
	flow := mustSafetyFlow(t)
	enabled := true
	minimum := 80.0
	step := selector{IDPrefix: "thermal", State: "active"}
	step.Message.Contains = "CPU"
	step.Value.NumGTE = &minimum
	spec := ruleSpec{ID: "thermal_condition", Name: "Thermal condition", Enabled: &enabled, CorrelateBy: "event_id", Steps: []selector{step}}
	spec.Action.Kind = "emit"
	spec.FSM.Type = detection.CountWithinWindow
	spec.FSM.Window = "1m"
	spec.FSM.Threshold = 1
	spec.FSM.MaxKeys = 16
	spec.FSM.MaxIdle = "10m"
	spec.Output.IDPrefix = "hotcpu"
	spec.Output.Category = "system"
	spec.Output.Severity = "warning"
	spec.Output.Message = "CPU temperature high"
	rule, err := compileRuleWithClock(spec, func() time.Time { return flow.now })
	if err != nil {
		t.Fatal(err)
	}
	flow.runtime.rules = map[string]*compiledRule{spec.ID: rule}
	flow.runtime.order = []string{spec.ID}
	value := 85.0
	flow.mustObserve(t, safetyObservation{kind: "occurrence", eventID: "thermal_1", alert: alertcontract.AlertPayload{}, occurrence: alertcontract.AlertOccurrencePayload{EventID: "thermal_1", ValueNum: &value}})
	flow.mustObserve(t, safetyObservation{kind: "state", eventID: "thermal_1", state: "active"})
	flow.mustObserve(t, safetyObservation{kind: "create", eventID: "thermal_1", alert: alertcontract.AlertPayload{EventID: "thermal_1", Source: "host_1", Message: "CPU hot"}})
	if got := flow.mustPublication(t); got.kind != "emit" || got.ruleID != spec.ID {
		t.Fatalf("combined-condition publication = %#v", got)
	}
}

// mustOperatorTask prepares a live security token for the existing flow task.
func (flow *safetyFlow) mustOperatorTask(t *testing.T) (string, string) {
	t.Helper()
	operator, token := mustSafetyOperator(t, []string{"lite-nas-security"})
	flow.runtime.verifier = operator.verifier
	flow.runtime.client = operator.client
	for id := range flow.runtime.tasks {
		return id, token
	}
	t.Fatal("flow did not create a task")
	return "", ""
}

// Requirements: operational-safety-automation/operator-surface.
func TestOperationalSafetyForceStopBlocksActiveAlert(t *testing.T) {
	flow := mustSafetyFlow(t)
	flow.mustStartCriticalTask(t)
	id, token := flow.mustOperatorTask(t)
	response, err := flow.runtime.forceStop(context.Background(), mustSafetyEnvelope(t, contract.OperationalSafetyTaskForceStopRequest{AccessToken: token, TaskID: id}))
	if err != nil {
		t.Fatal(err)
	}
	if got := response.(contract.OperationalSafetyOperationResponse).State; got != contract.OutcomeBlocked {
		t.Fatalf("force stop state = %q", got)
	}
}

// Requirements: operational-safety-automation/operator-surface.
func TestOperationalSafetyForceSetReappliesVerifiedCriticalCap(t *testing.T) {
	flow := mustSafetyFlow(t)
	flow.mustStartCriticalTask(t)
	id, token := flow.mustOperatorTask(t)
	flow.mustObserve(t, safetyObservation{kind: "state", eventID: "opsafe_1", state: "normal"})
	flow.now = flow.now.Add(time.Minute)
	flow.runtime.advanceCooldowns(context.Background())
	flow.mustPublication(t)
	response, err := flow.runtime.forceSet(context.Background(), mustSafetyEnvelope(t, contract.OperationalSafetyTaskForceSetRequest{AccessToken: token, TaskID: id, Level: contract.LevelCritical}))
	if err != nil {
		t.Fatal(err)
	}
	flow.mustPublication(t)
	flow.mu.Lock()
	got := struct {
		state contract.OperationalSafetyOutcomeState
		caps  []uint64
	}{state: response.(contract.OperationalSafetyOperationResponse).State, caps: append([]uint64(nil), flow.caps...)}
	flow.mu.Unlock()
	want := struct {
		state contract.OperationalSafetyOutcomeState
		caps  []uint64
	}{state: contract.OutcomeCompleted, caps: []uint64{1000000, 1200000, 1000000}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("force set outcome = %#v, want %#v", got, want)
	}
}
