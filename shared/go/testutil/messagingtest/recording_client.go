package messagingtest

import (
	"context"

	authcontract "lite-nas/shared/contracts/auth"
)

// RecordingClient is a minimal messaging.Client test double that records drain
// and close lifecycle calls.
type RecordingClient struct {
	DrainCalls   int
	CloseCalls   int
	AuthUsername string
}

func (c *RecordingClient) Publish(context.Context, string, any) error {
	return nil
}

func (c *RecordingClient) Request(_ context.Context, subject string, _ any, response any) error {
	if subject == authcontract.ValidateAccessTokenRPCSubject && c.AuthUsername != "" {
		*response.(*authcontract.ValidateAccessTokenResponse) = authcontract.ValidateAccessTokenResponse{Valid: true, Status: authcontract.StatusAuthenticated, Username: c.AuthUsername}
	}
	return nil
}

func (c *RecordingClient) Drain() error {
	c.DrainCalls++
	return nil
}

func (c *RecordingClient) Close() {
	c.CloseCalls++
}
