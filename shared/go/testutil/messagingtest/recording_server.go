package messagingtest

import "lite-nas/shared/messaging"

// RecordingServer is a minimal messaging.Server test double that records drain
// and close lifecycle calls.
type RecordingServer struct {
	DrainCalls  int
	CloseCalls  int
	Handlers    map[string]messaging.RPCHandler
	RegisterErr error
	DrainErr    error
}

func (s *RecordingServer) Subscribe(string, messaging.MessageHandler) error {
	return nil
}

func (s *RecordingServer) RegisterRPC(subject string, handler messaging.RPCHandler) error {
	if s.RegisterErr != nil {
		return s.RegisterErr
	}
	if s.Handlers == nil {
		s.Handlers = make(map[string]messaging.RPCHandler)
	}
	s.Handlers[subject] = handler
	return nil
}

func (s *RecordingServer) UseSubscriptionMiddleware(...messaging.SubscriptionMiddleware) {}

func (s *RecordingServer) UseRPCMiddleware(...messaging.RPCMiddleware) {}

func (s *RecordingServer) Drain() error {
	s.DrainCalls++
	return s.DrainErr
}

func (s *RecordingServer) Close() {
	s.CloseCalls++
}
