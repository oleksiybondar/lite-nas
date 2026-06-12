package workers

import (
	"os"
	"path/filepath"
	"testing"

	"lite-nas/shared/metrics"
)

// Requirements: network-metrics-svc/FR-001
func TestCollectInterfacesIncludesSymlinkEntriesFromSysClassNet(t *testing.T) {
	t.Parallel()

	fixture := newCollectInterfacesFixture(t)
	fixture.addVirtualInterface("eth-test")
	iface := mustCollectSingleInterface(t, fixture.worker())

	assertInterfaceName(t, iface, "eth-test")
	assertRXBytes(t, iface, 10)
	assertInterfaceKind(t, iface, "virtual")
	assertInterfaceBusNil(t, iface)
}

// Requirements: network-metrics-svc/FR-002
func TestCollectInterfacesClassifiesPhysicalPCIInterface(t *testing.T) {
	t.Parallel()

	fixture := newCollectInterfacesFixture(t)
	fixture.addPhysicalPCIInterface("enp0s3")
	iface := mustCollectSingleInterface(t, fixture.worker())

	assertInterfaceKind(t, iface, "physical")
	assertInterfaceBus(t, iface, "pci")
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(path, 0o750); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", filepath.Dir(path), err)
	}

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

func mustCollectSingleInterface(t *testing.T, worker PollingWorker) metrics.NetworkInterfaceSnapshot {
	t.Helper()

	interfaces, err := worker.collectInterfaces()
	if err != nil {
		t.Fatalf("collectInterfaces() error = %v", err)
	}
	if len(interfaces) != 1 {
		t.Fatalf("collectInterfaces() length = %d, want 1", len(interfaces))
	}

	return interfaces[0]
}

func assertInterfaceName(t *testing.T, iface metrics.NetworkInterfaceSnapshot, want string) {
	t.Helper()
	if iface.Name != want {
		t.Fatalf("collectInterfaces()[0].Name = %q, want %s", iface.Name, want)
	}
}

func assertRXBytes(t *testing.T, iface metrics.NetworkInterfaceSnapshot, want uint64) {
	t.Helper()
	if iface.Statistics.RXBytes != want {
		t.Fatalf("collectInterfaces()[0].Statistics.RXBytes = %d, want %d", iface.Statistics.RXBytes, want)
	}
}

func assertInterfaceKind(t *testing.T, iface metrics.NetworkInterfaceSnapshot, want string) {
	t.Helper()
	if iface.Kind != want {
		t.Fatalf("collectInterfaces()[0].Kind = %q, want %s", iface.Kind, want)
	}
}

func assertInterfaceBusNil(t *testing.T, iface metrics.NetworkInterfaceSnapshot) {
	t.Helper()
	if iface.Bus != nil {
		t.Fatalf("collectInterfaces()[0].Bus = %v, want nil", *iface.Bus)
	}
}

func assertInterfaceBus(t *testing.T, iface metrics.NetworkInterfaceSnapshot, want string) {
	t.Helper()
	if iface.Bus == nil || *iface.Bus != want {
		t.Fatalf("collectInterfaces()[0].Bus = %v, want %s", iface.Bus, want)
	}
}

type collectInterfacesFixture struct {
	t           *testing.T
	tempDir     string
	sysClassNet string
	procNetDev  string
}

func newCollectInterfacesFixture(t *testing.T) collectInterfacesFixture {
	t.Helper()

	tempDir := t.TempDir()
	fixture := collectInterfacesFixture{
		t:           t,
		tempDir:     tempDir,
		sysClassNet: filepath.Join(tempDir, "sys-class-net"),
		procNetDev:  filepath.Join(tempDir, "proc-net-dev"),
	}

	mustMkdirAll(t, fixture.sysClassNet)
	return fixture
}

func (f collectInterfacesFixture) worker() PollingWorker {
	return NewPollingWorker(
		f.procNetDev,
		f.sysClassNet,
		filepath.Join(f.tempDir, "snmp"),
		filepath.Join(f.tempDir, "netstat"),
		filepath.Join(f.tempDir, "tcp"),
		filepath.Join(f.tempDir, "tcp6"),
		filepath.Join(f.tempDir, "udp"),
		filepath.Join(f.tempDir, "udp6"),
		filepath.Join(f.tempDir, "sockstat"),
		filepath.Join(f.tempDir, "softirqs"),
		nil,
		nil,
		nil,
	)
}

