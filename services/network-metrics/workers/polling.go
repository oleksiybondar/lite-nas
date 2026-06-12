package workers

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"net/netip"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"lite-nas/shared/metrics"
)

const topSocketRankLimit = 10

// PollingWorker periodically reads host network metric sources and emits
// network snapshots into an output channel.
type PollingWorker struct {
	procNetDevPath      string
	sysClassNetPath     string
	procNetSNMPPath     string
	procNetNetstatPath  string
	procNetTCPPath      string
	procNetTCP6Path     string
	procNetUDPPath      string
	procNetUDP6Path     string
	procNetSockstatPath string
	procSoftIRQsPath    string
	ticks               <-chan struct{}
	output              chan<- metrics.NetworkMetricsSnapshot
	errors              chan<- error
}

// NewPollingWorker creates a PollingWorker with the dependencies required for
// periodic network snapshot collection.
func NewPollingWorker(
	procNetDevPath string,
	sysClassNetPath string,
	procNetSNMPPath string,
	procNetNetstatPath string,
	procNetTCPPath string,
	procNetTCP6Path string,
	procNetUDPPath string,
	procNetUDP6Path string,
	procNetSockstatPath string,
	procSoftIRQsPath string,
	ticks <-chan struct{},
	output chan<- metrics.NetworkMetricsSnapshot,
	errors chan<- error,
) PollingWorker {
	return PollingWorker{
		procNetDevPath:      procNetDevPath,
		sysClassNetPath:     sysClassNetPath,
		procNetSNMPPath:     procNetSNMPPath,
		procNetNetstatPath:  procNetNetstatPath,
		procNetTCPPath:      procNetTCPPath,
		procNetTCP6Path:     procNetTCP6Path,
		procNetUDPPath:      procNetUDPPath,
		procNetUDP6Path:     procNetUDP6Path,
		procNetSockstatPath: procNetSockstatPath,
		procSoftIRQsPath:    procSoftIRQsPath,
		ticks:               ticks,
		output:              output,
		errors:              errors,
	}
}

// Start launches the polling worker in a separate goroutine.
func (w PollingWorker) Start(ctx context.Context) {
	go w.run(ctx)
}

// run executes the polling loop until the provided context is canceled.
func (w PollingWorker) run(ctx context.Context) {
	for {
		if !w.waitNextPoll(ctx) {
			return
		}

		w.pollAndSend(ctx)
	}
}

// waitNextPoll blocks until the next poll tick arrives or the context is
// canceled.
func (w PollingWorker) waitNextPoll(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	case _, ok := <-w.ticks:
		return ok
	}
}

// pollAndSend performs one polling cycle and forwards the resulting snapshot.
func (w PollingWorker) pollAndSend(ctx context.Context) {
	snapshot, err := w.poll()
	if err != nil {
		w.emitError(ctx, err)
		return
	}

	select {
	case <-ctx.Done():
		return
	case w.output <- snapshot:
	}
}

// emitError reports one polling error to the runtime when possible.
func (w PollingWorker) emitError(ctx context.Context, err error) {
	if err == nil {
		return
	}

	select {
	case <-ctx.Done():
		return
	case w.errors <- err:
	default:
	}
}

// poll reads all network sources required for one snapshot cycle.
func (w PollingWorker) poll() (metrics.NetworkMetricsSnapshot, error) {
	interfaces, err := w.collectInterfaces()
	if err != nil {
		return metrics.NetworkMetricsSnapshot{}, err
	}

	protocols, err := w.collectProtocols()
	if err != nil {
		return metrics.NetworkMetricsSnapshot{}, err
	}

	sockets, err := w.collectSockets()
	if err != nil {
		return metrics.NetworkMetricsSnapshot{}, err
	}

	kernelPressure, err := w.collectKernelPressure()
	if err != nil {
		return metrics.NetworkMetricsSnapshot{}, err
	}

	return metrics.NetworkMetricsSnapshot{
		Timestamp:      time.Now(),
		Interfaces:     interfaces,
		Protocols:      protocols,
		Sockets:        sockets,
		KernelPressure: kernelPressure,
	}, nil
}

// collectInterfaces builds the interface section for one snapshot.
func (w PollingWorker) collectInterfaces() ([]metrics.NetworkInterfaceSnapshot, error) {
	devStatsByName, err := parseProcNetDev(w.procNetDevPath)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(w.sysClassNetPath)
	if err != nil {
		return nil, fmt.Errorf("read sysfs interfaces: %w", err)
	}

	interfaces := make([]metrics.NetworkInterfaceSnapshot, 0, len(entries))
	for _, entry := range entries {
		entryType := entry.Type()
		if !entry.IsDir() && entryType&os.ModeSymlink == 0 {
			continue
		}

		ifaceSnapshot := collectInterfaceSnapshot(
			filepath.Join(w.sysClassNetPath, entry.Name()),
			entry.Name(),
			devStatsByName[entry.Name()],
		)
		interfaces = append(interfaces, ifaceSnapshot)
	}

	sort.Slice(interfaces, func(i int, j int) bool {
		return interfaces[i].Name < interfaces[j].Name
	})

	return interfaces, nil
}

