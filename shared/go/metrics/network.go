package metrics

import "time"

// NetworkCounterGroup stores one named group of raw kernel counters.
//
// The map key is the kernel-provided counter name, and the value is the raw
// cumulative counter value collected for the current snapshot.
type NetworkCounterGroup map[string]uint64

// NetworkMetricsSnapshot is the top-level network metrics snapshot payload.
//
// Unlike the older system and ZFS snapshot payloads, this contract uses
// explicit JSON tags to keep the browser- and event-facing JSON shape stable
// with lowercase section names.
type NetworkMetricsSnapshot struct {
	// Timestamp is the time at which the snapshot was collected.
	Timestamp time.Time `json:"timestamp"`

	// Interfaces contains per-interface factual network metrics.
	Interfaces []NetworkInterfaceSnapshot `json:"interfaces"`

	// Protocols contains grouped raw protocol counters.
	Protocols NetworkProtocolSnapshot `json:"protocols"`

	// Sockets contains summarized socket metrics.
	Sockets NetworkSocketSnapshot `json:"sockets"`

	// KernelPressure contains kernel-side network pressure indicators.
	KernelPressure NetworkKernelPressureSnapshot `json:"kernel_pressure"`
}

// NetworkInterfaceSnapshot represents one host network interface snapshot.
type NetworkInterfaceSnapshot struct {
	// Name is the kernel interface name, for example "enp2s0".
	Name string `json:"name"`

	// IfIndex is the interface index when available from the host.
	IfIndex *uint64 `json:"ifindex"`

	// Address is the interface link-layer address as reported by the host.
	Address string `json:"address"`

	// MTU is the interface MTU when available from the host.
	MTU *uint64 `json:"mtu"`

	// OperState is the kernel operational state, for example "up" or "down".
	OperState string `json:"oper_state"`

	// CarrierUp reports whether carrier is currently detected when available.
	CarrierUp *bool `json:"carrier_up"`

	// SpeedMbps is the reported link speed in megabits per second when
	// available from the host.
	SpeedMbps *uint64 `json:"speed_mbps"`

	// Duplex is the reported duplex mode when available from the host.
	Duplex string `json:"duplex"`

	// TxQueueLen is the transmit queue length when available from the host.
	TxQueueLen *uint64 `json:"tx_queue_len"`

	// Kind classifies the interface role using a stable service-level category.
	Kind string `json:"kind"`

	// Bus classifies the backing device attachment bus when locally resolvable.
	Bus *string `json:"bus"`

	// Adapter contains stable adapter identity metadata when locally
	// resolvable for the interface.
	Adapter *NetworkInterfaceAdapter `json:"adapter"`

	// Statistics contains factual per-interface counters.
	Statistics NetworkInterfaceStatistics `json:"statistics"`
}

// NetworkInterfaceAdapter describes adapter identity metadata for one
// interface.
type NetworkInterfaceAdapter struct {
	// VendorID is the stable adapter vendor identifier when available.
	VendorID string `json:"vendor_id"`

	// DeviceID is the stable adapter device identifier when available.
	DeviceID string `json:"device_id"`

	// VendorName is the human-readable vendor name when locally resolvable.
	VendorName *string `json:"vendor_name"`

	// DeviceName is the human-readable device name when locally resolvable.
	DeviceName *string `json:"device_name"`

	// Description is the human-readable adapter description when locally
	// resolvable.
	Description *string `json:"description"`
}

