package main

import (
	"context"
	"testing"
	"time"

	"lite-nas/services/security-automation/detection"
)

// TestDispatcherPipesCompletedMatch verifies the worker hands a complete FSM
// result to the outbound queue and closes that queue after its producer stops.
func TestDispatcherPipesCompletedMatch(t *testing.T) {
	now := time.Now()
	plugin, err := detection.New(detection.Config{
		Name: detection.CountWithinWindow, Threshold: 2, Window: time.Minute,
		MaxKeys: 8, MaxIdle: time.Hour,
	}, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	input := make(chan observation, 2)
	output, done, err := startDispatcher(context.Background(), input, nil, map[string]*detection.Plugin{"rule": plugin}, 2, func(err error) { t.Error(err) })
	if err != nil {
		t.Fatal(err)
	}
	input <- observation{RuleID: "rule", Event: detection.Event{ID: "first", Key: "alice", At: now}}
	input <- observation{RuleID: "rule", Event: detection.Event{ID: "second", Key: "alice", At: now}}
	close(input)
	select {
	case intent := <-output:
		if intent.RuleID != "rule" || intent.Match.Key != "alice" {
			t.Fatalf("intent = %+v", intent)
		}
	case <-time.After(time.Second):
		t.Fatal("dispatcher did not emit completed match")
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("dispatcher did not stop after input closed")
	}
}

// Requirements: security-automation/IR-002.
// TestDispatcherRejectsInvalidConfiguration checks the worker cannot start without bounded channels and plugins.
func TestDispatcherRejectsInvalidConfiguration(t *testing.T) {
	cases := []struct {
		name     string
		input    chan observation
		plugins  map[string]*detection.Plugin
		capacity int
		report   func(error)
	}{
		{"nil input", nil, map[string]*detection.Plugin{"rule": {}}, 1, func(error) {}},
		{"nil plugins", make(chan observation), nil, 1, func(error) {}},
		{"unbounded output", make(chan observation), map[string]*detection.Plugin{"rule": {}}, 0, func(error) {}},
		{"no error handler", make(chan observation), map[string]*detection.Plugin{"rule": {}}, 1, nil},
		{"empty rule ID", make(chan observation), map[string]*detection.Plugin{"": {}}, 1, func(error) {}},
		{"nil plugin", make(chan observation), map[string]*detection.Plugin{"rule": nil}, 1, func(error) {}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			_, _, err := startDispatcher(context.Background(), test.input, nil, test.plugins, test.capacity, test.report)
			if err == nil {
				t.Fatal("invalid dispatcher configuration accepted")
			}
		})
	}
}

// Requirements: security-automation/IR-002.
// TestDispatcherReportsUnknownRule checks a valid but unconfigured observation reaches the error path.
func TestDispatcherReportsUnknownRule(t *testing.T) {
	now := time.Now()
	plugin, err := detection.New(detection.Config{Name: detection.CountWithinWindow, Threshold: 1, Window: time.Minute, MaxKeys: 8, MaxIdle: time.Hour}, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	input := make(chan observation, 1)
	reported := make(chan error, 1)
	_, done, err := startDispatcher(context.Background(), input, nil, map[string]*detection.Plugin{"known": plugin}, 1, func(err error) { reported <- err })
	if err != nil {
		t.Fatal(err)
	}
	input <- observation{RuleID: "unknown", Event: detection.Event{ID: "sec_1", Key: "alice", At: now}}
	close(input)
	select {
	case err := <-reported:
		if err == nil {
			t.Fatal("unknown rule was not reported")
		}
	case <-done:
		select {
		case <-reported:
		default:
			t.Fatal("unknown rule was not reported")
		}
	case <-time.After(time.Second):
		t.Fatal("dispatcher did not report unknown rule")
	}
}