// collectProtocols builds the protocol counter section for one snapshot.
func (w PollingWorker) collectProtocols() (metrics.NetworkProtocolSnapshot, error) {
	protocols := emptyProtocolSnapshot()

	if err := mergeProtocolCounters(w.procNetSNMPPath, &protocols); err != nil {
		return metrics.NetworkProtocolSnapshot{}, err
	}

	if err := mergeProtocolCounters(w.procNetNetstatPath, &protocols); err != nil {
		return metrics.NetworkProtocolSnapshot{}, err
	}

	return protocols, nil
}

// collectSockets builds the socket summary section for one snapshot.
func (w PollingWorker) collectSockets() (metrics.NetworkSocketSnapshot, error) {
	aggregator := newSocketSummaryAggregator()

	if err := aggregator.consumeTable(w.procNetTCPPath, "tcp"); err != nil {
		return metrics.NetworkSocketSnapshot{}, err
	}
	if err := aggregator.consumeTable(w.procNetTCP6Path, "tcp6"); err != nil {
		return metrics.NetworkSocketSnapshot{}, err
	}
	if err := aggregator.consumeTable(w.procNetUDPPath, "udp"); err != nil {
		return metrics.NetworkSocketSnapshot{}, err
	}
	if err := aggregator.consumeTable(w.procNetUDP6Path, "udp6"); err != nil {
		return metrics.NetworkSocketSnapshot{}, err
	}

	sockStat, err := parseSockStat(w.procNetSockstatPath)
	if err != nil {
		return metrics.NetworkSocketSnapshot{}, err
	}

	return aggregator.build(sockStat), nil
}

// collectKernelPressure builds the kernel pressure section for one snapshot.
func (w PollingWorker) collectKernelPressure() (metrics.NetworkKernelPressureSnapshot, error) {
	softIRQs, err := parseSoftIRQs(w.procSoftIRQsPath)
	if err != nil {
		return metrics.NetworkKernelPressureSnapshot{}, err
	}

	return metrics.NetworkKernelPressureSnapshot{
		SoftIRQs: softIRQs,
	}, nil
}

// collectInterfaceSnapshot builds one interface snapshot from sysfs and procfs
// data.
func collectInterfaceSnapshot(
	ifacePath string,
	ifaceName string,
	devStats procNetDevStats,
) metrics.NetworkInterfaceSnapshot {
	kind := classifyInterfaceKind(ifacePath, ifaceName)
	bus := classifyInterfaceBus(filepath.Join(ifacePath, "device"))

	return metrics.NetworkInterfaceSnapshot{
		Name:       ifaceName,
		IfIndex:    readOptionalUint(filepath.Join(ifacePath, "ifindex")),
		Address:    readOptionalString(filepath.Join(ifacePath, "address")),
		MTU:        readOptionalUint(filepath.Join(ifacePath, "mtu")),
		OperState:  readOptionalString(filepath.Join(ifacePath, "operstate")),
		CarrierUp:  readOptionalBool(filepath.Join(ifacePath, "carrier")),
		SpeedMbps:  readOptionalUint(filepath.Join(ifacePath, "speed")),
		Duplex:     readOptionalString(filepath.Join(ifacePath, "duplex")),
		TxQueueLen: readOptionalUint(filepath.Join(ifacePath, "tx_queue_len")),
		Kind:       kind,
		Bus:        bus,
		Adapter:    collectInterfaceAdapter(filepath.Join(ifacePath, "device")),
		Statistics: collectInterfaceStatistics(filepath.Join(ifacePath, "statistics"), devStats),
	}
}

// classifyInterfaceKind derives a stable service-level interface category
// from factual sysfs markers and interface naming.
func classifyInterfaceKind(ifacePath string, ifaceName string) string {
	for _, rule := range interfaceKindRules(ifacePath, ifaceName) {
		if rule.matches() {
			return rule.kind
		}
	}

	return "virtual"
}

// classifyInterfaceBus derives the device attachment bus from sysfs when the
// interface has a resolvable backing device.
func classifyInterfaceBus(devicePath string) *string {
	subsystemPath, err := filepath.EvalSymlinks(filepath.Join(devicePath, "subsystem"))
	if err != nil {
		return nil
	}

	bus := filepath.Base(subsystemPath)
	if bus == "" || bus == "." || bus == string(filepath.Separator) {
		return nil
	}

	return &bus
}

func readInterfaceType(path string) uint64 {
	value, ok := readOptionalUintValue(path)
	if !ok {
		return 0
	}

	return value
}

func hasBackingDevice(ifacePath string) bool {
	_, err := os.Stat(filepath.Join(ifacePath, "device"))
	return err == nil
}

func isVirtualInterface(ifacePath string) bool {
	resolvedPath, err := filepath.EvalSymlinks(ifacePath)
	if err != nil {
		return false
	}

	return strings.Contains(resolvedPath, "/virtual/")
}