// NetworkInterfaceStatistics stores factual per-interface counters.
type NetworkInterfaceStatistics struct {
	// RXBytes is the cumulative received byte count.
	RXBytes uint64 `json:"rx_bytes"`

	// RXPackets is the cumulative received packet count.
	RXPackets uint64 `json:"rx_packets"`

	// RXErrors is the cumulative received error count.
	RXErrors uint64 `json:"rx_errors"`

	// RXDropped is the cumulative received dropped-packet count.
	RXDropped uint64 `json:"rx_dropped"`

	// RXFIFOErrors is the cumulative receive FIFO error count.
	RXFIFOErrors uint64 `json:"rx_fifo_errors"`

	// RXFrameErrors is the cumulative receive frame error count.
	RXFrameErrors uint64 `json:"rx_frame_errors"`

	// RXCompressed is the cumulative received compressed-packet count.
	RXCompressed uint64 `json:"rx_compressed"`

	// RXMulticast is the cumulative received multicast-packet count.
	RXMulticast uint64 `json:"rx_multicast"`

	// RXCRCError is the cumulative receive CRC error count.
	RXCRCErrors uint64 `json:"rx_crc_errors"`

	// RXLengthErrors is the cumulative receive length error count.
	RXLengthErrors uint64 `json:"rx_length_errors"`

	// RXMissedErrors is the cumulative receive missed error count.
	RXMissedErrors uint64 `json:"rx_missed_errors"`

	// RXOverErrors is the cumulative receive overrun error count.
	RXOverErrors uint64 `json:"rx_over_errors"`

	// RXNoHandler is the cumulative receive no-handler count.
	RXNoHandler uint64 `json:"rx_nohandler"`

	// TXBytes is the cumulative transmitted byte count.
	TXBytes uint64 `json:"tx_bytes"`

	// TXPackets is the cumulative transmitted packet count.
	TXPackets uint64 `json:"tx_packets"`

	// TXErrors is the cumulative transmit error count.
	TXErrors uint64 `json:"tx_errors"`

	// TXDropped is the cumulative transmit dropped-packet count.
	TXDropped uint64 `json:"tx_dropped"`

	// TXFIFOErrors is the cumulative transmit FIFO error count.
	TXFIFOErrors uint64 `json:"tx_fifo_errors"`

	// TXCollisions is the cumulative transmit collision count.
	TXCollisions uint64 `json:"tx_collisions"`

	// TXCarrierErrors is the cumulative transmit carrier error count.
	TXCarrierErrors uint64 `json:"tx_carrier_errors"`

	// TXCompressed is the cumulative transmitted compressed-packet count.
	TXCompressed uint64 `json:"tx_compressed"`

	// TXAbortedErrors is the cumulative transmit aborted error count.
	TXAbortedErrors uint64 `json:"tx_aborted_errors"`

	// TXHeartbeatErrors is the cumulative transmit heartbeat error count.
	TXHeartbeatErrors uint64 `json:"tx_heartbeat_errors"`

	// TXWindowErrors is the cumulative transmit window error count.
	TXWindowErrors uint64 `json:"tx_window_errors"`
}

// NetworkProtocolSnapshot groups raw protocol counters by kernel source group.
type NetworkProtocolSnapshot struct {
	// IP stores counters from the IP group.
	IP NetworkCounterGroup `json:"ip"`

	// ICMP stores counters from the ICMP group.
	ICMP NetworkCounterGroup `json:"icmp"`

	// TCP stores counters from the TCP group.
	TCP NetworkCounterGroup `json:"tcp"`

	// UDP stores counters from the UDP group.
	UDP NetworkCounterGroup `json:"udp"`

	// UDPLite stores counters from the UDPLite group.
	UDPLite NetworkCounterGroup `json:"udplite"`

	// IPExt stores counters from the IP extension group.
	IPExt NetworkCounterGroup `json:"ip_ext"`

	// TCPExt stores counters from the TCP extension group.
	TCPExt NetworkCounterGroup `json:"tcp_ext"`
}

// NetworkSocketSnapshot contains summarized socket metrics for one snapshot.
type NetworkSocketSnapshot struct {
	// Total stores aggregate socket totals by family or protocol bucket.
	Total NetworkSocketTotals `json:"total"`

	// ByState stores counts grouped by socket state name.
	ByState map[string]uint64 `json:"by_state"`

	// TopLocalPorts stores the highest observed local port counts.
	TopLocalPorts []NetworkSocketPortCount `json:"top_local_ports"`

	// TopRemoteIPs stores the highest observed remote IP counts.
	TopRemoteIPs []NetworkSocketIPCount `json:"top_remote_ips"`

	// SockStat stores factual totals derived from /proc/net/sockstat.
	SockStat NetworkSockStat `json:"sockstat"`
}

