// Package detection compiles per-rule alert detectors over the shared FSM.
package detection

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"lite-nas/shared/fsm"
)

// PluginName selects one of the supported automation detection machines.
type PluginName string

const (
	// Sequence matches ordered steps before one shared deadline.
	Sequence PluginName = "sequence"
	// CountWithinWindow matches a threshold inside a sliding-start window.
	CountWithinWindow PluginName = "count_within_window"
)

type (
	state string
	input string
)

const (
	idle     state = "idle"
	tracking state = "tracking"
	observed input = "observed"
	expired  input = "expired"
)

// Event is a validated, correlated alert supplied by a transport adapter.
type Event struct {
	ID      string
	Key     string
	Step    int
	At      time.Time
	Message string
}

// Match is one completed detection; external effects belong to another worker.
type Match struct {
	Key         string
	AlertIDs    []string
	LastMessage string
}

// Config defines one compiled detector and its bounded instance pool.
type Config struct {
	Name      PluginName
	Steps     int
	Threshold int
	Window    time.Duration
	MaxKeys   int
	MaxIdle   time.Duration
}

type data struct {
	Current  Event
	Count    int
	NextStep int
	Deadline time.Time
	IDs      []string
	Seen     []string
}

// Plugin owns one table and a bounded keyed store for one configured rule.
type Plugin struct {
	mutex sync.Mutex
	cfg   Config
	clock func() time.Time
	store *fsm.MemoryStore[state, input, data, Match]
}

// New compiles one detector. An invalid rule fails before subscriptions start.
func New(cfg Config, clock func() time.Time) (*Plugin, error) {
	if err := ValidateConfig(cfg, clock); err != nil {
		return nil, err
	}
	p := &Plugin{cfg: cfg, clock: clock}
	store, err := p.newStore()
	if err != nil {
		return nil, err
	}
	p.store = store
	return p, nil
}

// newStore creates a fresh bounded state pool for this already validated detector.
func (p *Plugin) newStore() (*fsm.MemoryStore[state, input, data, Match], error) {
	machine, err := p.machine()
	if err != nil {
		return nil, err
	}
	return fsm.NewMemoryStore(machine, p.cfg.MaxKeys, p.cfg.MaxIdle, func(string) data { return data{} }, p.clock,
		func(instance fsm.Instance[state, data]) (time.Time, input, bool) {
			return instance.Data.Deadline, expired, instance.State == tracking && !instance.Data.Deadline.IsZero()
		})
}

// Reset forgets incomplete detection state before a temporarily disabled rule is re-enabled.
func (p *Plugin) Reset() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	store, err := p.newStore()
	if err != nil {
		return err
	}
	p.store = store
	return nil
}

// ResetKey clears only one in-memory detection instance, leaving incidents
// and other correlation keys untouched.
func (p *Plugin) ResetKey(ctx context.Context, key string) (bool, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.store.ResetKey(ctx, key)
}

// ValidateConfig checks one FSM definition without creating a disabled rule instance.
func ValidateConfig(cfg Config, clock func() time.Time) error {
	if clock == nil || cfg.Window <= 0 || cfg.MaxKeys < 1 || cfg.MaxIdle <= cfg.Window ||
		(cfg.Name == Sequence && (cfg.Steps < 1 || cfg.Steps > 64)) ||
		(cfg.Name == CountWithinWindow && (cfg.Threshold < 1 || cfg.Threshold > 10000)) ||
		(cfg.Name != Sequence && cfg.Name != CountWithinWindow) {
		return errors.New("invalid automation detector configuration")
	}
	return nil
}

// Observe advances one correlation key. Step is the zero-based matched rule step.
// A repeated alert ID is ignored and cannot satisfy a count or sequence twice.
func (p *Plugin) Observe(ctx context.Context, event Event) ([]Match, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if event.ID == "" || event.Key == "" || len(event.Message) > 256 || event.At.IsZero() || event.At.After(p.clock().Add(5*time.Minute)) || event.Step < 0 ||
		(p.cfg.Name == Sequence && event.Step >= p.cfg.Steps) || (p.cfg.Name == CountWithinWindow && event.Step != 0) {
		return nil, errors.New("invalid automation observation")
	}
	result, err := p.store.ApplyWith(ctx, event.Key, observed, func(previous data) data {
		previous.Current = event
		return previous
	})
	return result.Effects, err
}

// Advance expires incomplete windows. It never produces a mitigation.
func (p *Plugin) Advance(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	for _, due := range p.store.Due() {
		if _, err := p.store.Apply(ctx, due.Key, due.Input); err != nil {
			return err
		}
	}
	p.store.ExpireIdle()
	return nil
}

func (p *Plugin) machine() (*fsm.Machine[state, input, data, Match], error) {
	return fsm.New(fsm.Definition[state, input, data, Match]{
		Initial: idle, States: []state{idle, tracking}, Inputs: []input{observed, expired},
		Rows: []fsm.Row[state, input]{
			{From: idle, On: observed, Guard: "eligible", To: tracking, Action: "observe"},
			{From: tracking, On: observed, Guard: "eligible", To: tracking, Action: "observe"},
			{From: tracking, On: expired, Guard: "due", To: idle, Action: "clear"},
		},
		Guards: map[string]fsm.Guard[data, input]{
			"eligible": func(_ context.Context, d data, _ input) (bool, error) {
				for _, id := range d.Seen {
					if id == d.Current.ID {
						return false, nil
					}
				}
				return true, nil
			},
			"due": func(_ context.Context, d data, _ input) (bool, error) { return !p.clock().Before(d.Deadline), nil },
		},
		Actions: map[string]fsm.Action[data, input, Match]{
			"clear": func(_ context.Context, _ data, _ input) (data, []Match, error) { return data{}, nil, nil },
			"observe": func(_ context.Context, d data, _ input) (data, []Match, error) {
				e := d.Current
				if d.Deadline.IsZero() || !e.At.Before(d.Deadline) || e.At.Before(d.Deadline.Add(-p.cfg.Window)) {
					d = data{Current: e, Deadline: e.At.Add(p.cfg.Window), Seen: d.Seen}
				}
				if p.cfg.Name == Sequence {
					if e.Step != d.NextStep {
						// A selector may match multiple steps. Keep the first
						// accepted step until its deadline so a later alert can
						// advance the sequence instead of restarting it.
						if d.NextStep == 0 {
							return data{Seen: d.Seen}, nil, nil
						}
						return d, nil, nil
					}
					d.NextStep++
				} else {
					d.Count++
				}
				d.IDs = append(append([]string(nil), d.IDs...), e.ID)
				d.Seen = append(append([]string(nil), d.Seen...), e.ID)
				if len(d.Seen) > 10000 {
					return data{}, nil, fmt.Errorf("detector observation bound exceeded")
				}
				complete := (p.cfg.Name == Sequence && d.NextStep == p.cfg.Steps) || (p.cfg.Name == CountWithinWindow && d.Count == p.cfg.Threshold)
				if complete {
					match := Match{Key: e.Key, AlertIDs: append([]string(nil), d.IDs...), LastMessage: e.Message}
					return data{Seen: d.Seen}, []Match{match}, nil
				}
				return d, nil, nil
			},
		},
	})
}
