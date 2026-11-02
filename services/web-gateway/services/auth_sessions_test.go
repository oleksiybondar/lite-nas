package services

import (
	"context"
	"testing"

	authcontract "lite-nas/shared/contracts/auth"
)

// TestAuthSessionsListRPC verifies the list request forwards the caller token.
func TestAuthSessionsListRPC(t *testing.T) {
	client := &alertsClientStub{requestFunc: func(_ context.Context, subject string, request, reply any) error {
		if subject != authcontract.ListSessionsRPCSubject || request.(authcontract.ListSessionsRequest).AccessToken != "token" {
			t.Fatal("unexpected list RPC")
		}
		reply.(*authcontract.ListSessionsResponse).Sessions = []authcontract.SessionSummary{{SessionID: "session_1"}}
		return nil
	}}
	items, err := NewAuthSessionsService(client).List(context.Background(), "token")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("sessions=%d", len(items))
	}
}

// TestAuthSessionsRevokeRPC verifies the revoke request forwards the selected session.
func TestAuthSessionsRevokeRPC(t *testing.T) {
	client := &alertsClientStub{requestFunc: func(_ context.Context, subject string, request, reply any) error {
		value := request.(authcontract.RevokeSessionRequest)
		if subject != authcontract.RevokeSessionRPCSubject || value.AccessToken != "token" || value.SessionID != "session_1" {
			t.Fatal("unexpected revoke RPC")
		}
		reply.(*authcontract.RevokeSessionResponse).Revoked = true
		return nil
	}}
	revoked, err := NewAuthSessionsService(client).Revoke(context.Background(), "token", "session_1")
	if err != nil {
		t.Fatal(err)
	}
	if !revoked {
		t.Fatal("session was not revoked")
	}
}
