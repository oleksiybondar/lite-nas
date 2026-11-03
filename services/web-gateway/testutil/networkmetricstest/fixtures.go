package networkmetricstest

import (
	"time"

	"lite-nas/shared/metrics"
)

// Snapshot builds one representative network metrics snapshot for web-gateway tests.
func Snapshot(unixSeconds int64) metrics.NetworkMetricsSnapshot {
	return metrics.NetworkMetricsSnapshot{
		Timestamp:  time.Unix(unixSeconds, 0).UTC(),
		Interfaces: []metrics.NetworkInterfaceSnapshot{interfaceSnapshot()},
		Protocols:  protocolSnapshot(),
		Sockets:    socketSnapshot(),
	}
}

// interfaceSnapshot returns one realistic interface sample for controller and service tests.
func interfaceSnapshot() metrics.NetworkInterfaceSnapshot {
	carrierUp := true
	ifIndex := uint64(2)
	mtu := uint64(1500)
	speedMbps := uint64(1000)
	txQueueLen := uint64(1000)
	bus := "pci"
	vendorName := "Intel"
	deviceName := "I219-LM"
	description := "Ethernet Controller"

	return metrics.NetworkInterfaceSnapshot{
		Name:       "eth0",
		IfIndex:    &ifIndex,
		Address:    "00:11:22:33:44:55",
		MTU:        &mtu,
		OperState:  "up",
		CarrierUp:  &carrierUp,
		SpeedMbps:  &speedMbps,
		Duplex:     "full",
		TxQueueLen: &txQueueLen,
		Kind:       "ethernet",
		Bus:        &bus,
		Adapter: &metrics.NetworkInterfaceAdapter{
			VendorID:    "8086",
			DeviceID:    "15fb",
			VendorName:  &vendorName,
			DeviceName:  &deviceName,
			Description: &description,
		},
		Statistics: metrics.NetworkInterfaceStatistics{
			RXBytes:   100,
			RXPackets: 2,
			TXBytes:   200,
			TXPackets: 4,
		},
	}
}

// protocolSnapshot returns representative protocol counters for network metrics tests.
func protocolSnapshot() metrics.NetworkProtocolSnapshot {
	return metrics.NetworkProtocolSnapshot{
		TCP: metrics.NetworkCounterGroup{"ActiveOpens": 3},
		UDP: metrics.NetworkCounterGroup{"InDatagrams": 5},
	}
}

// socketSnapshot returns representative socket totals and rankings for network metrics tests.
func socketSnapshot() metrics.NetworkSocketSnapshot {
	return metrics.NetworkSocketSnapshot{
		Total:         metrics.NetworkSocketTotals{All: 6, TCP: 2, UDP: 1},
		ByState:       map[string]uint64{"LISTEN": 1},
		TopLocalPorts: []metrics.NetworkSocketPortCount{{Protocol: "tcp", Port: 443, Count: 1}},
		TopRemoteIPs:  []metrics.NetworkSocketIPCount{{IP: "192.0.2.10", Count: 1}},
		SockStat:      metrics.NetworkSockStat{SocketsUsed: 6, TCPInUse: 2, UDPInUse: 1},
	}
}
