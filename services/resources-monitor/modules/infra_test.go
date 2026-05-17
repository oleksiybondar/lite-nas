package modules

import "testing"

func TestNewInfraModuleReturnsErrorForMissingConfig(t *testing.T) {
	t.Parallel()

	_, err := NewInfraModule("/non-existent/resources-monitor.conf", "resources-monitor")
	if err == nil {
		t.Fatal("expected error for missing config")
	}
}
