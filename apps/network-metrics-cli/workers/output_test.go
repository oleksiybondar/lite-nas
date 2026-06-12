package workers

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"lite-nas/shared/metrics"
)

// Requirements: network-metrics-cli/FR-002, network-metrics-cli/FR-003, network-metrics-cli/IR-002
func TestOutputWriterWritesCurrentSnapshotInHumanReadableFormat(t *testing.T) {
	t.Parallel()

	output := mustWriteCurrentSnapshot(t, networkSnapshotFixture(), CurrentSelection{
		Interfaces: true,
		Protocols:  true,
		Sockets:    true,
		Pressure:   true,
	})

	want := "Interfaces\n----\nenp2s0 state=up rx=1000B tx=2000B kind=physical bus=pci adapter=Intel I211\n\n----\n\nProtocols\n----\ntcp: CurrEstab=2\n\n----\n\nSockets\n----\nTotals: all=4 tcp=2 tcp6=0 udp=2 udp6=0\nBy State: ESTABLISHED=2\nTop Local Ports: tcp/22=1\nTop Remote IPs: 192.168.1.10=1\nSockStat: used=4 tcp_inuse=2 tcp_tw=0 udp_inuse=2 raw_inuse=0\n\n----\n\nPressure\n----\nNET_RX Total: 10\nNET_TX Total: 20\nNET_RX Per CPU: [1, 2]\nNET_TX Per CPU: [3, 4]\n"
	if output != want {
		t.Fatalf("WriteCurrent() output = %q, want %q", output, want)
	}
}

// Requirements: network-metrics-cli/FR-003, network-metrics-cli/IR-002
func TestOutputWriterWritesOnlySelectedCurrentSection(t *testing.T) {
	t.Parallel()

	snapshot := metrics.NetworkMetricsSnapshot{
		Protocols: metrics.NetworkProtocolSnapshot{
			UDP: metrics.NetworkCounterGroup{"InDatagrams": 7},
		},
	}

	output := mustWriteCurrentSnapshot(t, snapshot, CurrentSelection{Protocols: true})
	want := "Protocols\n----\nudp: InDatagrams=7\n"
	if output != want {
		t.Fatalf("WriteCurrent() output = %q, want %q", output, want)
	}
}

// Requirements: network-metrics-cli/FR-002, network-metrics-cli/FR-003, network-metrics-cli/IR-002
func TestOutputWriterCurrentSnapshotIncludesExpectedSections(t *testing.T) {
	t.Parallel()

	output := mustWriteCurrentSnapshot(t, networkSnapshotFixture(), CurrentSelection{
		Interfaces: true,
		Protocols:  true,
		Sockets:    true,
		Pressure:   true,
	})

	assertContains(t, output, "Interfaces\n----\nenp2s0 state=up rx=1000B tx=2000B kind=physical bus=pci adapter=Intel I211")
	assertContains(t, output, "Protocols\n----\ntcp: CurrEstab=2")
	assertContains(t, output, "Sockets\n----\nTotals: all=4 tcp=2 tcp6=0 udp=2 udp6=0")
	assertContains(t, output, "Pressure\n----\nNET_RX Total: 10\nNET_TX Total: 20")
}

