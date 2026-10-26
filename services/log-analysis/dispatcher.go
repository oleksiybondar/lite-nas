package main

import (
	"context"
	"fmt"

	"lite-nas/services/log-analysis/plugins"
	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/workers"
)

// routeIndex groups precompiled rules by physical source and category. The
// stored plugin references are fully initialized before any listener starts.
type routeIndex map[string]map[string][]compiledRule

// buildRoutes fails if a rule names a plugin that startup did not construct.
func buildRoutes(spec rulesFile, instances map[string]*plugins.Plugin) (routeIndex, error) {
	routes := make(routeIndex)
	for _, rule := range spec.Rules {
		if !rule.Enabled {
			continue
		}
		plugin := instances[rule.Instance]
		if plugin == nil {
			return nil, fmt.Errorf("rule %s references missing plugin instance %s", rule.ID, rule.Instance)
		}
		compiled, err := compileRulePatterns(rule)
		if err != nil {
			return nil, err
		}
		compiled.plugin = plugin
		if routes[rule.Source] == nil {
			routes[rule.Source] = make(map[string][]compiledRule)
		}
		routes[rule.Source][rule.Category] = append(routes[rule.Source][rule.Category], compiled)
	}
	return routes, nil
}

// dispatchWorker selects cheap RE2 matches, normalizes them and sends one copy
// to every applicable initialized plugin. Its input is a single work stream;
// independent plugin channels provide the required broadcast behavior.
type dispatchWorker struct {
	input     <-chan logRecord
	routes    routeIndex
	instances map[string]*plugins.Plugin
	worker    workers.BaseWorker
	done      chan struct{}
	log       sharedlogger.Logger
}

// newDispatchWorker receives the already constructed plugin collection.
func newDispatchWorker(input <-chan logRecord, routes routeIndex, instances map[string]*plugins.Plugin, log sharedlogger.Logger) *dispatchWorker {
	dispatcher := &dispatchWorker{input: input, routes: routes, instances: instances, done: make(chan struct{}), log: log}
	dispatcher.worker = workers.NewBaseWorker(dispatcher.run)
	return dispatcher
}

// Start launches the dispatcher under the service context.
func (d *dispatchWorker) Start(ctx context.Context) { d.worker.Start(ctx) }

// Done closes once all records have been fanned out and plugin inputs closed.
func (d *dispatchWorker) Done() <-chan struct{} { return d.done }

func (d *dispatchWorker) run(_ context.Context) {
	defer close(d.done)
	defer func() {
		for _, plugin := range d.instances {
			close(plugin.Input())
		}
	}()
	for record := range d.input {
		for _, categoryRules := range d.routes[record.SourceID] {
			for _, rule := range categoryRules {
				kind := "failure"
				if rule.success != nil && rule.success.MatchString(record.Message) {
					kind = "success"
				} else if !rule.matches(record.Message) {
					continue
				}
				observation := plugins.Observation{
					RuleID: rule.spec.ID, Category: rule.spec.Category,
					Severity: rule.spec.Severity, Email: rule.spec.EmailOnCritical, Lockdown: rule.spec.Lockdown, Kind: kind,
					Account: record.Account, SourceAddress: record.SourceAddress,
					Service: record.Service, At: record.At,
				}
				rule.plugin.Input() <- observation
			}
		}
	}
}

func (rule compiledRule) matches(message string) bool {
	if !rule.match.MatchString(message) {
		return false
	}
	for _, required := range rule.require {
		if !required.MatchString(message) {
			return false
		}
	}
	for _, excluded := range rule.exclude {
		if excluded.MatchString(message) {
			return false
		}
	}
	return true
}
