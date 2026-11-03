package workers

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"lite-nas/shared/metrics"
)

type interfaceKindTestCase struct {
	name  string
	build func(t *testing.T, path string, ifaceName string)
	want  string
}

func TestClassifyInterfaceKindMarkerVariants(t *testing.T) {
	t.Parallel()

	for _, testCase := range interfaceKindTestCases() {
		t.Run(testCase.name, func(t *testing.T) { runInterfaceKindCase(t, testCase.name, testCase.build, testCase.want) })
	}
}

func TestClassifyInterfaceKindPhysicalByBackingDevice(t *testing.T) {
	t.Parallel()

	ifacePath := filepath.Join(t.TempDir(), "iface")
	mustMkdirAll(t, filepath.Join(ifacePath, "device"))

	if got := classifyInterfaceKind(ifacePath, "eth0"); got != "physical" {
		t.Fatalf("classifyInterfaceKind() = %q, want physical", got)
	}
}

func TestClassifyInterfaceBusAndAdapterCollection(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	devicePath := filepath.Join(tempDir, "device")
	subsystemTarget := filepath.Join(tempDir, "bus", "pci")

	mustMkdirAll(t, devicePath)
	mustMkdirAll(t, subsystemTarget)
	mustWriteWorkerFile(t, filepath.Join(devicePath, "vendor"), "0x8086\n")
	mustWriteWorkerFile(t, filepath.Join(devicePath, "device"), "0x100e\n")
	mustWriteWorkerFile(t, filepath.Join(devicePath, "vendor_name"), "Intel\n")
	mustWriteWorkerFile(t, filepath.Join(devicePath, "device_name"), "I211\n")
	mustWriteWorkerFile(t, filepath.Join(devicePath, "description"), "Intel I211\n")
	mustSymlink(t, subsystemTarget, filepath.Join(devicePath, "subsystem"))

	bus := classifyInterfaceBus(devicePath)
	if bus == nil || *bus != "pci" {
		t.Fatalf("classifyInterfaceBus() = %v, want pci", bus)
	}

	adapter := collectInterfaceAdapter(devicePath)
	if adapter == nil {
		t.Fatal("collectInterfaceAdapter() = nil, want adapter")
	}
	if adapter.VendorID != "0x8086" || adapter.DeviceID != "0x100e" {
		t.Fatalf("adapter IDs = %q/%q, want 0x8086/0x100e", adapter.VendorID, adapter.DeviceID)
	}
}

func TestMergeProtocolCountersAndParseProtocolGroups(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "snmp")
	mustWriteWorkerFile(t, path, "Tcp: ActiveOpens RetransSegs\nTcp: 2 3\nUdp: InDatagrams NoPorts\nUdp: 5 1\n")

	protocols := emptyProtocolSnapshot()
	if err := mergeProtocolCounters(path, &protocols); err != nil {
		t.Fatalf("mergeProtocolCounters() error = %v", err)
	}

	if protocols.TCP["ActiveOpens"] != 2 || protocols.TCP["RetransSegs"] != 3 {
		t.Fatalf("TCP counters = %#v, want ActiveOpens=2 RetransSegs=3", protocols.TCP)
	}
	if protocols.UDP["InDatagrams"] != 5 || protocols.UDP["NoPorts"] != 1 {
		t.Fatalf("UDP counters = %#v, want InDatagrams=5 NoPorts=1", protocols.UDP)
	}
}

func TestParseProtocolCounterGroupsRejectsLengthMismatch(t *testing.T) {
	t.Parallel()

	_, err := parseProtocolCounterGroups("Tcp: A B\nTcp: 1\n")
	if err == nil {
		t.Fatal("parseProtocolCounterGroups() error = nil, want mismatch error")
	}
}

func TestParseProcNetDevParsesCounters(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "dev")
	mustWriteWorkerFile(t, path, "Inter-|   Receive |  Transmit\n face |bytes packets errs drop fifo frame compressed multicast|bytes packets errs drop fifo colls carrier compressed\neth0: 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16\n")

	stats, err := parseProcNetDev(path)
	if err != nil {
		t.Fatalf("parseProcNetDev() error = %v", err)
	}

	if stats["eth0"].RXBytes != 1 || stats["eth0"].TXCompressed != 16 {
		t.Fatalf("parseProcNetDev() stats = %#v, want parsed counters", stats["eth0"])
	}
}

