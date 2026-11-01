package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"lite-nas/shared/contracts/networksecurity"
	"lite-nas/shared/fsm"
)

type (
	scanState string
	scanInput string
)

const (
	scanIdle        scanState = "idle"
	scanCounting    scanState = "counting"
	scanCooling     scanState = "cooling"
	probeObserved   scanInput = "probe"
	windowElapsed   scanInput = "window_elapsed"
	cooldownElapsed scanInput = "cooldown_elapsed"
)

// scanData is volatile correlation state for one source and destination pair.
type scanData struct {
	Ports         map[string]struct{}
	Port          string
	Count         int
	TargetPort    uint16
	At            time.Time
	WindowEnds    time.Time
	CooldownEnds  time.Time
	SourceIP      string
	DestinationIP string
}

// scanDetection is a typed alert intent produced by the FSM.
type scanDetection struct {
	Kind            string
	SourceIP        string
	DestinationIP   string
	DestinationPort uint16
	PortCount       int
	At              time.Time
}

// portScanPlugin detects several distinct blocked TCP destination ports from
// one source to one destination in a configured window. The MemoryStore is
// bounded and volatile; a process restart resets correlation and cooldown.
type portScanPlugin struct {
	kind      string
	store     *fsm.MemoryStore[scanState, scanInput, scanData, scanDetection]
	now       func() time.Time
	threshold int
	window    time.Duration
	cooldown  time.Duration
}

// newPortScanPlugin compiles the detector and its transition table at startup.
func newPortScanPlugin(threshold, maxPairs int, window, cooldown time.Duration, now func() time.Time) (*portScanPlugin, error) {
	return newPatternPlugin("port_scan", threshold, maxPairs, window, cooldown, now)
}

// newPatternPlugin compiles one configured blocked-attempt pattern at startup.
func newPatternPlugin(kind string, threshold, maxPairs int, window, cooldown time.Duration, now func() time.Time) (*portScanPlugin, error) {
	if threshold < 2 || threshold > 1024 || maxPairs < 1 || maxPairs > 16384 ||
		window < time.Second || window > time.Hour || cooldown < time.Second || cooldown > 24*time.Hour || now == nil ||
		(kind != "port_scan" && kind != "horizontal_scan" && kind != "repeated_attempts") {
		return nil, errors.New("invalid port scan plugin configuration")
	}
	p := &portScanPlugin{kind: kind, now: now, threshold: threshold, window: window, cooldown: cooldown}
	machine, err := p.machine()
	if err != nil {
		return nil, err
	}
	p.store, err = fsm.NewMemoryStore(machine, maxPairs, window+cooldown+time.Minute,
		func(string) scanData { return scanData{} }, now, p.deadline)
	return p, err
}

// Observe validates the domain input and advances one keyed FSM. The caller
// accounts for every returned effect before accepting another observation.
func (p *portScanPlugin) Observe(ctx context.Context, attempt networksecurity.FirewallAttempt) ([]scanDetection, error) {
	if attempt.SourceIP == "" || attempt.DestinationIP == "" || attempt.DestinationPort == 0 || attempt.Protocol != "TCP" {
		return nil, errors.New("invalid firewall attempt")
	}
	key := attempt.SourceIP + "/" + attempt.DestinationIP
	item := strconv.Itoa(int(attempt.DestinationPort))
	switch p.kind {
	case "horizontal_scan":
		key = attempt.SourceIP + "/" + item
		item = attempt.DestinationIP
	case "repeated_attempts":
		key += "/" + item
	}
	result, err := p.store.ApplyWith(ctx, key, probeObserved, func(data scanData) scanData {
		data.Port = item
		data.TargetPort = attempt.DestinationPort
		data.At = p.now()
		data.SourceIP = attempt.SourceIP
		data.DestinationIP = attempt.DestinationIP
		return data
	})
	if err != nil {
		return nil, fmt.Errorf("port scan transition: %w", err)
	}
	return result.Effects, nil
}

// Advance applies due timeout inputs and expires idle keys. The coordinator
// invokes it on a timer and before processing each new observation.
func (p *portScanPlugin) Advance(ctx context.Context) error {
	for _, due := range p.store.Due() {
		if _, err := p.store.Apply(ctx, due.Key, due.Input); err != nil {
			return err
		}
	}
	p.store.ExpireIdle()
	return nil
}

