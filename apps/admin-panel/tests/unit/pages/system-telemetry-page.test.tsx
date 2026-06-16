import type { NetworkMetricSnapshotDTO } from "@dto/monitoring/network-metric";
import { SystemTelemetryPage } from "@pages/SystemTelemetryPage/SystemTelemetryPage";
import { render, screen } from "@testing-library/react";
import { TestMemoryRouter } from "@tests/unit/test-utils/router";
import type { PropsWithChildren, ReactElement } from "react";
import { Route, Routes } from "react-router-dom";

const systemMetricHookValue = {
  error: null,
  isError: false,
  isFetching: false,
  isLoading: false,
  items: [
    {
      CPU: { PerCoreUsage: [12, 18], TotalUsagePct: 15.2 },
      Mem: { TotalBytes: 1000, UsedBytes: 400, UsedPct: 40 },
      Timestamp: "2026-06-07T12:00:00Z",
    },
    {
      CPU: { PerCoreUsage: [30, 36], TotalUsagePct: 33.5 },
      Mem: { TotalBytes: 1000, UsedBytes: 420, UsedPct: 42 },
      Timestamp: "2026-06-07T12:00:01Z",
    },
  ],
  latestItem: {
    CPU: { PerCoreUsage: [30, 36], TotalUsagePct: 33.5 },
    Mem: { TotalBytes: 1000, UsedBytes: 420, UsedPct: 42 },
    Timestamp: "2026-06-07T12:00:01Z",
  },
  mode: "history",
  refetch: vi.fn(),
};

const networkMetricItems: NetworkMetricSnapshotDTO[] = [
  {
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
        duplex: "full",
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
          tx_bytes: 150,
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
      top_local_ports: [],
      top_remote_ips: [],
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
        duplex: "full",
        kind: "ethernet",
        mtu: 1500,
        name: "eth0",
        oper_state: "up",
        speed_mbps: 1000,
        statistics: {
          rx_bytes: 140,
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
          rx_packets: 3,
          tx_aborted_errors: 0,
          tx_bytes: 190,
          tx_carrier_errors: 0,
          tx_collisions: 0,
          tx_compressed: 0,
          tx_dropped: 0,
          tx_errors: 0,
          tx_fifo_errors: 0,
          tx_heartbeat_errors: 0,
          tx_packets: 5,
          tx_window_errors: 0,
        },
      },
      {
        address: "66:77:88:99:AA:BB",
        duplex: "full",
        kind: "ethernet",
        name: "eth1",
        oper_state: "down",
        statistics: {
          rx_bytes: 60,
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
          rx_packets: 1,
          tx_aborted_errors: 0,
          tx_bytes: 80,
          tx_carrier_errors: 0,
          tx_collisions: 0,
          tx_compressed: 0,
          tx_dropped: 0,
          tx_errors: 0,
          tx_fifo_errors: 0,
          tx_heartbeat_errors: 0,
          tx_packets: 2,
          tx_window_errors: 0,
        },
      },
    ],
    kernel_pressure: {
      softirqs: {
        net_rx_per_cpu: [25, 35],
        net_rx_total: 60,
        net_tx_per_cpu: [20, 30],
        net_tx_total: 50,
      },
    },
    protocols: {
      icmp: {},
      ip: {},
      ip_ext: {},
      tcp: { ActiveOpens: 4 },
      tcp_ext: {},
      udp: { InDatagrams: 9 },
      udplite: {},
    },
    sockets: {
      by_state: { ESTABLISHED: 2, LISTEN: 1 },
      sockstat: {
        frag_inuse: 0,
        frag_memory: 0,
        raw_inuse: 0,
        sockets_used: 8,
        tcp_alloc: 3,
        tcp_inuse: 3,
        tcp_mem: 1,
        tcp_orphan: 0,
        tcp_time_wait: 0,
        udp_inuse: 2,
        udp_mem: 1,
        udplite_inuse: 0,
      },
      top_local_ports: [
        { count: 3, port: 5173, protocol: "tcp" },
        { count: 2, port: 53, protocol: "udp" },
        { count: 1, port: 34287, protocol: "tcp6" },
      ],
      top_remote_ips: [{ count: 2, ip: "127.0.0.1" }],
      total: {
        all: 8,
        tcp: 3,
        tcp6: 0,
        udp: 2,
        udp6: 0,
      },
    },
    timestamp: "2026-06-15T13:00:01Z",
  },
];

const networkMetricHookValue = {
  error: null,
  isError: false,
  isFetching: false,
  isLoading: false,
  items: networkMetricItems,
  latestItem: {
    interfaces: [{ name: "eth0" }, { name: "eth1" }],
    timestamp: "2026-06-15T13:00:01Z",
  },
  mode: "history",
  refetch: vi.fn(),
};

