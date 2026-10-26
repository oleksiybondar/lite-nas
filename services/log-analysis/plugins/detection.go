// Package plugins contains the statically compiled detection workers for Log
// Analysis. Construction happens at startup, outside listeners and dispatch.
package plugins

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"lite-nas/shared/fsm"
	sharedlogger "lite-nas/shared/logger"
	"lite-nas/shared/workers"
)

// Observation contains bounded normalized evidence selected by one rule.
type Observation struct {
	RuleID        string
	Category      string
	Severity      string
	Email         bool
	Lockdown      bool
	Kind          string
	Account       string
	SourceAddress string
	Service       string
	At            time.Time
}

// Detection is a typed manager or lockdown effect; plugin workers do not
// deliver it directly to NATS or email.
type Detection struct {
	InstanceID    string
	RuleID        string
	Category      string
	Severity      string
	Email         bool
	Lockdown      bool
	Operation     string
	EventID       string
	Account       string
	SourceAddress string
	At            time.Time
}

// Config specifies one validated plugin instance constructed at startup.
type Config struct {
	ID        string        `json:"id"`
	Type      string        `json:"type"`
	Threshold int           `json:"threshold"`
	Window    time.Duration `json:"-"`
	Cooldown  time.Duration `json:"-"`
	Idle      time.Duration `json:"-"`
	KeyBy     string        `json:"key_by"`
	MaxKeys   int           `json:"max_keys"`
}

type (
	state string
	input string
)

const (
	idle            state = "idle"
	counting        state = "counting"
	cooldown        state = "cooldown"
	observed        input = "observed"
	windowExpired   input = "window_expired"
	cooldownExpired input = "cooldown_expired"
	idleExpired     input = "idle_expired"
	succeeded       input = "succeeded"
)

// instanceData holds detection counters separately from manager lifecycle
// identity. The instance store is volatile and cannot claim restart survival.
type instanceData struct {
	Count            int
	WindowDeadline   time.Time
	CooldownDeadline time.Time
	IdleDeadline     time.Time
	EventID          string
	Last             Observation
}

// Plugin is one configured worker with its own bounded observation channel,
// FSM and deadline tick. Multiple instances of the same type are independent.
type Plugin struct {
	id     string
	cfg    Config
	input  chan Observation
	output chan<- Detection
	store  *fsm.MemoryStore[state, input, instanceData, Detection]
	timer  workers.TimerWorker
	ticks  <-chan struct{}
	worker workers.BaseWorker
	done   chan struct{}
	log    sharedlogger.Logger
}

// NewImmediate constructs the compiled one-observation detector.
func NewImmediate(cfg Config, output chan<- Detection, log sharedlogger.Logger) (*Plugin, error) {
	cfg.Type = "immediate"
	return New(cfg, output, log)
}

// NewThreshold constructs the compiled global threshold detector.
func NewThreshold(cfg Config, output chan<- Detection, log sharedlogger.Logger) (*Plugin, error) {
	cfg.Type = "threshold"
	return New(cfg, output, log)
}

// NewKeyedThreshold constructs the compiled per-account or per-source detector.
func NewKeyedThreshold(cfg Config, output chan<- Detection, log sharedlogger.Logger) (*Plugin, error) {
	cfg.Type = "keyed_threshold"
	return New(cfg, output, log)
}

// New constructs a configured, statically compiled immediate or threshold
// plugin. The composition root selects this constructor from its fixed map.
func New(cfg Config, output chan<- Detection, log sharedlogger.Logger) (*Plugin, error) {
	if cfg.ID == "" || len(cfg.ID) > 64 || strings.ContainsAny(cfg.ID, " /\\\t\n\r") {
		return nil, errors.New("invalid plugin instance ID")
	}
	if cfg.Type != "immediate" && cfg.Type != "threshold" && cfg.Type != "keyed_threshold" {
		return nil, fmt.Errorf("uncompiled plugin type %q", cfg.Type)
	}
	if cfg.Type == "immediate" {
		cfg.Threshold = 1
	}
	if cfg.Threshold < 1 || cfg.Threshold > 1000 || cfg.Window < time.Second || cfg.Window > 24*time.Hour || cfg.Cooldown < time.Second || cfg.Cooldown > 24*time.Hour || cfg.Idle <= cfg.Cooldown || cfg.Idle > 7*24*time.Hour {
		return nil, errors.New("invalid plugin threshold or deadlines")
	}
	if cfg.Type == "threshold" && cfg.KeyBy != "global" {
		return nil, errors.New("threshold plugin requires global key")
	}
	if cfg.Type == "keyed_threshold" && cfg.KeyBy != "account" && cfg.KeyBy != "source" {
		return nil, errors.New("keyed threshold requires account or source key")
	}
	if cfg.Type == "immediate" && cfg.KeyBy != "global" {
		return nil, errors.New("immediate plugin requires global key")
	}
	if cfg.MaxKeys < 1 || cfg.MaxKeys > 4096 || output == nil || log == nil {
		return nil, errors.New("invalid plugin capacity or dependencies")
	}
	plugin := &Plugin{id: cfg.ID, cfg: cfg, input: make(chan Observation, 32), output: output, done: make(chan struct{}), log: log}
	machine, err := plugin.newMachine()
	if err != nil {
		return nil, err
	}
	plugin.store, err = fsm.NewMemoryStore(machine, cfg.MaxKeys, cfg.Idle*2,
		func(string) instanceData { return instanceData{} }, time.Now, plugin.deadline)
	if err != nil {
		return nil, err
	}
	plugin.timer, plugin.ticks, err = workers.NewPollingTimerWorker(time.Second, 1)
	if err != nil {
		return nil, err
	}
	plugin.worker = workers.NewBaseWorker(plugin.run)
	return plugin, nil
}