func isVLANInterface(ifacePath string) bool {
	uevent := readOptionalString(filepath.Join(ifacePath, "uevent"))
	return strings.Contains(uevent, "DEVTYPE=vlan")
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

const loopbackInterfaceType uint64 = 772

type interfaceKindRule struct {
	kind    string
	matches func() bool
}

func interfaceKindRules(ifacePath string, ifaceName string) []interfaceKindRule {
	return []interfaceKindRule{
		{kind: "loopback", matches: func() bool {
			return ifaceName == "lo" || readInterfaceType(filepath.Join(ifacePath, "type")) == loopbackInterfaceType
		}},
		{kind: "vpn", matches: func() bool { return pathExists(filepath.Join(ifacePath, "wireguard")) }},
		{kind: "tunnel", matches: func() bool { return pathExists(filepath.Join(ifacePath, "tun_flags")) }},
		{kind: "bridge", matches: func() bool { return pathExists(filepath.Join(ifacePath, "bridge")) }},
		{kind: "bond", matches: func() bool { return pathExists(filepath.Join(ifacePath, "bonding")) }},
		{kind: "vlan", matches: func() bool { return isVLANInterface(ifacePath) }},
		{kind: "physical", matches: func() bool { return hasBackingDevice(ifacePath) && !isVirtualInterface(ifacePath) }},
	}
}

// collectInterfaceAdapter reads stable adapter identity metadata when the
// interface exposes a backing device in sysfs.
func collectInterfaceAdapter(devicePath string) *metrics.NetworkInterfaceAdapter {
	vendorID := readOptionalString(filepath.Join(devicePath, "vendor"))
	deviceID := readOptionalString(filepath.Join(devicePath, "device"))
	vendorName := readOptionalNamedString(filepath.Join(devicePath, "vendor_name"))
	deviceName := readOptionalNamedString(filepath.Join(devicePath, "device_name"))
	description := readOptionalNamedString(filepath.Join(devicePath, "description"))

	if vendorID == "" && deviceID == "" && vendorName == nil && deviceName == nil && description == nil {
		return nil
	}

	return &metrics.NetworkInterfaceAdapter{
		VendorID:    vendorID,
		DeviceID:    deviceID,
		VendorName:  vendorName,
		DeviceName:  deviceName,
		Description: description,
	}
}

// collectInterfaceStatistics reads sysfs statistics and falls back to procfs
// counters for fields that are not exposed individually.
func collectInterfaceStatistics(
	statisticsPath string,
	devStats procNetDevStats,
) metrics.NetworkInterfaceStatistics {
	return metrics.NetworkInterfaceStatistics{
		RXBytes:           readStatUint(statisticsPath, "rx_bytes", devStats.RXBytes),
		RXPackets:         readStatUint(statisticsPath, "rx_packets", devStats.RXPackets),
		RXErrors:          readStatUint(statisticsPath, "rx_errors", devStats.RXErrors),
		RXDropped:         readStatUint(statisticsPath, "rx_dropped", devStats.RXDropped),
		RXFIFOErrors:      readStatUint(statisticsPath, "rx_fifo_errors", devStats.RXFIFOErrors),
		RXFrameErrors:     readStatUint(statisticsPath, "rx_frame_errors", devStats.RXFrameErrors),
		RXCompressed:      readStatUint(statisticsPath, "rx_compressed", devStats.RXCompressed),
		RXMulticast:       readStatUint(statisticsPath, "multicast", devStats.RXMulticast),
		RXCRCErrors:       readStatUint(statisticsPath, "rx_crc_errors", 0),
		RXLengthErrors:    readStatUint(statisticsPath, "rx_length_errors", 0),
		RXMissedErrors:    readStatUint(statisticsPath, "rx_missed_errors", 0),
		RXOverErrors:      readStatUint(statisticsPath, "rx_over_errors", 0),
		RXNoHandler:       readStatUint(statisticsPath, "rx_nohandler", 0),
		TXBytes:           readStatUint(statisticsPath, "tx_bytes", devStats.TXBytes),
		TXPackets:         readStatUint(statisticsPath, "tx_packets", devStats.TXPackets),
		TXErrors:          readStatUint(statisticsPath, "tx_errors", devStats.TXErrors),
		TXDropped:         readStatUint(statisticsPath, "tx_dropped", devStats.TXDropped),
		TXFIFOErrors:      readStatUint(statisticsPath, "tx_fifo_errors", devStats.TXFIFOErrors),
		TXCollisions:      readStatUint(statisticsPath, "collisions", devStats.TXCollisions),
		TXCarrierErrors:   readStatUint(statisticsPath, "tx_carrier_errors", devStats.TXCarrierErrors),
		TXCompressed:      readStatUint(statisticsPath, "tx_compressed", devStats.TXCompressed),
		TXAbortedErrors:   readStatUint(statisticsPath, "tx_aborted_errors", 0),
		TXHeartbeatErrors: readStatUint(statisticsPath, "tx_heartbeat_errors", 0),
		TXWindowErrors:    readStatUint(statisticsPath, "tx_window_errors", 0),
	}
}

// emptyProtocolSnapshot creates a protocol snapshot with initialized counter
// maps for stable JSON object output.
func emptyProtocolSnapshot() metrics.NetworkProtocolSnapshot {
	return metrics.NetworkProtocolSnapshot{
		IP:      make(metrics.NetworkCounterGroup),
		ICMP:    make(metrics.NetworkCounterGroup),
		TCP:     make(metrics.NetworkCounterGroup),
		UDP:     make(metrics.NetworkCounterGroup),
		UDPLite: make(metrics.NetworkCounterGroup),
		IPExt:   make(metrics.NetworkCounterGroup),
		TCPExt:  make(metrics.NetworkCounterGroup),
	}
}

// mergeProtocolCounters merges parsed counter groups from one procfs source
// into the protocol snapshot.
func mergeProtocolCounters(path string, protocols *metrics.NetworkProtocolSnapshot) error {
	groups, err := readProtocolCounterGroups(path)
	if err != nil {
		return err
	}

	for groupName, values := range groups {
		mergeProtocolGroup(protocols, groupName, values)
	}

	return nil
}

func readProtocolCounterGroups(path string) (map[string]map[string]uint64, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- path is configured by the application
	if err != nil {
		return nil, fmt.Errorf("read protocol counters %s: %w", path, err)
	}

	groups, err := parseProtocolCounterGroups(string(data))
	if err != nil {
		return nil, fmt.Errorf("parse protocol counters %s: %w", path, err)
	}

	return groups, nil
}

func mergeProtocolGroup(
	protocols *metrics.NetworkProtocolSnapshot,
	groupName string,
	values map[string]uint64,
) {
	target := selectProtocolGroup(protocols, groupName)
	if target == nil {
		return
	}

	for counterName, counterValue := range values {
		(*target)[counterName] = counterValue
	}
}

// selectProtocolGroup resolves one kernel group name to its target snapshot map.
func selectProtocolGroup(
	protocols *metrics.NetworkProtocolSnapshot,
	groupName string,
) *metrics.NetworkCounterGroup {
	groups := map[string]*metrics.NetworkCounterGroup{
		"Ip":      &protocols.IP,
		"Icmp":    &protocols.ICMP,
		"Tcp":     &protocols.TCP,
		"Udp":     &protocols.UDP,
		"UdpLite": &protocols.UDPLite,
		"IpExt":   &protocols.IPExt,
		"TcpExt":  &protocols.TCPExt,
	}

	return groups[groupName]
}

// parseProtocolCounterGroups parses the paired-header procfs format used by
// /proc/net/snmp and /proc/net/netstat.
func parseProtocolCounterGroups(content string) (map[string]map[string]uint64, error) {
	lines := strings.Split(content, "\n")
	result := make(map[string]map[string]uint64)

	for lineIndex := 0; lineIndex < len(lines)-1; lineIndex++ {
		groupName, groupValues, ok, err := parseProtocolCounterGroupPair(lines[lineIndex], lines[lineIndex+1])
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}

		result[groupName] = groupValues

		lineIndex++
	}

	return result, nil
}

