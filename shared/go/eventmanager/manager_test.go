package eventmanager_test

import (
	"errors"
	"testing"

	"lite-nas/shared/eventmanager"
)

func TestBuildKeyReturnsEventFieldConditionKey(t *testing.T) {
	t.Parallel()

	got := eventmanager.BuildKey("system.metrics.events.stats", "snapshot.cpu.totalUsagePct", ">=")
	want := "system.metrics.events.stats:snapshot.cpu.totalUsagePct:>="
	if got != want {
		t.Fatalf("BuildKey() = %q, want %q", got, want)
	}
}

func TestBuildKeyAppendsQualifiers(t *testing.T) {
	t.Parallel()

	got := eventmanager.BuildKey("zfs.metrics.events.snapshot", "snapshot.Pools[].Health", "==", "1", "0")
	want := "zfs.metrics.events.snapshot:snapshot.Pools[].Health:==:1:0"
	if got != want {
		t.Fatalf("BuildKey() = %q, want %q", got, want)
	}
}

func TestCreateEventAndFindEvent(t *testing.T) {
	t.Parallel()

	manager := eventmanager.NewManager(0)
	payload := map[string]any{"event_id": "syscpu00000001"}

	if err := manager.CreateEvent("system.metrics.events.stats", "snapshot.cpu.totalUsagePct", ">=", payload); err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}

	got, exists := manager.FindEvent("system.metrics.events.stats", "snapshot.cpu.totalUsagePct", ">=")
	if !exists {
		t.Fatal("FindEvent() exists = false, want true")
	}

	assertFoundEventFields(t, got)
	assertFoundEventPayload(t, got.Payload)
}

func TestCreateEventAndFindEventWithQualifiers(t *testing.T) {
	t.Parallel()

	manager := eventmanager.NewManager(0)
	payload := map[string]any{"event_id": "zfspool00000001"}

	if err := manager.CreateEvent("zfs.metrics.events.snapshot", "snapshot.Pools[].Health", "==", payload, "1"); err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}

	got, exists := manager.FindEvent("zfs.metrics.events.snapshot", "snapshot.Pools[].Health", "==", "1")
	if !exists {
		t.Fatal("FindEvent() exists = false, want true")
	}

	if len(got.Qualifiers) != 1 || got.Qualifiers[0] != "1" {
		t.Fatalf("got.Qualifiers = %v, want [1]", got.Qualifiers)
	}
}

func TestCreateEventRejectsDuplicateKey(t *testing.T) {
	t.Parallel()

	manager := eventmanager.NewManager(0)
	if err := manager.CreateEvent("system.metrics.events.stats", "snapshot.cpu.totalUsagePct", ">=", nil); err != nil {
		t.Fatalf("CreateEvent() first error = %v", err)
	}

	err := manager.CreateEvent("system.metrics.events.stats", "snapshot.cpu.totalUsagePct", ">=", nil)
	if !errors.Is(err, eventmanager.ErrEventAlreadyExists) {
		t.Fatalf("CreateEvent() duplicate error = %v, want %v", err, eventmanager.ErrEventAlreadyExists)
	}
}

func TestDeleteEventRemovesCachedEntry(t *testing.T) {
	t.Parallel()

	manager := eventmanager.NewManager(0)
	if err := manager.CreateEvent("system.metrics.events.stats", "snapshot.mem.usedPct", ">=", nil); err != nil {
		t.Fatalf("CreateEvent() error = %v", err)
	}

	manager.DeleteEvent("system.metrics.events.stats", "snapshot.mem.usedPct", ">=")

	_, exists := manager.FindEvent("system.metrics.events.stats", "snapshot.mem.usedPct", ">=")
	if exists {
		t.Fatal("FindEvent() exists = true, want false")
	}
}

func TestCounterOperations(t *testing.T) {
	t.Parallel()

	manager := eventmanager.NewManager(41)

	if got := manager.GetCounter(); got != 41 {
		t.Fatalf("GetCounter() = %d, want 41", got)
	}

	manager.SetCounter(99)
	if got := manager.GetCounter(); got != 99 {
		t.Fatalf("GetCounter() after SetCounter = %d, want 99", got)
	}

	if got := manager.NextCounter(); got != 100 {
		t.Fatalf("NextCounter() = %d, want 100", got)
	}

	if got := manager.GetCounter(); got != 100 {
		t.Fatalf("GetCounter() after NextCounter = %d, want 100", got)
	}
}

func assertFoundEventFields(t *testing.T, got eventmanager.Event) {
	t.Helper()

	if got.Event != "system.metrics.events.stats" {
		t.Fatalf("got.Event = %q, want %q", got.Event, "system.metrics.events.stats")
	}
	if got.Field != "snapshot.cpu.totalUsagePct" {
		t.Fatalf("got.Field = %q, want %q", got.Field, "snapshot.cpu.totalUsagePct")
	}
	if got.Condition != ">=" {
		t.Fatalf("got.Condition = %q, want %q", got.Condition, ">=")
	}
}

func assertFoundEventPayload(t *testing.T, payload any) {
	t.Helper()

	gotPayload, ok := payload.(map[string]any)
	if !ok {
		t.Fatalf("got.Payload type = %T, want map[string]any", payload)
	}
	if gotPayload["event_id"] != "syscpu00000001" {
		t.Fatalf("got payload event_id = %v, want %q", gotPayload["event_id"], "syscpu00000001")
	}
}
