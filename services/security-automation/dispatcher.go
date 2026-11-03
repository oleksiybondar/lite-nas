package main

import (
	"context"
	"errors"
	"fmt"

	"lite-nas/services/security-automation/detection"
	"lite-nas/shared/workers"
)

// observation is the validated and correlated input to one configured rule.
// The transport adapter owns the input channel and closes it after NATS drain.
type observation struct {
	RuleID string
	Event  detection.Event
}

// detectionIntent is an immutable completed match for the outbound worker.
type detectionIntent struct {
	RuleID string
	Match  detection.Match
}

// startDispatcher runs the detection stage with the shared worker primitive.
// The caller owns input and must close it after all callback producers stop.
// This worker owns output, closes it after input drains, and reports all errors
// through the required handler. A blocked output send honors cancellation and
// reports the lost intent instead of silently discarding it.
func startDispatcher(ctx context.Context, input <-chan observation, ticks <-chan struct{},
	plugins map[string]*detection.Plugin, outputCapacity int, report func(error),
) (<-chan detectionIntent, <-chan struct{}, error) {
	if input == nil || plugins == nil || outputCapacity < 1 || report == nil {
		return nil, nil, errors.New("invalid automation dispatcher configuration")
	}
	for id, plugin := range plugins {
		if id == "" || plugin == nil {
			return nil, nil, errors.New("invalid automation plugin registry")
		}
	}
	output := make(chan detectionIntent, outputCapacity)
	done := make(chan struct{})
	workers.NewBaseWorker(func(workerCtx context.Context) {
		defer close(done)
		defer close(output)
		shutdown := workerCtx.Done()
		for input != nil {
			select {
			case item, ok := <-input:
				if !ok {
					input = nil
					continue
				}
				plugin := plugins[item.RuleID]
				if plugin == nil {
					report(fmt.Errorf("unknown automation rule %q", item.RuleID))
					continue
				}
				matches, err := plugin.Observe(workerCtx, item.Event)
				if err != nil {
					report(fmt.Errorf("automation rule %q: %w", item.RuleID, err))
					continue
				}
				for _, match := range matches {
					intent := detectionIntent{RuleID: item.RuleID, Match: match}
					select {
					case output <- intent:
					case <-workerCtx.Done():
						report(fmt.Errorf("automation intent for rule %q and key %q was not queued: %w", item.RuleID, match.Key, workerCtx.Err()))
					}
				}
			case _, ok := <-ticks:
				if !ok {
					ticks = nil
					continue
				}
				for id, plugin := range plugins {
					if err := plugin.Advance(workerCtx); err != nil {
						report(fmt.Errorf("automation rule %q deadline: %w", id, err))
					}
				}
			case <-shutdown:
				// The channel owner must stop producers and close input. This stage
				// keeps draining already accepted observations before closing output.
				ticks = nil
				shutdown = nil
			}
		}
	}).Start(ctx)
	return output, done, nil
}