func parseProtocolCounterGroupPair(
	headerLine string,
	valueLine string,
) (string, map[string]uint64, bool, error) {
	headerGroup, headerFields, valueGroup, valueFields, ok := parseProtocolCounterPairLines(headerLine, valueLine)
	if !ok {
		return "", nil, false, nil
	}
	if len(headerFields) != len(valueFields) {
		return "", nil, false, fmt.Errorf("counter/value length mismatch for %s", headerGroup)
	}

	groupValues, err := collectProtocolCounterValues(headerGroup, headerFields, valueFields)
	if err != nil {
		return "", nil, false, err
	}

	return valueGroup, groupValues, true, nil
}

func parseProtocolCounterPairLines(
	headerLine string,
	valueLine string,
) (string, []string, string, []string, bool) {
	headerLine = strings.TrimSpace(headerLine)
	valueLine = strings.TrimSpace(valueLine)
	if !isProcHeaderPairLine(headerLine) || !isProcHeaderPairLine(valueLine) {
		return "", nil, "", nil, false
	}

	headerGroup, headerFields, ok := mustSplitProcHeaderLine(headerLine)
	if !ok {
		return "", nil, "", nil, false
	}

	valueGroup, valueFields, ok := mustSplitProcHeaderLine(valueLine)
	if !ok {
		return "", nil, "", nil, false
	}
	if headerGroup != valueGroup {
		return "", nil, "", nil, false
	}

	return headerGroup, headerFields, valueGroup, valueFields, true
}

func isProcHeaderPairLine(line string) bool {
	return line != "" && strings.Contains(line, ":")
}

func mustSplitProcHeaderLine(line string) (string, []string, bool) {
	groupName, fields, ok := splitProcHeaderLine(line)
	if !ok {
		return "", nil, false
	}

	return groupName, fields, true
}