const zfsMetricHookValue = {
  error: null,
  isError: false,
  isFetching: true,
  isLoading: false,
  items: [
    {
      Pools: [
        {
          Errors: "none",
          Health: "ONLINE",
          IOStat: {
            Bandwidth: { Read: 10, Write: 20 },
            Operations: { Read: 3, Write: 4 },
          },
          Name: "tank",
          Root: {
            Children: null,
            Errors: { Checksum: 1, Read: 2, Write: 3 },
            Name: "root",
            Path: "/dev/root",
            Type: "disk",
          },
          Scan: "scrub repaired 0B",
          Usage: {
            AllocatedBytes: 300,
            CapacityPct: 60,
            FreeBytes: 200,
            SizeBytes: 500,
          },
        },
      ],
      Timestamp: "2026-06-07T13:00:00Z",
    },
  ],
  latestItem: { Pools: [{ Name: "tank" }], Timestamp: "2026-06-07T13:00:00Z" },
  mode: "snapshot",
  refetch: vi.fn(),
};

vi.mock("@providers/MonitoringPollingSettingsProvider", () => ({
  MonitoringPollingSettingsProvider: ({
    children,
    storageKey,
  }: PropsWithChildren<{ storageKey: string }>): ReactElement => (
    <div data-testid={`monitoring-settings-provider-${storageKey}`}>{children}</div>
  ),
}));

vi.mock("@providers/SystemMetricProvider", () => ({
  SystemMetricProvider: ({ children }: PropsWithChildren): ReactElement => (
    <div data-testid="system-metric-provider">{children}</div>
  ),
}));

vi.mock("@providers/NetworkMetricProvider", () => ({
  NetworkMetricProvider: ({ children }: PropsWithChildren): ReactElement => (
    <div data-testid="network-metric-provider">{children}</div>
  ),
}));

vi.mock("@providers/DiskMetricProvider", () => ({
  DiskMetricProvider: ({ children }: PropsWithChildren): ReactElement => (
    <div data-testid="disk-metric-provider">{children}</div>
  ),
}));

vi.mock("@providers/ZFSMetricProvider", () => ({
  ZFSMetricProvider: ({ children }: PropsWithChildren): ReactElement => (
    <div data-testid="zfs-metric-provider">{children}</div>
  ),
}));

vi.mock("@hooks/useSystemMetric", () => ({
  useSystemMetric: vi.fn(() => systemMetricHookValue),
}));

vi.mock("@hooks/useMonitoringPollingSettings", () => ({
  useMonitoringPollingSettings: vi.fn(() => ({
    historyIntervalMs: 15000,
    historyResetGapMs: 10000,
    maxRecords: 180,
    mode: "history",
    resetSettings: vi.fn(),
    setHistoryIntervalMs: vi.fn(),
    setHistoryResetGapMs: vi.fn(),
    setMaxRecords: vi.fn(),
    setMode: vi.fn(),
    setSettings: vi.fn(),
    setSnapshotIntervalMs: vi.fn(),
    snapshotIntervalMs: 1000,
  })),
}));

vi.mock("@hooks/useNetworkMetric", () => ({
  useNetworkMetric: vi.fn(() => networkMetricHookValue),
}));

vi.mock("@hooks/useZFSMetric", () => ({
  useZFSMetric: vi.fn(() => zfsMetricHookValue),
}));

test("renders gateway-backed system metrics state on the system performance route", () => {
  renderSystemTelemetryPage("/system/performance/system", "/system/performance/:category");

  expect(screen.getByTestId("system-telemetry-overline")).toHaveTextContent("Performance");
  expect(screen.getByTestId("system-telemetry-title")).toHaveTextContent("System (CPU & RAM)");
  expect(screen.getByTestId("system-metrics-card")).toBeInTheDocument();
  expect(screen.getByTestId("system-metrics-cpu-percent")).toHaveTextContent("CPU 33.5%");
  expect(screen.getByTestId("system-metrics-ram-percent")).toHaveTextContent("RAM 42.0%");
  expect(screen.getByTestId("system-metrics-total-cores")).toHaveTextContent("Total cores: 2");
  expect(screen.getByText("Total RAM: 1000 B")).toBeInTheDocument();
  expect(screen.getByText("Used RAM: 420 B")).toBeInTheDocument();
  expect(screen.getByText("Available RAM: 580 B")).toBeInTheDocument();
  expect(screen.getAllByTestId("percent-gradient-chart")).toHaveLength(2);
  expect(screen.getByTestId("percent-gradient-multi-chart")).toBeInTheDocument();
});

test("renders gateway-backed zfs metrics state on the zfs performance route", () => {
  renderSystemTelemetryPage("/system/performance/zfs", "/system/performance/:category");

  expect(screen.getByTestId("system-telemetry-overline")).toHaveTextContent("Performance");
  expect(screen.getByTestId("system-telemetry-title")).toHaveTextContent("Zfs");
  expect(screen.getByText("tank")).toBeInTheDocument();
  expect(screen.getByText("Online")).toBeInTheDocument();
  expect(screen.getByText("Used 60%")).toBeInTheDocument();
  expect(screen.getAllByTestId("value-line-chart")).toHaveLength(3);
  expect(screen.getByText("Errors: No known data errors")).toBeInTheDocument();
});

