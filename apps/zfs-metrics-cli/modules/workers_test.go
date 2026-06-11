package modules

import "testing"

func TestNewWorkersModuleReturnsProcessorWithDefaultConfigPath(t *testing.T) {
	t.Parallel()

	module := NewWorkersModule("/etc/lite-nas/zfs-metrics-cli.conf")
	invocation, err := module.ArgsProcessor.Process([]string{})
	if err != nil {
		t.Fatalf("ArgsProcessor.Process() error = %v", err)
	}

	if invocation.ConfigPath != "/etc/lite-nas/zfs-metrics-cli.conf" {
		t.Fatalf("ConfigPath = %q, want /etc/lite-nas/zfs-metrics-cli.conf", invocation.ConfigPath)
	}
}
