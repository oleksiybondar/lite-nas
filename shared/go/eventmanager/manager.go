package eventmanager

import (
	"errors"
	"strings"
	"sync"
)

// ErrEventAlreadyExists indicates that a cache entry already exists for the
// provided key fields.
var ErrEventAlreadyExists = errors.New("event already exists")

// Event stores cached event data associated with one monitor rule key.
type Event struct {
	Event      string
	Field      string
	Condition  string
	Qualifiers []string
	Payload    any
}

// Manager provides thread-safe in-memory event cache and event ID counter
// operations for monitor services.
type Manager struct {
	mu      sync.RWMutex
	events  map[string]Event
	counter uint64
}

// NewManager creates a new Manager with an empty event cache and counter set
// to the provided initial value.
func NewManager(initialCounter uint64) *Manager {
	return &Manager{
		events:  make(map[string]Event),
		counter: initialCounter,
	}
}

// BuildKey constructs the canonical cache key for one monitored rule.
func BuildKey(event, field, condition string, qualifiers ...string) string {
	key := event + ":" + field + ":" + condition
	if len(qualifiers) == 0 {
		return key
	}

	return key + ":" + strings.Join(qualifiers, ":")
}

// FindEvent returns a cached event for the provided rule key fields.
func (manager *Manager) FindEvent(event, field, condition string, qualifiers ...string) (Event, bool) {
	manager.mu.RLock()
	defer manager.mu.RUnlock()

	cachedEvent, exists := manager.events[BuildKey(event, field, condition, qualifiers...)]
	return cachedEvent, exists
}

// CreateEvent inserts a new cached event for the provided rule key fields.
//
// The function fails when an event with the same key already exists.
func (manager *Manager) CreateEvent(event, field, condition string, payload any, qualifiers ...string) error {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	key := BuildKey(event, field, condition, qualifiers...)
	if _, exists := manager.events[key]; exists {
		return ErrEventAlreadyExists
	}

	manager.events[key] = Event{
		Event:      event,
		Field:      field,
		Condition:  condition,
		Qualifiers: append([]string(nil), qualifiers...),
		Payload:    payload,
	}

	return nil
}

// DeleteEvent removes a cached event for the provided rule key fields.
func (manager *Manager) DeleteEvent(event, field, condition string, qualifiers ...string) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	delete(manager.events, BuildKey(event, field, condition, qualifiers...))
}

// GetCounter returns the current in-memory event ID counter value.
func (manager *Manager) GetCounter() uint64 {
	manager.mu.RLock()
	defer manager.mu.RUnlock()

	return manager.counter
}

// SetCounter replaces the current in-memory event ID counter value.
func (manager *Manager) SetCounter(value uint64) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	manager.counter = value
}

// NextCounter increments and returns the new in-memory event ID counter value.
func (manager *Manager) NextCounter() uint64 {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	manager.counter++
	return manager.counter
}
