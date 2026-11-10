package loggingmanagerauth

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type tokenManagerStub struct {
	mutex        sync.Mutex
	accessToken  string
	expiresAt    time.Time
	refreshToken string
	loginToken   string
	refreshErr   error
	loginErr     error
	refreshCalls int
	loginCalls   int
}

func (manager *tokenManagerStub) Token() (string, time.Time, error) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	if manager.accessToken == "" {
		return "", time.Time{}, errors.New("token unavailable")
	}
	return manager.accessToken, manager.expiresAt, nil
}

func (manager *tokenManagerStub) Refresh(context.Context) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	manager.refreshCalls++
	if manager.refreshErr != nil {
		return manager.refreshErr
	}
	manager.accessToken = manager.refreshToken
	manager.expiresAt = time.Now().Add(15 * time.Minute)
	return nil
}

func (manager *tokenManagerStub) Login(context.Context) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	manager.loginCalls++
	if manager.loginErr != nil {
		return manager.loginErr
	}
	manager.accessToken = manager.loginToken
	manager.expiresAt = time.Now().Add(15 * time.Minute)
	return nil
}

func (manager *tokenManagerStub) calls() (int, int) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	return manager.refreshCalls, manager.loginCalls
}

func TestCurrentTokenRefreshesExpiringToken(t *testing.T) {
	manager := &tokenManagerStub{
		accessToken:  "old-token",
		expiresAt:    time.Now().Add(30 * time.Second),
		refreshToken: "renewed-token",
	}
	client := NewClient(nil, manager, nil)

	token, err := client.CurrentToken(context.Background())
	if err != nil || token != "renewed-token" {
		t.Fatalf("CurrentToken() = %q, %v; want renewed token", token, err)
	}
}

func TestCurrentTokenLogsInWhenRefreshFails(t *testing.T) {
	manager := &tokenManagerStub{
		accessToken: "expired-token",
		expiresAt:   time.Now().Add(-time.Second),
		refreshErr:  errors.New("refresh rejected"),
		loginToken:  "login-token",
	}
	client := NewClient(nil, manager, nil)

	token, err := client.CurrentToken(context.Background())
	if err != nil || token != "login-token" {
		t.Fatalf("CurrentToken() = %q, %v; want login token", token, err)
	}
}

func TestCurrentTokenDoesNotRefreshSafeToken(t *testing.T) {
	manager := &tokenManagerStub{accessToken: "valid-token", expiresAt: time.Now().Add(10 * time.Minute)}
	client := NewClient(nil, manager, nil)

	_, err := client.CurrentToken(context.Background())
	refreshCalls, loginCalls := manager.calls()
	if err != nil || refreshCalls != 0 || loginCalls != 0 {
		t.Fatalf("CurrentToken() calls = refresh %d, login %d, err %v; want no renewal", refreshCalls, loginCalls, err)
	}
}

func TestConcurrentCurrentTokenCallsShareOneRefresh(t *testing.T) {
	manager := &tokenManagerStub{
		accessToken:  "expired-token",
		expiresAt:    time.Now().Add(-time.Second),
		refreshToken: "renewed-token",
	}
	client := NewClient(nil, manager, nil)
	var calls sync.WaitGroup
	for range 8 {
		calls.Add(1)
		go func() {
			defer calls.Done()
			if _, err := client.CurrentToken(context.Background()); err != nil {
				t.Errorf("CurrentToken() error = %v", err)
			}
		}()
	}
	calls.Wait()
	refreshCalls, _ := manager.calls()
	if refreshCalls != 1 {
		t.Fatalf("refresh calls = %d, want one", refreshCalls)
	}
}
