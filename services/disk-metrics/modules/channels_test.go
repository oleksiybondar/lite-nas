package modules

import "testing"

func TestNewChannelsModuleAllocatesBufferedChannels(t *testing.T) {
	t.Parallel()

	channels := NewChannelsModule(2)
	if cap(channels.DiskSnapshots) != 2 {
		t.Fatalf("cap(DiskSnapshots) = %d, want 2", cap(channels.DiskSnapshots))
	}
	if cap(channels.PollErrors) != 2 {
		t.Fatalf("cap(PollErrors) = %d, want 2", cap(channels.PollErrors))
	}
}
