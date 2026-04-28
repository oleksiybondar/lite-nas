package modules

import (
	"testing"

	"lite-nas/shared/testutil/messagingtest"
)

func TestCoreInfraCloseDrainsAndClosesClient(t *testing.T) {
	t.Parallel()

	client := &messagingtest.RecordingClient{}
	core := CoreInfra{Client: client}

	core.Close()

	if client.DrainCalls != 1 {
		t.Fatalf("DrainCalls = %d, want 1", client.DrainCalls)
	}

	if client.CloseCalls != 1 {
		t.Fatalf("CloseCalls = %d, want 1", client.CloseCalls)
	}
}

func TestCoreInfraCloseDrainsAndClosesServer(t *testing.T) {
	t.Parallel()

	server := &messagingtest.RecordingServer{}
	core := CoreInfra{Server: server}

	core.Close()

	if server.DrainCalls != 1 {
		t.Fatalf("DrainCalls = %d, want 1", server.DrainCalls)
	}

	if server.CloseCalls != 1 {
		t.Fatalf("CloseCalls = %d, want 1", server.CloseCalls)
	}
}

func TestCoreInfraCloseRunsLogCleanup(t *testing.T) {
	t.Parallel()

	cleanupCalls := 0
	core := CoreInfra{
		logCleanup: func() {
			cleanupCalls++
		},
	}

	core.Close()

	if cleanupCalls != 1 {
		t.Fatalf("cleanupCalls = %d, want 1", cleanupCalls)
	}
}