func TestParseSockStatParsesSections(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "sockstat")
	mustWriteWorkerFile(t, path, "sockets: used 10\nTCP: inuse 4 orphan 1 tw 2 alloc 3 mem 5\nUDP: inuse 6 mem 7\nUDPLITE: inuse 8\nRAW: inuse 9\nFRAG: inuse 10 memory 11\n")

	sockStat, err := parseSockStat(path)
	if err != nil {
		t.Fatalf("parseSockStat() error = %v", err)
	}

	if sockStat.SocketsUsed != 10 || sockStat.TCPTimeWait != 2 || sockStat.FragMemory != 11 {
		t.Fatalf("parseSockStat() = %#v, want parsed values", sockStat)
	}
}

func TestParseSoftIRQsParsesTotalsAndPerCPU(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "softirqs")
	mustWriteWorkerFile(t, path, "                    CPU0       CPU1\nNET_RX: 1 2\nNET_TX: 3 4\n")

	softIRQs, err := parseSoftIRQs(path)
	if err != nil {
		t.Fatalf("parseSoftIRQs() error = %v", err)
	}

	if softIRQs.NETRXTotal != 3 || softIRQs.NETTXTotal != 7 {
		t.Fatalf("parseSoftIRQs() totals = %#v, want NETRX=3 NETTX=7", softIRQs)
	}
}

func TestSocketSummaryAggregatorBuildsTotalsAndStates(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "tcp")
	mustWriteWorkerFile(t, path, "  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt uid timeout inode\n   0: 0100007F:0016 0200007F:0035 01 00000000:00000000 00:00000000 00000000 0 0 0 1 0000000000000000 100 0 0 10 0\n   1: 0100007F:0016 00000000:0000 0A 00000000:00000000 00:00000000 00000000 0 0 0 1 0000000000000000 100 0 0 10 0\n")

	aggregator := newSocketSummaryAggregator()
	if err := aggregator.consumeTable(path, "tcp"); err != nil {
		t.Fatalf("consumeTable() error = %v", err)
	}

	snapshot := aggregator.build(metrics.NetworkSockStat{})
	assertSocketTotals(t, snapshot)
	assertTopSocketPorts(t, snapshot.TopLocalPorts)
	assertTopSocketIPs(t, snapshot.TopRemoteIPs)
}

func TestDecodeProcAddressHexAndParseProcAddress(t *testing.T) {
	t.Parallel()

	ipv4, err := decodeProcAddressHex("0100007F")
	if err != nil {
		t.Fatalf("decodeProcAddressHex() IPv4 error = %v", err)
	}
	if ipv4 != "127.0.0.1" {
		t.Fatalf("decodeProcAddressHex() IPv4 = %q, want 127.0.0.1", ipv4)
	}

	ip, port, err := parseProcAddress("0100007F:0016")
	if err != nil {
		t.Fatalf("parseProcAddress() error = %v", err)
	}
	if ip != "127.0.0.1" || port != 22 {
		t.Fatalf("parseProcAddress() = (%q, %d), want (127.0.0.1, 22)", ip, port)
	}
}

func TestDecodeSocketStateMapsKnownAndUnknownCodes(t *testing.T) {
	t.Parallel()

	if got := decodeSocketState("0A"); got != "LISTEN" {
		t.Fatalf("decodeSocketState() = %q, want LISTEN", got)
	}
	if got := decodeSocketState("ff"); got != "FF" {
		t.Fatalf("decodeSocketState() default = %q, want FF", got)
	}
}

func TestMiscWorkerHelpers(t *testing.T) {
	t.Parallel()

	if !shouldCountRemoteIP("127.0.0.2") {
		t.Fatal("shouldCountRemoteIP(127.0.0.2) = false, want true")
	}
	if shouldCountRemoteIP("0.0.0.0") {
		t.Fatal("shouldCountRemoteIP(0.0.0.0) = true, want false")
	}
	if got := sumUint64([]uint64{1, 2, 3}); got != 6 {
		t.Fatalf("sumUint64() = %d, want 6", got)
	}
	if got := parseNamedUintPairs([]string{"used", "2", "bad", "x"}); got["used"] != 2 {
		t.Fatalf("parseNamedUintPairs() = %#v, want used=2", got)
	}
}