func collectProtocolCounterValues(
	groupName string,
	headerFields []string,
	valueFields []string,
) (map[string]uint64, error) {
	groupValues := make(map[string]uint64, len(headerFields))
	for fieldIndex, fieldName := range headerFields {
		value, ok, err := parseProtocolCounterValue(valueFields[fieldIndex])
		if err != nil {
			return nil, fmt.Errorf("parse %s.%s: %w", groupName, fieldName, err)
		}
		if ok {
			groupValues[fieldName] = value
		}
	}

	return groupValues, nil
}

// parseProtocolCounterValue parses one procfs protocol counter value. Negative
// values are skipped because the shared counter map uses uint64 values only.
func parseProtocolCounterValue(raw string) (uint64, bool, error) {
	value, err := strconv.ParseUint(raw, 10, 64)
	if err == nil {
		return value, true, nil
	}

	signedValue, signedErr := strconv.ParseInt(raw, 10, 64)
	if signedErr == nil && signedValue < 0 {
		return 0, false, nil
	}

	return 0, false, err
}

// splitProcHeaderLine splits one procfs header/value line into group and
// fields.
func splitProcHeaderLine(line string) (string, []string, bool) {
	groupName, fieldsText, ok := strings.Cut(line, ":")
	if !ok {
		return "", nil, false
	}

	return strings.TrimSpace(groupName), strings.Fields(strings.TrimSpace(fieldsText)), true
}

// parseProcNetDev parses the /proc/net/dev table into a per-interface fallback
// stats map.
func parseProcNetDev(path string) (map[string]procNetDevStats, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- path is configured by the application
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	lines := strings.Split(string(data), "\n")
	statsByName := make(map[string]procNetDevStats)
	for _, line := range lines[2:] {
		if strings.TrimSpace(line) == "" {
			continue
		}

		ifaceName, stats, err := parseProcNetDevLine(line)
		if err != nil {
			return nil, err
		}

		statsByName[ifaceName] = stats
	}

	return statsByName, nil
}

func parseProcNetDevLine(line string) (string, procNetDevStats, error) {
	ifaceName, valuesText, ok := strings.Cut(strings.TrimSpace(line), ":")
	if !ok {
		return "", procNetDevStats{}, fmt.Errorf("invalid /proc/net/dev line: %q", line)
	}

	fields := strings.Fields(valuesText)
	name := strings.TrimSpace(ifaceName)
	if len(fields) < 16 {
		return "", procNetDevStats{}, fmt.Errorf("unexpected /proc/net/dev field count for %s", name)
	}

	values, err := parseUintFields(fields[:16])
	if err != nil {
		return "", procNetDevStats{}, fmt.Errorf("parse /proc/net/dev counters for %s: %w", name, err)
	}

	return name, procNetDevStats{
		RXBytes:         values[0],
		RXPackets:       values[1],
		RXErrors:        values[2],
		RXDropped:       values[3],
		RXFIFOErrors:    values[4],
		RXFrameErrors:   values[5],
		RXCompressed:    values[6],
		RXMulticast:     values[7],
		TXBytes:         values[8],
		TXPackets:       values[9],
		TXErrors:        values[10],
		TXDropped:       values[11],
		TXFIFOErrors:    values[12],
		TXCollisions:    values[13],
		TXCarrierErrors: values[14],
		TXCompressed:    values[15],
	}, nil
}

// parseUintFields parses one slice of decimal counter fields.
func parseUintFields(fields []string) ([]uint64, error) {
	values := make([]uint64, 0, len(fields))
	for _, field := range fields {
		value, err := strconv.ParseUint(field, 10, 64)
		if err != nil {
			return nil, err
		}

		values = append(values, value)
	}

	return values, nil
}

// parseSockStat parses /proc/net/sockstat into one structured snapshot block.
func parseSockStat(path string) (metrics.NetworkSockStat, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- path is configured by the application
	if err != nil {
		return metrics.NetworkSockStat{}, fmt.Errorf("read %s: %w", path, err)
	}

	sockStat := metrics.NetworkSockStat{}
	for _, rawLine := range strings.Split(string(data), "\n") {
		fields := strings.Fields(strings.TrimSpace(rawLine))
		if len(fields) < 3 {
			continue
		}

		applySockStatLine(&sockStat, fields[0], parseNamedUintPairs(fields[1:]))
	}

	return sockStat, nil
}

func applySockStatLine(sockStat *metrics.NetworkSockStat, section string, values map[string]uint64) {
	if applier, ok := sockStatAppliers[section]; ok {
		applier(sockStat, values)
	}
}

