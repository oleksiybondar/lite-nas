import type { NetworkMetricSnapshotDTO } from "@dto/monitoring/network-metric";
import { buildNetworkProtocolsPanelData } from "@helpers/network-protocols-panel";

const networkMetricItems: NetworkMetricSnapshotDTO[] = [
  {
    interfaces: null,
    kernel_pressure: {
      softirqs: {
        net_rx_per_cpu: null,
        net_rx_total: 0,
        net_tx_per_cpu: null,
        net_tx_total: 0,
      },
    },
    protocols: {
      icmp: { InMsgs: 10, OutMsgs: 12 },
      ip: {},
      ip_ext: {},
      tcp: { ActiveOpens: 3 },
      tcp_ext: {},
      udp: { InDatagrams: 5 },
      udplite: {},
    },
    sockets: {
      by_state: { ESTABLISHED: 1, LISTEN: 1 },
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
      top_local_ports: [{ count: 2, port: 9090, protocol: "tcp" }],
      top_remote_ips: [{ count: 1, ip: "127.0.0.1" }],
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
  {
    interfaces: null,
    kernel_pressure: {
      softirqs: {
        net_rx_per_cpu: null,
        net_rx_total: 0,
        net_tx_per_cpu: null,
        net_tx_total: 0,
      },
    },
    protocols: {
      icmp: { InMsgs: 11, OutMsgs: 13 },
      ip: {},
      ip_ext: {},
      tcp: { ActiveOpens: 4 },
      tcp_ext: {},
      udp: { InDatagrams: 9 },
      udplite: {},
    },
    sockets: {
      by_state: { CLOSE: 7, CLOSE_WAIT: 1, ESTABLISHED: 52, LISTEN: 22, TIME_WAIT: 135 },
      sockstat: {
        frag_inuse: 0,
        frag_memory: 0,
        raw_inuse: 0,
        sockets_used: 1174,
        tcp_alloc: 74,
        tcp_inuse: 61,
        tcp_mem: 0,
        tcp_orphan: 0,
        tcp_time_wait: 135,
        udp_inuse: 7,
        udp_mem: 0,
        udplite_inuse: 0,
      },
      top_local_ports: [
        { count: 69, port: 5173, protocol: "tcp" },
        { count: 67, port: 9090, protocol: "tcp" },
        { count: 2, port: 53, protocol: "udp" },
        { count: 1, port: 34287, protocol: "tcp6" },
      ],
      top_remote_ips: [
        { count: 183, ip: "127.0.0.1" },
        { count: 1, ip: "10.211.55.1" },
      ],
      total: {
        all: 217,
        tcp: 196,
        tcp6: 12,
        udp: 7,
        udp6: 2,
      },
    },
    timestamp: "2026-06-15T13:00:01Z",
  },
];

describe("network protocols panel helpers", () => {
  test("builds browser-facing protocol rows, chart data, and detailed tables", () => {
    expect(buildNetworkProtocolsPanelData(networkMetricItems)).toEqual({
      portRows: [
        { app: "Vite dev server", connection: "TCP 5173", count: "69" },
        { app: "Metrics / admin", connection: "TCP 9090", count: "67" },
        { app: "DNS", connection: "UDP 53", count: "2" },
        { app: "Unknown app", connection: "TCP6 34287", count: "1" },
      ],
      protocolTotalsLabels: [
        { key: "All sockets", value: "217" },
        { key: "TCP", value: "196" },
        { key: "TCP6", value: "12" },
        { key: "UDP", value: "7" },
        { key: "UDP6", value: "2" },
      ],
      protocolTotalsSeries: {
        stamps: ["2026-06-15T13:00:00Z", "2026-06-15T13:00:01Z"],
        valuesByKey: {
          TCP: [2, 196],
          TCP6: [0, 12],
          UDP: [1, 7],
          UDP6: [0, 2],
        },
      },
      remoteIPRows: [
        { count: "183", ip: "127.0.0.1" },
        { count: "1", ip: "10.211.55.1" },
      ],
      socketStateLabels: [
        { key: "Established", value: "52" },
        { key: "Listen", value: "22" },
        { key: "Time wait", value: "135" },
        { key: "Close wait", value: "1" },
        { key: "Close", value: "7" },
      ],
      summaryLabels: [
        { key: "Sockets used", value: "1,174" },
        { key: "Active TCP", value: "52" },
        { key: "Listening", value: "22" },
        { key: "Time wait", value: "135" },
      ],
    });
  });
});