// NetworkSocketTotals stores aggregate socket totals.
type NetworkSocketTotals struct {
	// All is the total number of summarized sockets.
	All uint64 `json:"all"`

	// TCP is the summarized IPv4 TCP socket count.
	TCP uint64 `json:"tcp"`

	// TCP6 is the summarized IPv6 TCP socket count.
	TCP6 uint64 `json:"tcp6"`

	// UDP is the summarized IPv4 UDP socket count.
	UDP uint64 `json:"udp"`

	// UDP6 is the summarized IPv6 UDP socket count.
	UDP6 uint64 `json:"udp6"`
}

// NetworkSocketPortCount stores one summarized local-port count item.
type NetworkSocketPortCount struct {
	// Protocol is the socket protocol name for the port count.
	Protocol string `json:"protocol"`

	// Port is the local port number.
	Port uint16 `json:"port"`

	// Count is the number of summarized sockets for the port.
	Count uint64 `json:"count"`
}

// NetworkSocketIPCount stores one summarized remote-IP count item.
type NetworkSocketIPCount struct {
	// IP is the remote IP address string.
	IP string `json:"ip"`

	// Count is the number of summarized sockets for the IP.
	Count uint64 `json:"count"`
}

// NetworkSockStat stores factual totals from /proc/net/sockstat.
type NetworkSockStat struct {
	// SocketsUsed is the total sockets-used value.
	SocketsUsed uint64 `json:"sockets_used"`

	// TCPInUse is the tcp inuse value.
	TCPInUse uint64 `json:"tcp_inuse"`

	// TCPOrphan is the tcp orphan value.
	TCPOrphan uint64 `json:"tcp_orphan"`

	// TCPTimeWait is the tcp tw value.
	TCPTimeWait uint64 `json:"tcp_time_wait"`

	// TCPAlloc is the tcp alloc value.
	TCPAlloc uint64 `json:"tcp_alloc"`

	// TCPMem is the tcp mem value.
	TCPMem uint64 `json:"tcp_mem"`

	// UDPInUse is the udp inuse value.
	UDPInUse uint64 `json:"udp_inuse"`

	// UDPMem is the udp mem value.
	UDPMem uint64 `json:"udp_mem"`

	// UDPLiteInUse is the udplite inuse value.
	UDPLiteInUse uint64 `json:"udplite_inuse"`

	// RAWInUse is the raw inuse value.
	RAWInUse uint64 `json:"raw_inuse"`

	// FragInUse is the frag inuse value.
	FragInUse uint64 `json:"frag_inuse"`

	// FragMemory is the frag memory value.
	FragMemory uint64 `json:"frag_memory"`
}

// NetworkKernelPressureSnapshot stores kernel-side pressure indicators.
type NetworkKernelPressureSnapshot struct {
	// SoftIRQs stores network-related softirq counters.
	SoftIRQs NetworkSoftIRQSnapshot `json:"softirqs"`
}

// NetworkSoftIRQSnapshot stores summarized network softirq counters.
type NetworkSoftIRQSnapshot struct {
	// NETRXTotal is the total NET_RX softirq count.
	NETRXTotal uint64 `json:"net_rx_total"`

	// NETTXTotal is the total NET_TX softirq count.
	NETTXTotal uint64 `json:"net_tx_total"`

	// NETRXPerCPU stores NET_RX counters in CPU index order.
	NETRXPerCPU []uint64 `json:"net_rx_per_cpu"`

	// NETTXPerCPU stores NET_TX counters in CPU index order.
	NETTXPerCPU []uint64 `json:"net_tx_per_cpu"`
}