var sockStatAppliers = map[string]func(*metrics.NetworkSockStat, map[string]uint64){
	"sockets:": func(sockStat *metrics.NetworkSockStat, values map[string]uint64) {
		sockStat.SocketsUsed = values["used"]
	},
	"TCP:": func(sockStat *metrics.NetworkSockStat, values map[string]uint64) {
		sockStat.TCPInUse = values["inuse"]
		sockStat.TCPOrphan = values["orphan"]
		sockStat.TCPTimeWait = values["tw"]
		sockStat.TCPAlloc = values["alloc"]
		sockStat.TCPMem = values["mem"]
	},
	"UDP:": func(sockStat *metrics.NetworkSockStat, values map[string]uint64) {
		sockStat.UDPInUse = values["inuse"]
		sockStat.UDPMem = values["mem"]
	},
	"UDPLITE:": func(sockStat *metrics.NetworkSockStat, values map[string]uint64) {
		sockStat.UDPLiteInUse = values["inuse"]
	},
	"RAW:": func(sockStat *metrics.NetworkSockStat, values map[string]uint64) {
		sockStat.RAWInUse = values["inuse"]
	},
	"FRAG:": func(sockStat *metrics.NetworkSockStat, values map[string]uint64) {
		sockStat.FragInUse = values["inuse"]
		sockStat.FragMemory = values["memory"]
	},
}

// parseNamedUintPairs parses alternating key/value decimal pairs into one map.
func parseNamedUintPairs(fields []string) map[string]uint64 {
	values := make(map[string]uint64, len(fields)/2)

	for fieldIndex := 0; fieldIndex+1 < len(fields); fieldIndex += 2 {
		value, err := strconv.ParseUint(fields[fieldIndex+1], 10, 64)
		if err != nil {
			continue
		}

		values[fields[fieldIndex]] = value
	}

	return values
}

// parseSoftIRQs parses /proc/softirqs into the network softirq snapshot block.
func parseSoftIRQs(path string) (metrics.NetworkSoftIRQSnapshot, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- path is configured by the application
	if err != nil {
		return metrics.NetworkSoftIRQSnapshot{}, fmt.Errorf("read %s: %w", path, err)
	}

	snapshot := metrics.NetworkSoftIRQSnapshot{
		NETRXPerCPU: make([]uint64, 0),
		NETTXPerCPU: make([]uint64, 0),
	}

	for _, rawLine := range strings.Split(string(data), "\n") {
		name, values, ok, err := parseSoftIRQLine(rawLine)
		if err != nil {
			return metrics.NetworkSoftIRQSnapshot{}, err
		}
		if !ok {
			continue
		}

		assignSoftIRQValues(&snapshot, name, values)
	}

	return snapshot, nil
}

func parseSoftIRQLine(rawLine string) (string, []uint64, bool, error) {
	line := strings.TrimSpace(rawLine)
	if line == "" || !strings.Contains(line, ":") {
		return "", nil, false, nil
	}

	name, valuesText, ok := strings.Cut(line, ":")
	if !ok {
		return "", nil, false, nil
	}

	values, err := parseUintFields(strings.Fields(valuesText))
	if err != nil {
		return "", nil, false, fmt.Errorf("parse %s softirq values: %w", strings.TrimSpace(name), err)
	}

	return strings.TrimSpace(name), values, true, nil
}

func assignSoftIRQValues(snapshot *metrics.NetworkSoftIRQSnapshot, name string, values []uint64) {
	switch name {
	case "NET_RX":
		snapshot.NETRXPerCPU = values
		snapshot.NETRXTotal = sumUint64(values)
	case "NET_TX":
		snapshot.NETTXPerCPU = values
		snapshot.NETTXTotal = sumUint64(values)
	}
}

// sumUint64 sums one uint64 slice.
func sumUint64(values []uint64) uint64 {
	var total uint64
	for _, value := range values {
		total += value
	}

	return total
}

// newSocketSummaryAggregator creates an initialized socket summary accumulator.
func newSocketSummaryAggregator() socketSummaryAggregator {
	return socketSummaryAggregator{
		byState:        make(map[string]uint64),
		localPortCount: make(map[string]uint64),
		remoteIPCount:  make(map[string]uint64),
	}
}

// consumeTable reads one proc socket table and merges its data into the
// accumulator.
func (a *socketSummaryAggregator) consumeTable(path string, protocol string) error {
	lines, err := readSocketTableLines(path)
	if err != nil {
		return err
	}

	for _, line := range lines {
		if err := a.consumeSocketLine(path, protocol, line); err != nil {
			return err
		}
	}

	return nil
}

func readSocketTableLines(path string) ([]string, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- path is configured by the application
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) <= 1 {
		return nil, nil
	}

	return lines[1:], nil
}

func (a *socketSummaryAggregator) consumeSocketLine(path string, protocol string, rawLine string) error {
	line := strings.TrimSpace(rawLine)
	if line == "" {
		return nil
	}

	localPort, remoteIP, stateName, ok, err := parseSocketTableLine(line)
	if err != nil {
		return fmt.Errorf("parse socket row from %s: %w", path, err)
	}
	if ok {
		a.record(protocol, localPort, remoteIP, stateName)
	}

	return nil
}

