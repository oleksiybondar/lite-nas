package modules

import (
	"testing"
	"time"

	serviceconfig "lite-nas/services/network-metrics/config"
)

func TestNewWorkersModuleBuildsTimerAndPollingWorker(t *testing.T) {
	t.Parallel()

	workers, err := NewWorkersModule(
		serviceconfig.MetricsConfig{PollInterval: time.Second},
		NewChannelsModule(1),
		SourcePaths{
			ProcNetDev:      "/proc/net/dev",
			SysClassNet:     "/sys/class/net",
			ProcNetSNMP:     "/proc/net/snmp",
			ProcNetNetstat:  "/proc/net/netstat",
			ProcNetTCP:      "/proc/net/tcp",
			ProcNetTCP6:     "/proc/net/tcp6",
			ProcNetUDP:      "/proc/net/udp",
			ProcNetUDP6:     "/proc/net/udp6",
			ProcNetSockstat: "/proc/net/sockstat",
			ProcSoftIRQs:    "/proc/softirqs",
		},
	)
	if err != nil {
		t.Fatalf("NewWorkersModule() error = %v", err)
	}
	_ = workers
}

func TestNewWorkersModuleRejectsInvalidPollInterval(t *testing.T) {
	t.Parallel()

	_, err := NewWorkersModule(serviceconfig.MetricsConfig{}, NewChannelsModule(1), SourcePaths{})
	if err == nil {
		t.Fatal("NewWorkersModule() error = nil, want invalid poll interval error")
	}
}
