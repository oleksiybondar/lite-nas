package servicetoken

import (
	"context"
	"errors"
	"time"

	authcontract "lite-nas/shared/contracts/auth"
)

var (
	// ErrTokenUnavailable reports that the manager does not currently hold a token.
	ErrTokenUnavailable = errors.New("service token unavailable")
	errInvalidConfig    = errors.New("service token manager invalid config")
)

// AuthClient defines auth-service request/reply behavior needed by the manager.
type AuthClient interface {
	// Request submits one RPC call and decodes its reply into response.
	Request(ctx context.Context, subject string, request any, response any) error
}

// Options configures service token lifecycle behavior.
type Options struct {
	// Service identifies the calling service.
	Service string
}

// State is the latest token state emitted to subscribers.
type State struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// Manager keeps one active service token and refreshes it before expiry.
type Manager struct {
	client  AuthClient
	options Options

	state State
}

// NewManager creates a service-to-service token manager.
func NewManager(client AuthClient, options Options) (*Manager, error) {
	if client == nil {
		return nil, errInvalidConfig
	}
	if options.Service == "" {
		return nil, errors.New("service name is required")
	}

	return &Manager{
		client:  client,
		options: options,
	}, nil
}

// Token returns the current access token and expiry.
func (m *Manager) Token() (string, time.Time, error) {
	if m.state.AccessToken == "" {
		return "", time.Time{}, ErrTokenUnavailable
	}

	return m.state.AccessToken, m.state.ExpiresAt, nil
}

// RefreshToken returns the current refresh token.
func (m *Manager) RefreshToken() (string, error) {
	if m.state.RefreshToken == "" {
		return "", ErrTokenUnavailable
	}

	return m.state.RefreshToken, nil
}

// Login requests an initial S2S token pair from auth-service.
func (m *Manager) Login(ctx context.Context) error {
	var response authcontract.ServiceTokenLoginResponse
	if err := m.client.Request(ctx, authcontract.ServiceTokenLoginRPCSubject, authcontract.ServiceTokenLoginRequest{
		Service: m.options.Service,
	}, &response); err != nil {
		return err
	}

	m.setState(State{
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
		ExpiresAt:    response.ExpiresAt,
	})
	return nil
}

// Refresh rotates the current S2S token pair through auth-service.
func (m *Manager) Refresh(ctx context.Context) error {
	if m.state.RefreshToken == "" {
		return ErrTokenUnavailable
	}

	var response authcontract.ServiceTokenRefreshResponse
	if err := m.client.Request(ctx, authcontract.ServiceTokenRefreshRPCSubject, authcontract.ServiceTokenRefreshRequest{
		Service:      m.options.Service,
		RefreshToken: m.state.RefreshToken,
	}, &response); err != nil {
		return err
	}

	m.setState(State{
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
		ExpiresAt:    response.ExpiresAt,
	})
	return nil
}

// setState updates current state and emits it to subscribers.
func (m *Manager) setState(state State) {
	m.state = state
}
