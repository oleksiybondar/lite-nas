package workers

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"slices"
	"sort"
	"strconv"
	"strings"

	"lite-nas/shared/metrics"
)

// OutputWriter renders the selected CLI output format.
type OutputWriter interface {
	WriteCurrent(writer io.Writer, snapshot metrics.NetworkMetricsSnapshot, selection CurrentSelection) error
	WriteHistory(writer io.Writer, history []metrics.NetworkMetricsSnapshot) error
}

type outputWriter struct{}

// NewOutputWriter creates an output rendering worker.
func NewOutputWriter() OutputWriter {
	return outputWriter{}
}

// WriteCurrent renders a human-readable snapshot view.
func (outputWriter) WriteCurrent(
	writer io.Writer,
	snapshot metrics.NetworkMetricsSnapshot,
	selection CurrentSelection,
) error {
	sections := make([]string, 0, 4)

	if selection.Interfaces {
		sections = append(sections, renderInterfacesSection(snapshot))
	}
	if selection.Protocols {
		sections = append(sections, renderProtocolsSection(snapshot))
	}
	if selection.Sockets {
		sections = append(sections, renderSocketsSection(snapshot))
	}
	if selection.Pressure {
		sections = append(sections, renderPressureSection(snapshot))
	}

	_, err := fmt.Fprintf(writer, "%s\n", strings.Join(sections, "\n\n----\n\n"))
	return err
}

// WriteHistory renders history as pretty-printed JSON.
func (outputWriter) WriteHistory(writer io.Writer, history []metrics.NetworkMetricsSnapshot) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(history)
}

func renderInterfacesSection(snapshot metrics.NetworkMetricsSnapshot) string {
	lines := []string{"Interfaces", "----"}

	if len(snapshot.Interfaces) == 0 {
		lines = append(lines, "No interfaces in snapshot.")
		return strings.Join(lines, "\n")
	}

	for _, iface := range snapshot.Interfaces {
		lines = append(lines, renderInterfaceLine(iface))
	}

	return strings.Join(lines, "\n")
}

func renderInterfaceLine(iface metrics.NetworkInterfaceSnapshot) string {
	parts := []string{
		iface.Name,
		fmt.Sprintf("state=%s", fallbackString(iface.OperState, "unknown")),
		fmt.Sprintf("rx=%dB", iface.Statistics.RXBytes),
		fmt.Sprintf("tx=%dB", iface.Statistics.TXBytes),
	}

	if iface.Kind != "" {
		parts = append(parts, fmt.Sprintf("kind=%s", iface.Kind))
	}
	if bus := optionalString(iface.Bus); bus != "" {
		parts = append(parts, fmt.Sprintf("bus=%s", bus))
	}
	if adapterText := adapterDescription(iface.Adapter); adapterText != "" {
		parts = append(parts, fmt.Sprintf("adapter=%s", adapterText))
	}

	return strings.Join(parts, " ")
}

func renderProtocolsSection(snapshot metrics.NetworkMetricsSnapshot) string {
	lines := []string{"Protocols", "----"}

	for _, protocol := range protocolSections(snapshot.Protocols) {
		if len(protocol.Counters) == 0 {
			continue
		}

		lines = append(lines, fmt.Sprintf("%s: %s", protocol.Name, formatCounterMap(protocol.Counters)))
	}

	if len(lines) == 2 {
		lines = append(lines, "No protocol counters in snapshot.")
	}

	return strings.Join(lines, "\n")
}

func renderSocketsSection(snapshot metrics.NetworkMetricsSnapshot) string {
	lines := []string{
		"Sockets",
		"----",
		fmt.Sprintf(
			"Totals: all=%d tcp=%d tcp6=%d udp=%d udp6=%d",
			snapshot.Sockets.Total.All,
			snapshot.Sockets.Total.TCP,
			snapshot.Sockets.Total.TCP6,
			snapshot.Sockets.Total.UDP,
			snapshot.Sockets.Total.UDP6,
		),
		fmt.Sprintf("By State: %s", formatCountMap(snapshot.Sockets.ByState)),
		fmt.Sprintf("Top Local Ports: %s", formatTopLocalPorts(snapshot.Sockets.TopLocalPorts)),
		fmt.Sprintf("Top Remote IPs: %s", formatTopRemoteIPs(snapshot.Sockets.TopRemoteIPs)),
		fmt.Sprintf(
			"SockStat: used=%d tcp_inuse=%d tcp_tw=%d udp_inuse=%d raw_inuse=%d",
			snapshot.Sockets.SockStat.SocketsUsed,
			snapshot.Sockets.SockStat.TCPInUse,
			snapshot.Sockets.SockStat.TCPTimeWait,
			snapshot.Sockets.SockStat.UDPInUse,
			snapshot.Sockets.SockStat.RAWInUse,
		),
	}

	return strings.Join(lines, "\n")
}

