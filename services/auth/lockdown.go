package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	authcontract "lite-nas/shared/contracts/auth"
	"lite-nas/shared/messaging"
)

const lockdownStatePath = "/var/lib/lite-nas-auth/lockdown.json"

// lockdownState is durable so a service restart cannot silently lift an
// active lockdown. Pending indicates a state-change event still needs sending.
type lockdownState struct {
	Lockdown  bool   `json:"lockdown"`
	Reason    string `json:"reason"`
	Initiator string `json:"initiator"`
	Pending   bool   `json:"pending"`
}

// lockdownStore serializes state changes and makes the applied policy visible
// to login and refresh handlers. Its file is owned by the auth service.
type lockdownStore struct {
	mu    sync.RWMutex
	path  string
	state lockdownState
}

// openLockdownStore loads durable policy. Missing state means the service has
// never been locked; malformed or unreadable state prevents startup.
func openLockdownStore(path string) (*lockdownStore, error) {
	store := &lockdownStore{path: path}
	data, err := os.ReadFile(path) // #nosec G304 -- fixed service path or test fixture
	if errors.Is(err, os.ErrNotExist) {
		return store, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read lockdown state: %w", err)
	}
	if err := json.Unmarshal(data, &store.state); err != nil {
		return nil, fmt.Errorf("decode lockdown state: %w", err)
	}
	if store.state.Reason == "" || store.state.Initiator == "" {
		return nil, errors.New("invalid lockdown state audit context")
	}
	return store, nil
}

// active reports whether new user authentication must be refused.
func (s *lockdownStore) active() bool {
	if s == nil {
		return false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Lockdown
}

// set applies a validated request durably before attempting publication.
// Repeated identical requests retry a pending event but create no new event.
func (s *lockdownStore) set(ctx context.Context, request authcontract.SetLockdownRequest, client messaging.Client) (authcontract.SetLockdownResponse, error) {
	if s == nil || client == nil {
		return authcontract.SetLockdownResponse{}, errors.New("lockdown dependencies unavailable")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state.Lockdown != request.Lockdown || s.state.Reason == "" {
		next := lockdownState{Lockdown: request.Lockdown, Reason: request.Reason, Initiator: request.Initiator, Pending: true}
		if err := s.write(next); err != nil {
			return authcontract.SetLockdownResponse{}, err
		}
		s.state = next
	}
	response := authcontract.SetLockdownResponse{Lockdown: s.state.Lockdown, Reason: s.state.Reason, Initiator: s.state.Initiator}
	if err := s.publishPending(ctx, client); err != nil {
		return response, err
	}
	return response, nil
}

// publishPending retries a pending transition after restart or an earlier
// delivery failure. Core NATS does not guarantee durable event delivery.
func (s *lockdownStore) publishPending(ctx context.Context, client messaging.Client) error {
	if !s.state.Pending {
		return nil
	}
	event := authcontract.LockdownChangedEvent{Lockdown: s.state.Lockdown, Reason: s.state.Reason, Initiator: s.state.Initiator}
	if err := client.Publish(ctx, authcontract.LockdownChangedEventSubject, event); err != nil {
		return fmt.Errorf("publish lockdown change: %w", err)
	}
	next := s.state
	next.Pending = false
	if err := s.write(next); err != nil {
		return err
	}
	s.state = next
	return nil
}

// replayPending sends an unfinished state-change event on service startup.
func (s *lockdownStore) replayPending(ctx context.Context, client messaging.Client) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.publishPending(ctx, client)
}

// write replaces the state atomically, including a file sync so a successful
// RPC cannot acknowledge policy that has not reached the filesystem.
func (s *lockdownStore) write(state lockdownState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	dir := filepath.Dir(s.path)
	file, err := os.CreateTemp(dir, ".lockdown-*")
	if err != nil {
		return fmt.Errorf("create lockdown state: %w", err)
	}
	defer os.Remove(file.Name())
	if err := file.Chmod(0o600); err != nil {
		file.Close()
		return err
	}
	if _, err := file.Write(data); err != nil {
		file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Rename(file.Name(), s.path); err != nil {
		return err
	}
	// A directory sync can fail after rename. Apply the stricter in-memory
	// policy immediately so that this failure cannot admit new logins.
	s.state = state
	directory, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}

// handleSetLockdownRPC validates the transport payload before changing auth
// policy. NATS ACLs restrict who may publish this RPC subject.
func handleSetLockdownRPC(ctx context.Context, runtimeDeps authRuntime, envelope messaging.Envelope) (authcontract.SetLockdownResponse, error) {
	var request authcontract.SetLockdownRequest
	if !decodeRPCRequest(envelope, &request, runtimeDeps.Validator) {
		return authcontract.SetLockdownResponse{}, errors.New("invalid lockdown request")
	}
	return runtimeDeps.Lockdown.set(ctx, request, runtimeDeps.Client)
}
