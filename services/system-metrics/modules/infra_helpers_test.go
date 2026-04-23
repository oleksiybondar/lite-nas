package modules

import (
	"context"

	"lite-nas/shared/messaging"
)

type recordingMessagingClient struct {
	drained bool
	closed  bool
}

func (c *recordingMessagingClient) Publish(context.Context, string, any) error {
	return nil
}

func (c *recordingMessagingClient) Request(context.Context, string, any, any) error {
	return nil
}

func (c *recordingMessagingClient) Drain() error {
	c.drained = true
	return nil
}

func (c *recordingMessagingClient) Close() {
	c.closed = true
}

type recordingMessagingServer struct {
	drained bool
	closed  bool
}

func (s *recordingMessagingServer) Subscribe(string, messaging.MessageHandler) error {
	return nil
}

func (s *recordingMessagingServer) RegisterRPC(string, messaging.RPCHandler) error {
	return nil
}

func (s *recordingMessagingServer) Drain() error {
	s.drained = true
	return nil
}

func (s *recordingMessagingServer) Close() {
	s.closed = true
}