func renderPressureSection(snapshot metrics.NetworkMetricsSnapshot) string {
	return strings.Join([]string{
		"Pressure",
		"----",
		fmt.Sprintf("NET_RX Total: %d", snapshot.KernelPressure.SoftIRQs.NETRXTotal),
		fmt.Sprintf("NET_TX Total: %d", snapshot.KernelPressure.SoftIRQs.NETTXTotal),
		fmt.Sprintf("NET_RX Per CPU: %s", formatUintSlice(snapshot.KernelPressure.SoftIRQs.NETRXPerCPU)),
		fmt.Sprintf("NET_TX Per CPU: %s", formatUintSlice(snapshot.KernelPressure.SoftIRQs.NETTXPerCPU)),
	}, "\n")
}

func adapterDescription(adapter *metrics.NetworkInterfaceAdapter) string {
	if adapter == nil {
		return ""
	}

	for _, description := range []string{
		optionalString(adapter.Description),
		adapterVendorAndDevice(adapter),
		optionalString(adapter.DeviceName),
		optionalString(adapter.VendorName),
		adapterPCIIdentifier(adapter),
	} {
		if description != "" {
			return description
		}
	}

	return ""
}

func fallbackString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func optionalString(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

func adapterVendorAndDevice(adapter *metrics.NetworkInterfaceAdapter) string {
	vendorName := optionalString(adapter.VendorName)
	deviceName := optionalString(adapter.DeviceName)
	if vendorName == "" || deviceName == "" {
		return ""
	}

	return fmt.Sprintf("%s %s", vendorName, deviceName)
}

func adapterPCIIdentifier(adapter *metrics.NetworkInterfaceAdapter) string {
	if adapter.VendorID == "" || adapter.DeviceID == "" {
		return ""
	}

	return fmt.Sprintf("%s/%s", adapter.VendorID, adapter.DeviceID)
}

type protocolSection struct {
	Name     string
	Counters metrics.NetworkCounterGroup
}

func protocolSections(snapshot metrics.NetworkProtocolSnapshot) []protocolSection {
	return []protocolSection{
		{Name: "ip", Counters: snapshot.IP},
		{Name: "icmp", Counters: snapshot.ICMP},
		{Name: "tcp", Counters: snapshot.TCP},
		{Name: "udp", Counters: snapshot.UDP},
		{Name: "udplite", Counters: snapshot.UDPLite},
		{Name: "ip_ext", Counters: snapshot.IPExt},
		{Name: "tcp_ext", Counters: snapshot.TCPExt},
	}
}

func formatCounterMap(values map[string]uint64) string {
	if len(values) == 0 {
		return "none"
	}

	keys := slices.Collect(maps.Keys(values))
	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, key := range keys {
		pairs = append(pairs, fmt.Sprintf("%s=%d", key, values[key]))
	}

	return strings.Join(pairs, ", ")
}

func formatCountMap(values map[string]uint64) string {
	if len(values) == 0 {
		return "none"
	}

	keys := slices.Collect(maps.Keys(values))
	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, key := range keys {
		pairs = append(pairs, fmt.Sprintf("%s=%d", key, values[key]))
	}

	return strings.Join(pairs, ", ")
}

func formatTopLocalPorts(values []metrics.NetworkSocketPortCount) string {
	if len(values) == 0 {
		return "none"
	}

	items := make([]string, 0, len(values))
	for _, value := range values {
		items = append(items, fmt.Sprintf("%s/%d=%d", value.Protocol, value.Port, value.Count))
	}

	return strings.Join(items, ", ")
}

func formatTopRemoteIPs(values []metrics.NetworkSocketIPCount) string {
	if len(values) == 0 {
		return "none"
	}

	items := make([]string, 0, len(values))
	for _, value := range values {
		items = append(items, fmt.Sprintf("%s=%d", value.IP, value.Count))
	}

	return strings.Join(items, ", ")
}

func formatUintSlice(values []uint64) string {
	if len(values) == 0 {
		return "[]"
	}

	items := make([]string, 0, len(values))
	for _, value := range values {
		items = append(items, strconv.FormatUint(value, 10))
	}

	return "[" + strings.Join(items, ", ") + "]"
}