// deadline selects the next timeout input for a tracked pair.
func (p *portScanPlugin) deadline(instance fsm.Instance[scanState, scanData]) (time.Time, scanInput, bool) {
	switch instance.State {
	case scanCounting:
		return instance.Data.WindowEnds, windowElapsed, true
	case scanCooling:
		return instance.Data.CooldownEnds, cooldownElapsed, true
	default:
		return time.Time{}, "", false
	}
}

// machine compiles the explicit count, detect, and cooldown transitions.
func (p *portScanPlugin) machine() (*fsm.Machine[scanState, scanInput, scanData, scanDetection], error) {
	guards := map[string]fsm.Guard[scanData, scanInput]{
		"new_below": func(_ context.Context, data scanData, _ scanInput) (bool, error) {
			_, seen := data.Ports[data.Port]
			return (p.kind == "repeated_attempts" || !seen) && data.Count+1 < p.threshold, nil
		},
		"new_reached": func(_ context.Context, data scanData, _ scanInput) (bool, error) {
			_, seen := data.Ports[data.Port]
			return (p.kind == "repeated_attempts" || !seen) && data.Count+1 >= p.threshold, nil
		},
		"window_due": func(_ context.Context, data scanData, _ scanInput) (bool, error) {
			return !p.now().Before(data.WindowEnds), nil
		},
		"cooldown_due": func(_ context.Context, data scanData, _ scanInput) (bool, error) {
			return !p.now().Before(data.CooldownEnds), nil
		},
	}
	actions := map[string]fsm.Action[scanData, scanInput, scanDetection]{
		"start": func(_ context.Context, data scanData, _ scanInput) (scanData, []scanDetection, error) {
			data.Ports = map[string]struct{}{data.Port: {}}
			data.Count = 1
			data.WindowEnds = data.At.Add(p.window)
			return data, nil, nil
		},
		"add": func(_ context.Context, data scanData, _ scanInput) (scanData, []scanDetection, error) {
			return actionsAddPort(data)
		},
		"detect": func(_ context.Context, data scanData, _ scanInput) (scanData, []scanDetection, error) {
			data, _, _ = actionsAddPort(data)
			data.CooldownEnds = data.At.Add(p.cooldown)
			return data, []scanDetection{{Kind: p.kind, SourceIP: data.SourceIP, DestinationIP: data.DestinationIP, DestinationPort: data.TargetPort, PortCount: data.Count, At: data.At}}, nil
		},
		"clear": func(_ context.Context, _ scanData, _ scanInput) (scanData, []scanDetection, error) {
			return scanData{}, nil, nil
		},
	}
	rows := []fsm.Row[scanState, scanInput]{
		{From: scanIdle, On: probeObserved, To: scanCounting, Action: "start"},
		{From: scanCounting, On: probeObserved, Guard: "new_below", To: scanCounting, Action: "add"},
		{From: scanCounting, On: probeObserved, Guard: "new_reached", To: scanCooling, Action: "detect"},
		{From: scanCounting, On: windowElapsed, Guard: "window_due", To: scanIdle, Action: "clear"},
		{From: scanCooling, On: probeObserved, To: scanCooling},
		{From: scanCooling, On: cooldownElapsed, Guard: "cooldown_due", To: scanIdle, Action: "clear"},
	}
	return fsm.New(fsm.Definition[scanState, scanInput, scanData, scanDetection]{
		Initial: scanIdle, States: []scanState{scanIdle, scanCounting, scanCooling},
		Inputs: []scanInput{probeObserved, windowElapsed, cooldownElapsed},
		Rows:   rows, Guards: guards, Actions: actions,
	})
}

// actionsAddPort copies the port set before an FSM action mutates it.
func actionsAddPort(data scanData) (scanData, []scanDetection, error) {
	ports := make(map[string]struct{}, len(data.Ports)+1)
	for port := range data.Ports {
		ports[port] = struct{}{}
	}
	ports[data.Port] = struct{}{}
	data.Ports = ports
	data.Count++
	return data, nil, nil
}