test("renders the first draft of the network telemetry panels on the network performance route", () => {
  renderSystemTelemetryPage("/system/performance/network", "/system/performance/:category");

  assertNetworkTelemetryPage();
});

test("wraps the disk performance route with the disk metrics providers", () => {
  renderSystemTelemetryPage("/system/performance/disk", "/system/performance/:category");

  expect(screen.getByTestId("monitoring-settings-provider-disk-metrics")).toBeInTheDocument();
  expect(screen.getByTestId("disk-metric-provider")).toBeInTheDocument();
  expect(screen.getByTestId("system-telemetry-title")).toHaveTextContent("Disk");
  expect(screen.getByTestId("system-telemetry-placeholder-title")).toHaveTextContent(
    "Route pending backend support",
  );
});

test("renders a placeholder state for unsupported telemetry routes", () => {
  renderSystemTelemetryPage("/system/sensors/temperature", "/system/sensors/:category");

  expect(screen.getByTestId("system-telemetry-overline")).toHaveTextContent("Sensors");
  expect(screen.getByTestId("system-telemetry-title")).toHaveTextContent("Temperature");
  expect(screen.getByTestId("system-telemetry-placeholder-title")).toHaveTextContent(
    "Route pending backend support",
  );
});

/**
 * Asserts the network telemetry route content rendered by the mocked provider data.
 */
const assertNetworkTelemetryPage = (): void => {
  expect(screen.getByTestId("monitoring-settings-provider-network-metrics")).toBeInTheDocument();
  expect(screen.getByTestId("network-metric-provider")).toBeInTheDocument();
  expect(screen.getByTestId("system-telemetry-title")).toHaveTextContent("Network");
  expect(screen.getByTestId("network-telemetry-sections")).toBeInTheDocument();
  expect(screen.getByTestId("network-interfaces-panel-title")).toHaveTextContent("Interfaces");
  expect(screen.getByTestId("network-interfaces-subrow")).toBeInTheDocument();
  expect(screen.getByText("eth0")).toBeInTheDocument();
  expect(screen.getByText("Intel I219-LM")).toBeInTheDocument();
  expect(screen.getByText("ethernet | 1,000 Mbps | full | MTU 1,500")).toBeInTheDocument();
  expect(screen.getByText("eth1")).toBeInTheDocument();
  expect(screen.getByText("up")).toBeInTheDocument();
  expect(screen.getByText("down")).toBeInTheDocument();
  expect(screen.getByText("Total RX: 140 B")).toBeInTheDocument();
  expect(screen.getByText("Total TX: 190 B")).toBeInTheDocument();
  expect(screen.getAllByTestId("value-line-chart")).toHaveLength(7);
  expect(screen.getAllByText("RX/TX packets per second")).toHaveLength(2);
  expect(screen.getAllByText("RX/TX errors + drops")).toHaveLength(2);
  expect(screen.getByTestId("network-protocols-panel-title")).toHaveTextContent(
    "Protocols & sockets",
  );
  expect(screen.getByText("Sockets used: 8")).toBeInTheDocument();
  expect(screen.getByText("Active TCP: 2")).toBeInTheDocument();
  expect(screen.getByText("Listening: 1")).toBeInTheDocument();
  expect(screen.getAllByText("Time wait: 0")).toHaveLength(2);
  expect(screen.getByTestId("network-protocols-state-row")).toBeInTheDocument();
  expect(screen.getByTestId("network-protocols-distribution-row")).toBeInTheDocument();
  expect(screen.getByText("Connection state")).toBeInTheDocument();
  expect(screen.getByText("Protocol distribution")).toBeInTheDocument();
  expect(screen.getByText("Protocol socket totals")).toBeInTheDocument();
  expect(screen.getByTestId("network-protocols-ports-table")).toBeInTheDocument();
  expect(screen.getByTestId("network-protocols-remote-ips-table")).toBeInTheDocument();
  expect(screen.getByText("Vite dev server")).toBeInTheDocument();
  expect(screen.getByText("TCP 5173")).toBeInTheDocument();
  expect(screen.getByText("DNS")).toBeInTheDocument();
  expect(screen.getByText("UDP 53")).toBeInTheDocument();
  expect(screen.getByText("Unknown app")).toBeInTheDocument();
  expect(screen.getByText("TCP6 34287")).toBeInTheDocument();
  expect(screen.getByText("127.0.0.1")).toBeInTheDocument();
};

/**
 * Renders the system telemetry page under one concrete route path.
 */
const renderSystemTelemetryPage = (initialEntry: string, routePath: string): void => {
  render(
    <TestMemoryRouter initialEntries={[initialEntry]}>
      <Routes>
        <Route element={<SystemTelemetryPage />} path={routePath} />
      </Routes>
    </TestMemoryRouter>,
  );
};