// Requirements: network-metrics-cli/FR-005, network-metrics-cli/IR-002
func TestOutputWriterWritesHistoryAsIndentedJSON(t *testing.T) {
	t.Parallel()

	writer := NewOutputWriter()
	var output bytes.Buffer

	history := []metrics.NetworkMetricsSnapshot{
		{
			Timestamp: time.Unix(1700000000, 0).UTC(),
			Interfaces: []metrics.NetworkInterfaceSnapshot{
				{Name: "enp2s0", Kind: "unknown"},
			},
		},
	}

	err := writer.WriteHistory(outputWriterBuffer(&output), history)
	if err != nil {
		t.Fatalf("WriteHistory() error = %v", err)
	}

	want := "[\n  {\n    \"timestamp\": \"2023-11-14T22:13:20Z\",\n    \"interfaces\": [\n      {\n        \"name\": \"enp2s0\",\n        \"ifindex\": null,\n        \"address\": \"\",\n        \"mtu\": null,\n        \"oper_state\": \"\",\n        \"carrier_up\": null,\n        \"speed_mbps\": null,\n        \"duplex\": \"\",\n        \"tx_queue_len\": null,\n        \"kind\": \"unknown\",\n        \"bus\": null,\n        \"adapter\": null,\n        \"statistics\": {\n          \"rx_bytes\": 0,\n          \"rx_packets\": 0,\n          \"rx_errors\": 0,\n          \"rx_dropped\": 0,\n          \"rx_fifo_errors\": 0,\n          \"rx_frame_errors\": 0,\n          \"rx_compressed\": 0,\n          \"rx_multicast\": 0,\n          \"rx_crc_errors\": 0,\n          \"rx_length_errors\": 0,\n          \"rx_missed_errors\": 0,\n          \"rx_over_errors\": 0,\n          \"rx_nohandler\": 0,\n          \"tx_bytes\": 0,\n          \"tx_packets\": 0,\n          \"tx_errors\": 0,\n          \"tx_dropped\": 0,\n          \"tx_fifo_errors\": 0,\n          \"tx_collisions\": 0,\n          \"tx_carrier_errors\": 0,\n          \"tx_compressed\": 0,\n          \"tx_aborted_errors\": 0,\n          \"tx_heartbeat_errors\": 0,\n          \"tx_window_errors\": 0\n        }\n      }\n    ],\n    \"protocols\": {\n      \"ip\": null,\n      \"icmp\": null,\n      \"tcp\": null,\n      \"udp\": null,\n      \"udplite\": null,\n      \"ip_ext\": null,\n      \"tcp_ext\": null\n    },\n    \"sockets\": {\n      \"total\": {\n        \"all\": 0,\n        \"tcp\": 0,\n        \"tcp6\": 0,\n        \"udp\": 0,\n        \"udp6\": 0\n      },\n      \"by_state\": null,\n      \"top_local_ports\": null,\n      \"top_remote_ips\": null,\n      \"sockstat\": {\n        \"sockets_used\": 0,\n        \"tcp_inuse\": 0,\n        \"tcp_orphan\": 0,\n        \"tcp_time_wait\": 0,\n        \"tcp_alloc\": 0,\n        \"tcp_mem\": 0,\n        \"udp_inuse\": 0,\n        \"udp_mem\": 0,\n        \"udplite_inuse\": 0,\n        \"raw_inuse\": 0,\n        \"frag_inuse\": 0,\n        \"frag_memory\": 0\n      }\n    },\n    \"kernel_pressure\": {\n      \"softirqs\": {\n        \"net_rx_total\": 0,\n        \"net_tx_total\": 0,\n        \"net_rx_per_cpu\": null,\n        \"net_tx_per_cpu\": null\n      }\n    }\n  }\n]\n"
	if output.String() != want {
		t.Fatalf("WriteHistory() output = %q, want %q", output.String(), want)
	}
}

func outputWriterBuffer(buffer *bytes.Buffer) *bytes.Buffer {
	return buffer
}

func mustWriteCurrentSnapshot(
	t *testing.T,
	snapshot metrics.NetworkMetricsSnapshot,
	selection CurrentSelection,
) string {
	t.Helper()

	writer := NewOutputWriter()
	var output bytes.Buffer

	if err := writer.WriteCurrent(outputWriterBuffer(&output), snapshot, selection); err != nil {
		t.Fatalf("WriteCurrent() error = %v", err)
	}

	return output.String()
}

func networkSnapshotFixture() metrics.NetworkMetricsSnapshot {
	vendorName := "Intel"
	deviceName := "I211"

	return metrics.NetworkMetricsSnapshot{
		Timestamp: time.Unix(1700000000, 0).UTC(),
		Interfaces: []metrics.NetworkInterfaceSnapshot{
			{
				Name:      "enp2s0",
				OperState: "up",
				Kind:      "physical",
				Bus:       stringPointer("pci"),
				Adapter: &metrics.NetworkInterfaceAdapter{
					VendorName: &vendorName,
					DeviceName: &deviceName,
				},
				Statistics: metrics.NetworkInterfaceStatistics{
					RXBytes: 1000,
					TXBytes: 2000,
				},
			},
		},
		Protocols: metrics.NetworkProtocolSnapshot{
			TCP: metrics.NetworkCounterGroup{"CurrEstab": 2},
		},
		Sockets: metrics.NetworkSocketSnapshot{
			Total:         metrics.NetworkSocketTotals{All: 4, TCP: 2, UDP: 2},
			ByState:       map[string]uint64{"ESTABLISHED": 2},
			TopLocalPorts: []metrics.NetworkSocketPortCount{{Protocol: "tcp", Port: 22, Count: 1}},
			TopRemoteIPs:  []metrics.NetworkSocketIPCount{{IP: "192.168.1.10", Count: 1}},
			SockStat:      metrics.NetworkSockStat{SocketsUsed: 4, TCPInUse: 2, UDPInUse: 2},
		},
		KernelPressure: metrics.NetworkKernelPressureSnapshot{
			SoftIRQs: metrics.NetworkSoftIRQSnapshot{
				NETRXTotal:  10,
				NETTXTotal:  20,
				NETRXPerCPU: []uint64{1, 2},
				NETTXPerCPU: []uint64{3, 4},
			},
		},
	}
}

func assertContains(t *testing.T, output string, want string) {
	t.Helper()

	if !strings.Contains(output, want) {
		t.Fatalf("output = %q, want substring %q", output, want)
	}
}

func stringPointer(value string) *string {
	return &value
}
