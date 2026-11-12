package servicetoken

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	authcontract "lite-nas/shared/contracts/auth"
)

type authClientStub struct {
	active    atomic.Int32
	maxActive atomic.Int32
}

func (client *authClientStub) Request(_ context.Context, subject string, _ any, response any) error {
	active := client.active.Add(1)
	defer client.active.Add(-1)
	for current := client.maxActive.Load(); active > current && !client.maxActive.CompareAndSwap(current, active); current = client.maxActive.Load() {
	}
	time.Sleep(time.Millisecond)
	switch subject {
	case authcontract.ServiceTokenLoginRPCSubject:
		*response.(*authcontract.ServiceTokenLoginResponse) = authcontract.ServiceTokenLoginResponse{
			AccessToken: "access-token", RefreshToken: "refresh-token", ExpiresAt: time.Now().Add(15 * time.Minute),
		}
	case authcontract.ServiceTokenRefreshRPCSubject:
		*response.(*authcontract.ServiceTokenRefreshResponse) = authcontract.ServiceTokenRefreshResponse{
			AccessToken: "refreshed-token", RefreshToken: "rotated-refresh-token", ExpiresAt: time.Now().Add(15 * time.Minute),
		}
	}
	return nil
}

func TestConcurrentLoginsSerializeTokenStateUpdates(t *testing.T) {
	client := &authClientStub{}
	manager, err := NewManager(client, Options{Service: "log-analysis"})
	if err != nil {
		t.Fatal(err)
	}
	var calls sync.WaitGroup
	for range 8 {
		calls.Add(1)
		go func() {
			defer calls.Done()
			if loginErr := manager.Login(context.Background()); loginErr != nil {
				t.Errorf("Login() error = %v", loginErr)
			}
		}()
	}
	calls.Wait()
	if client.maxActive.Load() != 1 {
		t.Fatalf("concurrent auth requests = %d, want serialized requests", client.maxActive.Load())
	}
}