func parseSocketTableLine(line string) (uint16, string, string, bool, error) {
	fields := strings.Fields(line)
	if len(fields) < 4 {
		return 0, "", "", false, nil
	}

	_, localPort, err := parseProcAddress(fields[1])
	if err != nil {
		return 0, "", "", false, err
	}

	remoteIP, _, err := parseProcAddress(fields[2])
	if err != nil {
		return 0, "", "", false, err
	}

	return localPort, remoteIP, decodeSocketState(fields[3]), true, nil
}

// build materializes the final socket snapshot from the accumulator state.
func (a socketSummaryAggregator) build(sockStat metrics.NetworkSockStat) metrics.NetworkSocketSnapshot {
	return metrics.NetworkSocketSnapshot{
		Total:         a.total,
		ByState:       copyCountMap(a.byState),
		TopLocalPorts: buildTopPortCounts(a.localPortCount, topSocketRankLimit),
		TopRemoteIPs:  buildTopIPCounts(a.remoteIPCount, topSocketRankLimit),
		SockStat:      sockStat,
	}
}

// record merges one parsed socket row into the accumulator.
func (a *socketSummaryAggregator) record(
	protocol string,
	localPort uint16,
	remoteIP string,
	stateName string,
) {
	a.total.All++
	incrementProtocolTotal(&a.total, protocol)

	if stateName != "" {
		a.byState[stateName]++
	}

	portKey := fmt.Sprintf("%s:%d", protocol, localPort)
	a.localPortCount[portKey]++

	if shouldCountRemoteIP(remoteIP) {
		a.remoteIPCount[remoteIP]++
	}
}

func incrementProtocolTotal(total *metrics.NetworkSocketTotals, protocol string) {
	switch protocol {
	case "tcp":
		total.TCP++
	case "tcp6":
		total.TCP6++
	case "udp":
		total.UDP++
	case "udp6":
		total.UDP6++
	}
}

func shouldCountRemoteIP(remoteIP string) bool {
	if remoteIP == "" {
		return false
	}

	addr, err := netip.ParseAddr(remoteIP)
	return err == nil && !addr.IsUnspecified()
}

// copyCountMap copies one count map so callers cannot mutate accumulator state.
func copyCountMap(source map[string]uint64) map[string]uint64 {
	result := make(map[string]uint64, len(source))
	for key, value := range source {
		result[key] = value
	}

	return result
}

// buildTopPortCounts converts port-count maps into ranked snapshot items.
func buildTopPortCounts(source map[string]uint64, limit int) []metrics.NetworkSocketPortCount {
	items := rankPortCountItems(source, limit)

	result := make([]metrics.NetworkSocketPortCount, 0, len(items))
	for _, item := range items {
		portCount, ok := parseTopPortCount(item.key, item.count)
		if !ok {
			continue
		}

		result = append(result, portCount)
	}

	return result
}

type rankedPortCount struct {
	key   string
	count uint64
}

func rankPortCountItems(source map[string]uint64, limit int) []rankedPortCount {
	items := make([]rankedPortCount, 0, len(source))
	for key, count := range source {
		items = append(items, rankedPortCount{key: key, count: count})
	}

	sort.Slice(items, func(i int, j int) bool {
		if items[i].count == items[j].count {
			return items[i].key < items[j].key
		}

		return items[i].count > items[j].count
	})

	if len(items) > limit {
		items = items[:limit]
	}

	return items
}

func parseTopPortCount(key string, count uint64) (metrics.NetworkSocketPortCount, bool) {
	protocol, portText, ok := strings.Cut(key, ":")
	if !ok {
		return metrics.NetworkSocketPortCount{}, false
	}

	portValue, err := strconv.ParseUint(portText, 10, 16)
	if err != nil {
		return metrics.NetworkSocketPortCount{}, false
	}

	return metrics.NetworkSocketPortCount{
		Protocol: protocol,
		Port:     uint16(portValue),
		Count:    count,
	}, true
}

// buildTopIPCounts converts remote-IP count maps into ranked snapshot items.
func buildTopIPCounts(source map[string]uint64, limit int) []metrics.NetworkSocketIPCount {
	items := make([]metrics.NetworkSocketIPCount, 0, len(source))
	for ip, count := range source {
		items = append(items, metrics.NetworkSocketIPCount{
			IP:    ip,
			Count: count,
		})
	}

	sort.Slice(items, func(i int, j int) bool {
		if items[i].Count == items[j].Count {
			return items[i].IP < items[j].IP
		}

		return items[i].Count > items[j].Count
	})

	if len(items) > limit {
		items = items[:limit]
	}

	return items
}

// parseProcAddress parses one procfs socket address field into IP and port.
func parseProcAddress(value string) (string, uint16, error) {
	addressHex, portHex, ok := strings.Cut(value, ":")
	if !ok {
		return "", 0, fmt.Errorf("missing port separator in %q", value)
	}

	portValue, err := strconv.ParseUint(portHex, 16, 16)
	if err != nil {
		return "", 0, err
	}

	ipValue, err := decodeProcAddressHex(addressHex)
	if err != nil {
		return "", 0, err
	}

	return ipValue, uint16(portValue), nil
}

