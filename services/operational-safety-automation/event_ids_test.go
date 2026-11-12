package main

import (
	"testing"

	"lite-nas/shared/loggingmanager"
)

// Requirements: operational-safety-automation/operator-surface.
func TestEventIDAllocatorContinuesAfterRestart(t *testing.T) {
	path := t.TempDir()
	first, err := newEventIDAllocator(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := first.Next("opsafe"); err != nil {
		t.Fatal(err)
	}
	restarted, err := newEventIDAllocator(path)
	if err != nil {
		t.Fatal(err)
	}
	id, err := restarted.Next("opsafe")
	if err != nil {
		t.Fatal(err)
	}
	if id != "opsafe_1" {
		t.Fatalf("ID after restart = %q", id)
	}
}

// Requirements: operational-safety-automation/operator-surface.
func TestEventIDAllocatorProducesLoggingManagerID(t *testing.T) {
	allocator, err := newEventIDAllocator(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	id, err := allocator.Next("opsafe")
	if err != nil {
		t.Fatal(err)
	}
	validate, err := loggingmanager.NewInputValidator()
	if err != nil {
		t.Fatal(err)
	}
	if err := validate.Struct(struct {
		EventID string `validate:"required,loggingmanager_event_id"`
	}{EventID: id}); err != nil {
		t.Fatalf("invalid generated ID: %v", err)
	}
}
