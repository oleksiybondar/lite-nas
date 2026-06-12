package modules

import "testing"

func TestNewChannelsModuleAllocatesBufferedChannels(t *testing.T) {
	t.Parallel()

	channels := NewChannelsModule(2)
	if cap(channels.NetworkSnapshots) != 2 {
		t.Fatalf("cap(NetworkSnapshots) = %d, want 2", cap(channels.NetworkSnapshots))
	}
	if cap(channels.PollErrors) != 2 {
		t.Fatalf("cap(PollErrors) = %d, want 2", cap(channels.PollErrors))
	}
}