func TestPollingWorkerWaitNextPollAndEmitError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ticks := make(chan struct{}, 1)
	errorsCh := make(chan error, 1)
	worker := NewPollingWorker("", "", "", "", "", "", "", "", "", "", ticks, nil, errorsCh)

	ticks <- struct{}{}
	if !worker.waitNextPoll(ctx) {
		t.Fatal("waitNextPoll() = false, want true when tick is available")
	}

	worker.emitError(ctx, nil)
	worker.emitError(ctx, errors.New("poll failed"))
	if len(errorsCh) != 1 {
		t.Fatalf("len(errors channel) = %d, want 1", len(errorsCh))
	}
}

func TestPollingWorkerPollAndSendAndPoll(t *testing.T) {
	t.Parallel()

	fixture := newPollingDataFixture(t)
	output := make(chan metrics.NetworkMetricsSnapshot, 1)
	errorsCh := make(chan error, 1)
	worker := fixture.worker(output, errorsCh)

	snapshot, err := worker.poll()
	if err != nil {
		t.Fatalf("poll() error = %v", err)
	}
	if len(snapshot.Interfaces) != 1 || snapshot.Protocols.TCP["ActiveOpens"] != 2 {
		t.Fatalf("poll() snapshot = %#v, want parsed interface and protocol data", snapshot)
	}

	worker.pollAndSend(context.Background())
	if len(output) != 1 {
		t.Fatalf("len(output) = %d, want 1", len(output))
	}
	if len(errorsCh) != 0 {
		t.Fatalf("len(errors) = %d, want 0", len(errorsCh))
	}
}

func TestPollingWorkerPollAndSendReportsErrors(t *testing.T) {
	t.Parallel()

	output := make(chan metrics.NetworkMetricsSnapshot, 1)
	errorsCh := make(chan error, 1)
	worker := NewPollingWorker("/missing", "/missing", "", "", "", "", "", "", "", "", nil, output, errorsCh)

	worker.pollAndSend(context.Background())
	if len(errorsCh) != 1 {
		t.Fatalf("len(errors) = %d, want 1", len(errorsCh))
	}
}

func mustWriteWorkerFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}

type pollingDataFixture struct {
	tempDir            string
	procNetDevPath     string
	sysClassNetPath    string
	procNetSNMPPath    string
	procNetNetstatPath string
	procNetTCPPath     string
	procNetTCP6Path    string
	procNetUDPPath     string
	procNetUDP6Path    string
	procSockstatPath   string
	procSoftIRQsPath   string
}

func newPollingDataFixture(t *testing.T) pollingDataFixture {
	t.Helper()

	tempDir := t.TempDir()
	fixture := pollingDataFixture{
		tempDir:            tempDir,
		procNetDevPath:     filepath.Join(tempDir, "proc-net-dev"),
		sysClassNetPath:    filepath.Join(tempDir, "sys-class-net"),
		procNetSNMPPath:    filepath.Join(tempDir, "proc-net-snmp"),
		procNetNetstatPath: filepath.Join(tempDir, "proc-net-netstat"),
		procNetTCPPath:     filepath.Join(tempDir, "proc-net-tcp"),
		procNetTCP6Path:    filepath.Join(tempDir, "proc-net-tcp6"),
		procNetUDPPath:     filepath.Join(tempDir, "proc-net-udp"),
		procNetUDP6Path:    filepath.Join(tempDir, "proc-net-udp6"),
		procSockstatPath:   filepath.Join(tempDir, "proc-net-sockstat"),
		procSoftIRQsPath:   filepath.Join(tempDir, "proc-softirqs"),
	}

	fixture.populate(t)
	return fixture
}

func (f pollingDataFixture) populate(t *testing.T) {
	t.Helper()

	deviceRoot := filepath.Join(f.tempDir, "devices", "virtual", "net", "eth0")
	statsDir := filepath.Join(deviceRoot, "statistics")
	mustMkdirAll(t, f.sysClassNetPath)
	mustMkdirAll(t, statsDir)
	mustWriteWorkerFile(t, f.procNetDevPath, procNetDevFixtureLine("eth0"))
	mustWriteWorkerFile(t, filepath.Join(deviceRoot, "ifindex"), "1\n")
	mustWriteWorkerFile(t, filepath.Join(deviceRoot, "operstate"), "up\n")
	mustWriteWorkerFile(t, filepath.Join(statsDir, "rx_bytes"), "10\n")
	mustWriteWorkerFile(t, filepath.Join(statsDir, "tx_bytes"), "20\n")
	mustSymlink(t, deviceRoot, filepath.Join(f.sysClassNetPath, "eth0"))
	mustWriteWorkerFile(t, f.procNetSNMPPath, "Tcp: ActiveOpens\nTcp: 2\n")
	mustWriteWorkerFile(t, f.procNetNetstatPath, "TcpExt: SyncookiesSent\nTcpExt: 1\n")
	mustWriteWorkerFile(t, f.procNetTCPPath, "sl local_address rem_address st\n0: 0100007F:0016 0200007F:0035 01\n")
	mustWriteWorkerFile(t, f.procNetTCP6Path, "sl local_address rem_address st\n")
	mustWriteWorkerFile(t, f.procNetUDPPath, "sl local_address rem_address st\n")
	mustWriteWorkerFile(t, f.procNetUDP6Path, "sl local_address rem_address st\n")
	mustWriteWorkerFile(t, f.procSockstatPath, "sockets: used 1\nTCP: inuse 1 orphan 0 tw 0 alloc 1 mem 1\n")
	mustWriteWorkerFile(t, f.procSoftIRQsPath, " CPU0 CPU1\nNET_RX: 1 2\nNET_TX: 3 4\n")
}

