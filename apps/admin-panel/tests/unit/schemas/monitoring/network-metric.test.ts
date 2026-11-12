import {
  parseNetworkMetricHistoryResponse,
  parseNetworkMetricSnapshotResponse,
} from "@schemas/monitoring/network-metric";

const networkMetricSnapshotBody = {
  data: {
    interfaces: [
      {
        adapter: {
          description: "Ethernet Controller",
          device_id: "15fb",
          device_name: "I219-LM",
          vendor_id: "8086",
          vendor_name: "Intel",
        },
        address: "00:11:22:33:44:55",
        bus: "pci",
        carrier_up: true,
        duplex: "full",
        ifindex: 2,
        kind: "ethernet",
        mtu: 1500,
        name: "eth0",
        oper_state: "up",
        speed_mbps: 1000,
        statistics: {
          rx_bytes: 100,
          rx_compressed: 0,
          rx_crc_errors: 0,
          rx_dropped: 0,
          rx_errors: 0,
          rx_fifo_errors: 0,
          rx_frame_errors: 0,
          rx_length_errors: 0,
          rx_missed_errors: 0,
          rx_multicast: 0,
          rx_nohandler: 0,
          rx_over_errors: 0,
          rx_packets: 2,
          tx_aborted_errors: 0,
          tx_bytes: 200,
          tx_carrier_errors: 0,
          tx_collisions: 0,
          tx_compressed: 0,
          tx_dropped: 0,
          tx_errors: 0,
          tx_fifo_errors: 0,
          tx_heartbeat_errors: 0,
          tx_packets: 4,
          tx_window_errors: 0,
        },
        tx_queue_len: 1000,
      },
    ],
    kernel_pressure: {
      softirqs: {
        net_rx_per_cpu: [10, 20],
        net_rx_total: 30,
        net_tx_per_cpu: [5, 15],
        net_tx_total: 20,
      },
    },
    protocols: {
      icmp: {},
      ip: {},
      ip_ext: {},
      tcp: { ActiveOpens: 3 },
      tcp_ext: {},
      udp: { InDatagrams: 5 },
      udplite: {},
    },
    sockets: {
      by_state: { LISTEN: 1 },
      sockstat: {
        frag_inuse: 0,
        frag_memory: 0,
        raw_inuse: 0,
        sockets_used: 6,
        tcp_alloc: 2,
        tcp_inuse: 2,
        tcp_mem: 1,
        tcp_orphan: 0,
        tcp_time_wait: 0,
        udp_inuse: 1,
        udp_mem: 1,
        udplite_inuse: 0,
      },
      top_local_ports: [{ count: 1, port: 443, protocol: "tcp" }],
      top_remote_ips: [{ count: 1, ip: "192.0.2.10" }],
      total: {
        all: 6,
        tcp: 2,
        tcp6: 0,
        udp: 1,
        udp6: 0,
      },
    },
    timestamp: "2026-06-15T13:00:00Z",
  },
  success: true,
  timestamp: "2026-06-15T13:00:01Z",
};

describe("network metrics history schemas", () => {
  test("parses history envelopes and falls back null history data to an empty list", () => {
    expect(
      parseNetworkMetricHistoryResponse({
        data: null,
        success: true,
        timestamp: "2026-06-15T13:00:00Z",
      }),
    ).toEqual([]);
  });
});

describe("network metrics snapshot schemas", () => {
  test("parses snapshot envelopes with interface, socket, and kernel pressure sections", () => {
    expect(parseNetworkMetricSnapshotResponse(networkMetricSnapshotBody)).toHaveProperty(
      "kernel_pressure.softirqs.net_rx_total",
      30,
    );
  });
});