// ID is the stable configured instance identifier.
func (p *Plugin) ID() string { return p.id }

// Input accepts validated observations; the dispatcher owns closing it.
func (p *Plugin) Input() chan<- Observation { return p.input }

// Done closes after all queued observations and deadline inputs are handled.
func (p *Plugin) Done() <-chan struct{} { return p.done }

// Start launches this plugin's timer and processing worker.
func (p *Plugin) Start(ctx context.Context) { p.timer.Start(ctx); p.worker.Start(ctx) }

func (p *Plugin) run(ctx context.Context) {
	defer close(p.done)
	inputChannel := (<-chan Observation)(p.input)
	ticks := p.ticks
	shutdown := ctx.Done()
	for inputChannel != nil {
		select {
		case observation, ok := <-inputChannel:
			if !ok {
				inputChannel = nil
				continue
			}
			processCtx := ctx
			if ctx.Err() != nil {
				processCtx = context.Background()
			}
			p.process(processCtx, observation)
		case <-ticks:
			p.processDeadlines(ctx)
		case <-shutdown:
			// The runtime closes the input only after listeners and dispatcher
			// stop; keep consuming queued security observations on shutdown.
			ticks = nil
			shutdown = nil
		}
	}
}

func (p *Plugin) process(ctx context.Context, observation Observation) {
	p.processDeadlines(ctx)
	key, err := p.key(observation)
	if err != nil {
		p.log.Warn("plugin observation lacks required correlation key", "plugin", p.id, "rule", observation.RuleID, "error", err)
		return
	}
	event := observed
	if observation.Kind == "success" {
		event = succeeded
	}
	result, err := p.store.ApplyWith(ctx, key, event, func(data instanceData) instanceData { data.Last = observation; return data })
	if err != nil {
		p.log.Error("plugin transition failed", "plugin", p.id, "rule", observation.RuleID, "error", err)
		return
	}
	for _, effect := range result.Effects {
		p.output <- effect
	}
}

func (p *Plugin) processDeadlines(ctx context.Context) {
	for _, due := range p.store.Due() {
		result, err := p.store.Apply(ctx, due.Key, due.Input)
		if err != nil {
			p.log.Error("plugin deadline transition failed", "plugin", p.id, "error", err)
			continue
		}
		for _, effect := range result.Effects {
			p.output <- effect
		}
	}
	p.store.ExpireIdle()
}

func (p *Plugin) key(observation Observation) (string, error) {
	key := "global"
	switch p.cfg.KeyBy {
	case "account":
		key = observation.Account
	case "source":
		key = observation.SourceAddress
	}
	if key == "" || len(key) > 128 || strings.ContainsAny(key, "\n\r\x00") {
		return "", errors.New("missing or invalid key")
	}
	return observation.RuleID + "/" + key, nil
}

func (p *Plugin) deadline(instance fsm.Instance[state, instanceData]) (time.Time, input, bool) {
	data := instance.Data
	deadline := data.IdleDeadline
	kind := idleExpired
	if instance.State == counting && !data.WindowDeadline.IsZero() && (deadline.IsZero() || data.WindowDeadline.Before(deadline)) {
		deadline = data.WindowDeadline
		kind = windowExpired
	}
	if instance.State == cooldown && !data.CooldownDeadline.IsZero() && (deadline.IsZero() || data.CooldownDeadline.Before(deadline)) {
		deadline = data.CooldownDeadline
		kind = cooldownExpired
	}
	return deadline, kind, !deadline.IsZero()
}

