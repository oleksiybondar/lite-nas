package services

import (
	"context"

	authcontract "lite-nas/shared/contracts/auth"
	"lite-nas/shared/messaging"
)

// AuthSessionsService lists and revokes LiteNAS authentication sessions.
type AuthSessionsService interface {
	List(context.Context, string) ([]authcontract.SessionSummary, error)
	Revoke(context.Context, string, string) (bool, error)
}

type authSessionsService struct{ client messaging.Client }

// NewAuthSessionsService connects session administration to Auth Service RPCs.
func NewAuthSessionsService(client messaging.Client) AuthSessionsService {
	return authSessionsService{client: client}
}

// List retrieves active authentication sessions for an administrator.
func (s authSessionsService) List(ctx context.Context, token string) ([]authcontract.SessionSummary, error) {
	var response authcontract.ListSessionsResponse
	err := s.client.Request(ctx, authcontract.ListSessionsRPCSubject, authcontract.ListSessionsRequest{AccessToken: token}, &response)
	return response.Sessions, err
}

// Revoke invalidates one active authentication session.
func (s authSessionsService) Revoke(ctx context.Context, token, sessionID string) (bool, error) {
	var response authcontract.RevokeSessionResponse
	err := s.client.Request(ctx, authcontract.RevokeSessionRPCSubject, authcontract.RevokeSessionRequest{AccessToken: token, SessionID: sessionID}, &response)
	return response.Revoked, err
}
