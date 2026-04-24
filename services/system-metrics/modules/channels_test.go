package modules

import "testing"

func TestNewChannelsModuleUsesConfiguredBufferSize(t *testing.T) {
	t.Parallel()

	module := NewChannelsModule(3)

	if cap(module.RawSnapshots()) != 3 {
		t.Fatalf("RawSnapshots() cap = %d, want 3", cap(module.RawSnapshots()))
	}

	if cap(module.SystemSnapshots()) != 3 {
		t.Fatalf("SystemSnapshots() cap = %d, want 3", cap(module.SystemSnapshots()))
	}
}
