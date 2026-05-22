package modules

import (
	"testing"
	"time"

	sharedconfig "lite-nas/shared/config"
	"lite-nas/shared/testutil/messagingtest"
)

func TestBuildCoreClientAuthInfraWiresAuthDeps(t *testing.T) {
	t.Parallel()

	module, err := buildCoreClientAuthInfra(
		CoreInfra{Client: &messagingtest.RecordingClient{}},
		sharedconfig.AuthConfig{ServiceName: "resources-monitor"},
		24*time.Hour,
	)
	if err != nil {
		t.Fatalf("buildCoreClientAuthInfra() error = %v", err)
	}
	if module.AuthTokenManager == nil {
		t.Fatal("AuthTokenManager = nil, want initialized manager")
	}
	if module.AuthRefreshTicks == nil {
		t.Fatal("AuthRefreshTicks = nil, want initialized channel")
	}
}

func TestBuildCoreClientAuthInfraRejectsInvalidServiceName(t *testing.T) {
	t.Parallel()

	if _, err := buildCoreClientAuthInfra(
		CoreInfra{Client: &messagingtest.RecordingClient{}},
		sharedconfig.AuthConfig{},
		24*time.Hour,
	); err == nil {
		t.Fatal("buildCoreClientAuthInfra() error = nil, want invalid service name error")
	}
}

func TestBuildCoreClientAuthInfraRejectsInvalidRefreshInterval(t *testing.T) {
	t.Parallel()

	if _, err := buildCoreClientAuthInfra(
		CoreInfra{Client: &messagingtest.RecordingClient{}},
		sharedconfig.AuthConfig{ServiceName: "resources-monitor"},
		0,
	); err == nil {
		t.Fatal("buildCoreClientAuthInfra() error = nil, want invalid interval error")
	}
}