// decodeProcAddressHex decodes one procfs socket-address hex value into a
// human-readable IP string.
func decodeProcAddressHex(addressHex string) (string, error) {
	switch len(addressHex) {
	case 8:
		return decodeProcIPv4Address(addressHex)
	case 32:
		return decodeProcIPv6Address(addressHex)
	default:
		return "", fmt.Errorf("unsupported address hex length %d", len(addressHex))
	}
}

func decodeProcIPv4Address(addressHex string) (string, error) {
	bytesValue, err := hex.DecodeString(addressHex)
	if err != nil {
		return "", err
	}
	if len(bytesValue) != 4 {
		return "", errors.New("unexpected IPv4 length")
	}

	return netip.AddrFrom4([4]byte{
		bytesValue[3],
		bytesValue[2],
		bytesValue[1],
		bytesValue[0],
	}).String(), nil
}

func decodeProcIPv6Address(addressHex string) (string, error) {
	bytesValue, err := hex.DecodeString(addressHex)
	if err != nil {
		return "", err
	}
	if len(bytesValue) != 16 {
		return "", errors.New("unexpected IPv6 length")
	}

	return netip.AddrFrom16(reorderProcIPv6Bytes(bytesValue)).String(), nil
}

func reorderProcIPv6Bytes(bytesValue []byte) [16]byte {
	var reordered [16]byte
	for chunkIndex := 0; chunkIndex < 4; chunkIndex++ {
		sourceOffset := chunkIndex * 4
		reordered[sourceOffset] = bytesValue[sourceOffset+3]
		reordered[sourceOffset+1] = bytesValue[sourceOffset+2]
		reordered[sourceOffset+2] = bytesValue[sourceOffset+1]
		reordered[sourceOffset+3] = bytesValue[sourceOffset]
	}

	return reordered
}

// decodeSocketState maps procfs state codes into stable state labels.
func decodeSocketState(code string) string {
	normalized := strings.ToUpper(strings.TrimSpace(code))
	if stateName, ok := socketStateNames[normalized]; ok {
		return stateName
	}

	return normalized
}

var socketStateNames = map[string]string{
	"01": "ESTABLISHED",
	"02": "SYN_SENT",
	"03": "SYN_RECV",
	"04": "FIN_WAIT1",
	"05": "FIN_WAIT2",
	"06": "TIME_WAIT",
	"07": "CLOSE",
	"08": "CLOSE_WAIT",
	"09": "LAST_ACK",
	"0A": "LISTEN",
	"0B": "CLOSING",
	"0C": "NEW_SYN_RECV",
}

// readStatUint reads one optional decimal statistics file with a fallback.
func readStatUint(basePath string, fileName string, fallback uint64) uint64 {
	value, ok := readOptionalUintValue(filepath.Join(basePath, fileName))
	if !ok {
		return fallback
	}

	return value
}

// readOptionalUint reads one optional decimal file into a pointer value.
func readOptionalUint(path string) *uint64 {
	value, ok := readOptionalUintValue(path)
	if !ok {
		return nil
	}

	return &value
}

// readOptionalUintValue reads one optional decimal file into a plain value and
// presence flag.
func readOptionalUintValue(path string) (uint64, bool) {
	data, err := os.ReadFile(path) // #nosec G304 -- path is configured by the application
	if err != nil {
		return 0, false
	}

	value, err := strconv.ParseUint(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return 0, false
	}

	return value, true
}

// readOptionalBool reads one optional bool-like file into a pointer value.
func readOptionalBool(path string) *bool {
	value, ok := readOptionalUintValue(path)
	if !ok {
		return nil
	}

	boolValue := value != 0
	return &boolValue
}

// readOptionalString reads one optional text file into a trimmed string.
func readOptionalString(path string) string {
	data, err := os.ReadFile(path) // #nosec G304 -- path is configured by the application
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(data))
}

// readOptionalNamedString reads one optional text file into a pointer string.
func readOptionalNamedString(path string) *string {
	value := readOptionalString(path)
	if value == "" {
		return nil
	}

	return &value
}

// procNetDevStats stores the /proc/net/dev counters used as interface
// fallbacks.
type procNetDevStats struct {
	RXBytes         uint64
	RXPackets       uint64
	RXErrors        uint64
	RXDropped       uint64
	RXFIFOErrors    uint64
	RXFrameErrors   uint64
	RXCompressed    uint64
	RXMulticast     uint64
	TXBytes         uint64
	TXPackets       uint64
	TXErrors        uint64
	TXDropped       uint64
	TXFIFOErrors    uint64
	TXCollisions    uint64
	TXCarrierErrors uint64
	TXCompressed    uint64
}

// socketSummaryAggregator accumulates summarized socket counts during one poll
// cycle.
type socketSummaryAggregator struct {
	total          metrics.NetworkSocketTotals
	byState        map[string]uint64
	localPortCount map[string]uint64
	remoteIPCount  map[string]uint64
}