func (p *Plugin) newMachine() (*fsm.Machine[state, input, instanceData, Detection], error) {
	guards := map[string]fsm.Guard[instanceData, input]{
		"below": func(_ context.Context, data instanceData, _ input) (bool, error) {
			return data.Count+1 < p.cfg.Threshold, nil
		},
		"reached": func(_ context.Context, data instanceData, _ input) (bool, error) {
			return data.Count+1 >= p.cfg.Threshold, nil
		},
		"one":  func(_ context.Context, _ instanceData, _ input) (bool, error) { return p.cfg.Threshold == 1, nil },
		"many": func(_ context.Context, _ instanceData, _ input) (bool, error) { return p.cfg.Threshold > 1, nil },
		"window_due": func(_ context.Context, data instanceData, _ input) (bool, error) {
			return !time.Now().Before(data.WindowDeadline), nil
		},
		"cooldown_due": func(_ context.Context, data instanceData, _ input) (bool, error) {
			return !time.Now().Before(data.CooldownDeadline), nil
		},
		"idle_due": func(_ context.Context, data instanceData, _ input) (bool, error) {
			return !time.Now().Before(data.IdleDeadline), nil
		},
	}
	actions := map[string]fsm.Action[instanceData, input, Detection]{
		"start": func(_ context.Context, data instanceData, _ input) (instanceData, []Detection, error) {
			data.Count = 1
			data.WindowDeadline = time.Now().Add(p.cfg.Window)
			data.IdleDeadline = time.Now().Add(p.cfg.Idle)
			return data, nil, nil
		},
		"increment": func(_ context.Context, data instanceData, _ input) (instanceData, []Detection, error) {
			data.Count++
			data.IdleDeadline = time.Now().Add(p.cfg.Idle)
			return data, nil, nil
		},
		"detect": func(_ context.Context, data instanceData, _ input) (instanceData, []Detection, error) {
			data.Count++
			data.CooldownDeadline = time.Now().Add(p.cfg.Cooldown)
			data.IdleDeadline = time.Now().Add(p.cfg.Idle)
			operation := "occurrence"
			if data.EventID == "" {
				var bytes [8]byte
				if _, err := rand.Read(bytes[:]); err != nil {
					return data, nil, err
				}
				data.EventID = "L" + hex.EncodeToString(bytes[:])
				operation = "create"
			}
			return data, []Detection{p.detection(data, operation)}, nil
		},
		"touch": func(_ context.Context, data instanceData, _ input) (instanceData, []Detection, error) {
			data.IdleDeadline = time.Now().Add(p.cfg.Idle)
			return data, nil, nil
		},
		"window_reset": func(_ context.Context, data instanceData, _ input) (instanceData, []Detection, error) {
			data.Count = 0
			data.WindowDeadline = time.Time{}
			return data, nil, nil
		},
		"cooldown_reset": func(_ context.Context, data instanceData, _ input) (instanceData, []Detection, error) {
			data.Count = 0
			data.CooldownDeadline = time.Time{}
			return data, nil, nil
		},
		"normalize": func(_ context.Context, data instanceData, _ input) (instanceData, []Detection, error) {
			var effects []Detection
			if data.EventID != "" {
				effects = []Detection{p.detection(data, "normalize")}
			}
			return instanceData{}, effects, nil
		},
	}
	rows := []fsm.Row[state, input]{
		{From: idle, On: observed, Guard: "one", To: cooldown, Action: "detect"},
		{From: idle, On: observed, Guard: "many", To: counting, Action: "start"},
		{From: counting, On: observed, Guard: "below", To: counting, Action: "increment"},
		{From: counting, On: observed, Guard: "reached", To: cooldown, Action: "detect"},
		{From: counting, On: windowExpired, Guard: "window_due", To: idle, Action: "window_reset"},
		{From: cooldown, On: observed, To: cooldown, Action: "touch"},
		{From: cooldown, On: cooldownExpired, Guard: "cooldown_due", To: idle, Action: "cooldown_reset"},
	}
	for _, from := range []state{idle, counting, cooldown} {
		rows = append(rows, fsm.Row[state, input]{From: from, On: idleExpired, Guard: "idle_due", To: idle, Action: "normalize"})
		rows = append(rows, fsm.Row[state, input]{From: from, On: succeeded, To: idle, Action: "normalize"})
	}
	return fsm.New(fsm.Definition[state, input, instanceData, Detection]{
		Initial: idle, States: []state{idle, counting, cooldown},
		Inputs: []input{observed, windowExpired, cooldownExpired, idleExpired, succeeded},
		Rows:   rows, Guards: guards, Actions: actions,
	})
}

func (p *Plugin) detection(data instanceData, operation string) Detection {
	return Detection{
		InstanceID: p.id, RuleID: data.Last.RuleID, Category: data.Last.Category,
		Severity: data.Last.Severity, Email: data.Last.Email, Lockdown: data.Last.Lockdown, Operation: operation,
		EventID: data.EventID, Account: data.Last.Account, SourceAddress: data.Last.SourceAddress, At: data.Last.At,
	}
}
