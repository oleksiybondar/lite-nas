package messagingtest

import "context"

// RecordingClient is a minimal messaging.Client test double that records drain
// and close lifecycle calls.
type RecordingClient struct {
	DrainCalls int
	CloseCalls int
}

func (c *RecordingClient) Publish(context.Context, string, any) error {
	return nil
}

func (c *RecordingClient) Request(context.Context, string, any, any) error {
	return nil
}

func (c *RecordingClient) Drain() error {
	c.DrainCalls++
	return nil
}

func (c *RecordingClient) Close() {
	c.CloseCalls++
}