func (f collectInterfacesFixture) addVirtualInterface(name string) {
	deviceRoot := filepath.Join(f.tempDir, "devices", "virtual", "net", name)
	statsDir := filepath.Join(deviceRoot, "statistics")

	mustMkdirAll(f.t, statsDir)
	mustWriteFile(f.t, f.procNetDev, procNetDevFixtureLine(name))
	mustWriteFile(f.t, filepath.Join(deviceRoot, "ifindex"), "7\n")
	mustWriteFile(f.t, filepath.Join(deviceRoot, "address"), "00:11:22:33:44:55\n")
	mustWriteFile(f.t, filepath.Join(deviceRoot, "mtu"), "1500\n")
	mustWriteFile(f.t, filepath.Join(deviceRoot, "operstate"), "up\n")
	mustWriteFile(f.t, filepath.Join(deviceRoot, "carrier"), "1\n")
	mustWriteFile(f.t, filepath.Join(deviceRoot, "tx_queue_len"), "1000\n")
	mustWriteFile(f.t, filepath.Join(statsDir, "rx_bytes"), "10\n")
	mustWriteFile(f.t, filepath.Join(statsDir, "tx_bytes"), "20\n")
	mustSymlink(f.t, deviceRoot, filepath.Join(f.sysClassNet, name))
}

func (f collectInterfacesFixture) addPhysicalPCIInterface(name string) {
	deviceRoot := filepath.Join(f.tempDir, "devices", "pci0000:00", "0000:00:03.0", "net", name)
	statsDir := filepath.Join(deviceRoot, "statistics")
	subsystemDir := filepath.Join(f.tempDir, "bus", "pci")

	mustMkdirAll(f.t, statsDir)
	mustMkdirAll(f.t, subsystemDir)
	mustWriteFile(f.t, f.procNetDev, procNetDevFixtureLine(name))
	mustWriteFile(f.t, filepath.Join(deviceRoot, "type"), "1\n")
	mustWriteFile(f.t, filepath.Join(deviceRoot, "ifindex"), "8\n")
	mustWriteFile(f.t, filepath.Join(statsDir, "rx_bytes"), "10\n")
	mustWriteFile(f.t, filepath.Join(statsDir, "tx_bytes"), "20\n")
	mustWriteFile(f.t, filepath.Join(deviceRoot, "device", "vendor"), "0x8086\n")
	mustWriteFile(f.t, filepath.Join(deviceRoot, "device", "device"), "0x100e\n")
	mustSymlink(f.t, subsystemDir, filepath.Join(deviceRoot, "device", "subsystem"))
	mustSymlink(f.t, deviceRoot, filepath.Join(f.sysClassNet, name))
}

func procNetDevFixtureLine(name string) string {
	return "Inter-|   Receive                                                |  Transmit\n face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed\n" +
		name + ": 10 1 0 0 0 0 0 0 20 2 0 0 0 0 0 0\n"
}

func mustSymlink(t *testing.T, target string, path string) {
	t.Helper()

	if err := os.Symlink(target, path); err != nil {
		t.Fatalf("Symlink(%q, %q) error = %v", target, path, err)
	}
}

// Requirements: network-metrics-svc/FR-003, network-metrics-svc/RR-001
func TestParseProtocolCounterGroupsSkipsNegativeValues(t *testing.T) {
	t.Parallel()

	groups, err := parseProtocolCounterGroups("Tcp: RtoAlgorithm RtoMin RtoMax MaxConn ActiveOpens\nTcp: 1 200 120000 -1 42\n")
	if err != nil {
		t.Fatalf("parseProtocolCounterGroups() error = %v", err)
	}

	tcpGroup := groups["Tcp"]
	if tcpGroup["ActiveOpens"] != 42 {
		t.Fatalf("Tcp.ActiveOpens = %d, want 42", tcpGroup["ActiveOpens"])
	}
	if _, ok := tcpGroup["MaxConn"]; ok {
		t.Fatal("Tcp.MaxConn should be omitted for negative values")
	}
}