func (f pollingDataFixture) worker(output chan<- metrics.NetworkMetricsSnapshot, errorsCh chan<- error) PollingWorker {
	return NewPollingWorker(
		f.procNetDevPath,
		f.sysClassNetPath,
		f.procNetSNMPPath,
		f.procNetNetstatPath,
		f.procNetTCPPath,
		f.procNetTCP6Path,
		f.procNetUDPPath,
		f.procNetUDP6Path,
		f.procSockstatPath,
		f.procSoftIRQsPath,
		nil,
		output,
		errorsCh,
	)
}

func interfaceKindTestCases() []interfaceKindTestCase {
	return []interfaceKindTestCase{
		{name: "loopback by name", build: func(_ *testing.T, _ string, _ string) {}, want: "loopback"},
		{name: "vpn by wireguard marker", build: func(t *testing.T, path string, _ string) {
			mustWriteWorkerFile(t, filepath.Join(path, "wireguard"), "")
		}, want: "vpn"},
		{name: "tunnel by tun flags", build: func(t *testing.T, path string, _ string) {
			mustWriteWorkerFile(t, filepath.Join(path, "tun_flags"), "")
		}, want: "tunnel"},
		{name: "bridge by marker", build: func(t *testing.T, path string, _ string) {
			mustWriteWorkerFile(t, filepath.Join(path, "bridge"), "")
		}, want: "bridge"},
		{name: "bond by marker", build: func(t *testing.T, path string, _ string) {
			mustWriteWorkerFile(t, filepath.Join(path, "bonding"), "")
		}, want: "bond"},
		{name: "vlan by uevent", build: func(t *testing.T, path string, _ string) {
			mustWriteWorkerFile(t, filepath.Join(path, "uevent"), "DEVTYPE=vlan\n")
		}, want: "vlan"},
		{name: "virtual fallback", build: func(_ *testing.T, _ string, _ string) {}, want: "virtual"},
	}
}

func runInterfaceKindCase(
	t *testing.T,
	name string,
	build func(*testing.T, string, string),
	want string,
) {
	t.Helper()
	t.Parallel()

	ifacePath := filepath.Join(t.TempDir(), "iface")
	mustMkdirAll(t, ifacePath)
	ifaceName := "eth0"
	if name == "loopback by name" {
		ifaceName = "lo"
	}

	build(t, ifacePath, ifaceName)

	if got := classifyInterfaceKind(ifacePath, ifaceName); got != want {
		t.Fatalf("classifyInterfaceKind() = %q, want %q", got, want)
	}
}

func assertSocketTotals(t *testing.T, snapshot metrics.NetworkSocketSnapshot) {
	t.Helper()
	if snapshot.Total.TCP != 2 || snapshot.ByState["ESTABLISHED"] != 1 || snapshot.ByState["LISTEN"] != 1 {
		t.Fatalf("build() snapshot = %#v, want TCP totals and states", snapshot)
	}
}

func assertTopSocketPorts(t *testing.T, ports []metrics.NetworkSocketPortCount) {
	t.Helper()
	if len(ports) != 1 || ports[0].Port != 22 {
		t.Fatalf("TopLocalPorts = %#v, want port 22", ports)
	}
}

func assertTopSocketIPs(t *testing.T, ips []metrics.NetworkSocketIPCount) {
	t.Helper()
	if len(ips) != 1 || ips[0].IP != "127.0.0.2" {
		t.Fatalf("TopRemoteIPs = %#v, want 127.0.0.2", ips)
	}
}
