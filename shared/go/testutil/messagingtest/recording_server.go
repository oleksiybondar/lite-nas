package messagingtest

import "lite-nas/shared/messaging"

// RecordingServer is a minimal messaging.Server test double that records drain
// and close lifecycle calls.
type RecordingServer struct {
	DrainCalls int
	CloseCalls int
}

func (s *RecordingServer) Subscribe(string, messaging.MessageHandler) error {
	return nil
}

func (s *RecordingServer) RegisterRPC(string, messaging.RPCHandler) error {
	return nil
}

func (s *RecordingServer) Drain() error {
	s.DrainCalls++
	return nil
}

func (s *RecordingServer) Close() {
	s.CloseCalls++
}
