import type {
  DiskMetricDeviceIOCountersDTO,
  DiskMetricDeviceSnapshotDTO,
  DiskMetricMountSnapshotDTO,
  DiskMetricSnapshotDTO,
} from "@dto/monitoring/disk-metric";
import type {
  NetworkMetricInterfaceSnapshotDTO,
  NetworkMetricSnapshotDTO,
} from "@dto/monitoring/network-metric";
import { SystemTelemetryPage } from "@pages/SystemTelemetryPage/SystemTelemetryPage";
import { fireEvent, render, screen, within } from "@testing-library/react";
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
  buildNetworkSnapshot({
    interfaces: [
      buildEthernetInterface({
        statistics: buildNetworkStatistics({
          rx_bytes: 100,
          rx_packets: 2,
          tx_bytes: 150,
          tx_packets: 4,
        }),
      }),
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
      sockstat: buildSockstat({
        sockets_used: 6,
        tcp_alloc: 2,
        tcp_inuse: 2,
        tcp_mem: 1,
        udp_inuse: 1,
        udp_mem: 1,
      }),
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
  }),
  buildNetworkSnapshot({
    interfaces: [
      buildEthernetInterface({
        statistics: buildNetworkStatistics({
          rx_bytes: 140,
          rx_packets: 3,
          tx_bytes: 190,
          tx_packets: 5,
        }),
      }),
      buildEthernetInterface({
        address: "66:77:88:99:AA:BB",
        name: "eth1",
        oper_state: "down",
        statistics: buildNetworkStatistics({
          rx_bytes: 60,
          rx_packets: 1,
          tx_bytes: 80,
          tx_packets: 2,
        }),
      }),
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
      sockstat: buildSockstat({
        sockets_used: 8,
        tcp_alloc: 3,
        tcp_inuse: 3,
        tcp_mem: 1,
        udp_inuse: 2,
        udp_mem: 1,
      }),
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
  }),
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

const diskMetricItems: DiskMetricSnapshotDTO[] = [
  buildDiskSnapshot({
    devices: [
      buildDiskDevice({
        io: buildDiskIoCounters({
          io_in_progress: 1,
          io_time_ms: 200,
          read_time_ms: 30,
          reads_completed: 120,
          sectors_read: 4000,
          sectors_written: 8000,
          weighted_io_time_ms: 200,
          write_time_ms: 20,
          writes_completed: 180,
        }),
      }),
      buildDiskLoopDevice({ mounts: ["/snap/core"] }),
      buildDiskPartitionedLoopDevice({
        io: buildDiskIoCounters({
          io_in_progress: 1,
          io_time_ms: 80,
          read_time_ms: 10,
          reads_completed: 24,
          sectors_read: 1200,
          sectors_written: 2400,
          weighted_io_time_ms: 80,
          write_time_ms: 14,
          writes_completed: 40,
        }),
      }),
      buildDiskPartition({}),
    ],
    mounts: {
      "/": buildDiskMount({
        available_bytes: 10 * 1024 ** 3,
        device: "sda2",
        mountpoint: "/",
        source: "/dev/sda2",
        total_bytes: 64 * 1024 ** 3,
        used_bytes: 52 * 1024 ** 3,
        used_percent: 81.25,
      }),
      "/snap/core": buildDiskMount({
        device: "loop0",
        filesystem: "squashfs",
        mountpoint: "/snap/core",
        readonly: true,
        source: "/dev/loop0",
        total_bytes: 1000,
        used_bytes: 1000,
        used_percent: 100,
      }),
      "/mnt/archive": buildDiskMount({
        available_bytes: 160 * 1024 ** 2,
        device: "loop7p1",
        mountpoint: "/mnt/archive",
        source: "/dev/loop7p1",
        total_bytes: 256 * 1024 ** 2,
        used_bytes: 96 * 1024 ** 2,
        used_percent: 37.5,
      }),
      "/testpool": buildDiskMount({
        available_bytes: 1.5 * 1024 ** 3,
        device: null,
        filesystem: "zfs",
        mountpoint: "/testpool",
        source: "testpool",
        total_bytes: 2 * 1024 ** 3,
        used_bytes: 0.5 * 1024 ** 3,
        used_percent: 25,
      }),
    },
    timestamp: "2026-06-16T10:57:17.000Z",
  }),
  buildDiskSnapshot({
    devices: [
      buildDiskDevice({
        io: buildDiskIoCounters({
          io_in_progress: 2,
          io_time_ms: 450,
          read_time_ms: 40,
          reads_completed: 180,
          sectors_read: 5200,
          sectors_written: 9800,
          weighted_io_time_ms: 450,
          write_time_ms: 28,
          writes_completed: 260,
        }),
      }),
      buildDiskLoopDevice({ mounts: ["/snap/core"] }),
      buildDiskPartitionedLoopDevice({
        io: buildDiskIoCounters({
          io_in_progress: 0,
          io_time_ms: 160,
          read_time_ms: 14,
          reads_completed: 54,
          sectors_read: 1800,
          sectors_written: 3000,
          weighted_io_time_ms: 160,
          write_time_ms: 18,
          writes_completed: 70,
        }),
      }),
      buildDiskPartition({
        io: buildDiskIoCounters({
          io_time_ms: 40,
          reads_completed: 40,
          sectors_read: 1200,
          sectors_written: 2400,
          weighted_io_time_ms: 40,
          writes_completed: 50,
          write_time_ms: 12,
        }),
      }),
    ],
    mounts: {
      "/": buildDiskMount({
        available_bytes: 9 * 1024 ** 3,
        device: "sda2",
        mountpoint: "/",
        source: "/dev/sda2",
        total_bytes: 64 * 1024 ** 3,
        used_bytes: 53 * 1024 ** 3,
        used_percent: 82.8,
      }),
      "/snap/core": buildDiskMount({
        device: "loop0",
        filesystem: "squashfs",
        mountpoint: "/snap/core",
        readonly: true,
        source: "/dev/loop0",
        total_bytes: 1000,
        used_bytes: 1000,
        used_percent: 100,
      }),
      "/mnt/archive": buildDiskMount({
        available_bytes: 128 * 1024 ** 2,
        device: "loop7p1",
        mountpoint: "/mnt/archive",
        source: "/dev/loop7p1",
        total_bytes: 256 * 1024 ** 2,
        used_bytes: 128 * 1024 ** 2,
        used_percent: 50,
      }),
      "/testpool": buildDiskMount({
        available_bytes: 1.4 * 1024 ** 3,
        device: null,
        filesystem: "zfs",
        mountpoint: "/testpool",
        source: "testpool",
        total_bytes: 2 * 1024 ** 3,
        used_bytes: 0.6 * 1024 ** 3,
        used_percent: 30,
      }),
    },
    timestamp: "2026-06-16T10:57:18.000Z",
  }),
];

const diskMetricHookValue = {
  error: null,
  isError: false,
  isFetching: false,
  isLoading: false,
  items: diskMetricItems,
  latestItem: diskMetricItems[1],
  mode: "history",
  refetch: vi.fn(),
};

const serviceMetricServices = [
  {
    active_state: "active",
    enabled_state: "enabled",
    manager: "systemd",
    name: "lite-nas-service-metrics.service",
    sub_state: "running",
    unit_type: "service",
  },
  {
    active_state: "inactive",
    enabled_state: "disabled",
    manager: "systemd",
    name: "ssh.service",
    sub_state: "dead",
    unit_type: "service",
  },
];

const serviceMetricHookValue = {
  activeStateFilter: "",
  availableActiveStates: ["active", "inactive"],
  availableEnabledStates: ["disabled", "enabled"],
  clearFilters: vi.fn(),
  enabledStateFilter: "",
  error: null,
  getServiceByName: vi.fn((serviceName: string) => {
    return serviceMetricServices.find((service) => service.name === serviceName) ?? null;
  }),
  hasNextPage: false,
  hasPreviousPage: false,
  isError: false,
  isFetching: false,
  isLoading: false,
  nextPage: vi.fn(),
  page: 1,
  pageSize: 12,
  previousPage: vi.fn(),
  refetch: vi.fn(),
  resetPage: vi.fn(),
  resetPagination: vi.fn(),
  restart: vi.fn(),
  search: "",
  services: serviceMetricServices,
  setActiveStateFilter: vi.fn(),
  setEnabledStateFilter: vi.fn(),
  setPage: vi.fn(),
  setPageSize: vi.fn(),
  setSearch: vi.fn(),
  snapshot: {
    services: serviceMetricServices,
    timestamp: "2026-06-26T04:37:21.316499269+02:00",
  },
  toggleStartupBehaviour: vi.fn(),
  toggleState: vi.fn(),
  totalPages: 1,
  totalServices: 2,
  visibleServicesCount: 2,
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

vi.mock("@providers/ServiceMetricProvider", () => ({
  ServiceMetricProvider: ({ children }: PropsWithChildren): ReactElement => (
    <div data-testid="service-metric-provider">{children}</div>
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

vi.mock("@hooks/useDiskMetric", () => ({
  useDiskMetric: vi.fn(() => diskMetricHookValue),
}));

vi.mock("@hooks/useZFSMetric", () => ({
  useZFSMetric: vi.fn(() => zfsMetricHookValue),
}));

vi.mock("@hooks/useServiceMetric", () => ({
  useServiceMetric: vi.fn(() => serviceMetricHookValue),
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
  expect(screen.getByTestId("zfs-metric-search-control")).toBeInTheDocument();
  expect(screen.getByText("tank")).toBeInTheDocument();
  expect(screen.getByText("Online")).toBeInTheDocument();
  expect(screen.getByText("Used 60%")).toBeInTheDocument();
  expect(screen.getAllByTestId("value-line-chart")).toHaveLength(3);
  expect(screen.getByText("Errors: No known data errors")).toBeInTheDocument();

  fireEvent.change(within(screen.getByTestId("zfs-metric-search-control")).getByRole("textbox"), {
    target: { value: "missing" },
  });

  expect(screen.getByTestId("system-telemetry-zfs-search-empty")).toBeInTheDocument();
  expect(screen.queryByText("tank")).not.toBeInTheDocument();
});

test("renders the first draft of the network telemetry panels on the network performance route", () => {
  renderSystemTelemetryPage("/system/performance/network", "/system/performance/:category");

  assertNetworkTelemetryPage();

  fireEvent.change(
    within(screen.getByTestId("network-metric-search-control")).getByRole("textbox"),
    {
      target: { value: "eth1" },
    },
  );

  expect(screen.queryByText("eth0")).not.toBeInTheDocument();
  expect(screen.getByText("eth1")).toBeInTheDocument();
});

test("renders the first draft of the disk telemetry panels on the disk performance route", () => {
  renderSystemTelemetryPage("/system/performance/disk", "/system/performance/:category");

  assertDiskTelemetryPage();

  fireEvent.change(within(screen.getByTestId("disk-metric-search-control")).getByRole("textbox"), {
    target: { value: "loop0" },
  });

  expect(screen.getByText("loop0")).toBeInTheDocument();
  expect(screen.queryByText("sda | ATA | Ubuntu Linux 24.")).not.toBeInTheDocument();
});

test("renders a placeholder state for unsupported telemetry routes", () => {
  renderSystemTelemetryPage("/system/sensors/temperature", "/system/sensors/:category");

  expect(screen.getByTestId("system-telemetry-overline")).toHaveTextContent("Sensors");
  expect(screen.getByTestId("system-telemetry-title")).toHaveTextContent("Temperature");
  expect(screen.getByTestId("system-telemetry-placeholder-title")).toHaveTextContent(
    "Route pending backend support",
  );
});

test("keeps the processes route placeholder for categories without a backend contract", () => {
  renderSystemTelemetryPage("/system/processes/processes", "/system/processes/:category");

  expect(screen.getByTestId("system-telemetry-overline")).toHaveTextContent("Processes");
  expect(screen.getByTestId("system-telemetry-title")).toHaveTextContent("Processes");
  expect(screen.getByTestId("system-telemetry-placeholder-title")).toHaveTextContent(
    "Route pending backend support",
  );
});

test("renders services telemetry on the processes services route", () => {
  renderSystemTelemetryPage("/system/processes/services", "/system/processes/:category");

  expect(screen.getByTestId("monitoring-settings-provider-service-metrics")).toBeInTheDocument();
  expect(screen.getByTestId("service-metric-provider")).toBeInTheDocument();
  expect(screen.getByTestId("system-telemetry-overline")).toHaveTextContent("Processes");
  expect(screen.getByTestId("system-telemetry-title")).toHaveTextContent("Services");
  expect(screen.getByTestId("service-metric-state-card")).toBeInTheDocument();
  expect(screen.getByTestId("service-metric-search-control")).toBeInTheDocument();
  expect(screen.getByTestId("service-metric-active-state-filter")).toBeInTheDocument();
  expect(screen.getByTestId("service-metric-enabled-state-filter")).toBeInTheDocument();
  expect(screen.getByTestId("service-metric-top-pagination-control")).toBeInTheDocument();
  expect(screen.getByTestId("service-metric-bottom-pagination-control")).toBeInTheDocument();
  expect(screen.getByTestId("service-metric-total-services")).toHaveTextContent(
    "2 of 2 services match",
  );
  expect(
    screen.getByTestId("service-metric-card-title-lite-nas-service-metrics.service"),
  ).toHaveTextContent("lite-nas-service-metrics.service");
  expect(
    screen.getByTestId("service-metric-status-lite-nas-service-metrics.service"),
  ).toHaveTextContent("Status: active");
  expect(
    screen.getByTestId("service-metric-substatus-lite-nas-service-metrics.service"),
  ).toHaveTextContent("Substatus: running");
  expect(
    screen.getByTestId("service-metric-toggle-state-lite-nas-service-metrics.service"),
  ).toHaveTextContent("Stop");
  expect(
    screen.getByTestId("service-metric-toggle-startup-lite-nas-service-metrics.service"),
  ).toHaveTextContent("Disable");
  expect(screen.getByText("Startup: enabled")).toBeInTheDocument();
  expect(screen.getAllByText("Manager: systemd")).toHaveLength(2);
});

/**
 * Asserts the network telemetry route content rendered by the mocked provider data.
 */
const assertNetworkTelemetryPage = (): void => {
  expect(screen.getByTestId("monitoring-settings-provider-network-metrics")).toBeInTheDocument();
  expect(screen.getByTestId("network-metric-provider")).toBeInTheDocument();
  expect(screen.getByTestId("system-telemetry-title")).toHaveTextContent("Network");
  expect(screen.getByTestId("network-telemetry-sections")).toBeInTheDocument();
  expect(screen.getByTestId("network-metric-search-control")).toBeInTheDocument();
  expect(screen.getByTestId("network-interfaces-panel-title")).toHaveTextContent("Interfaces");
  expect(screen.getByTestId("network-interfaces-subrow")).toBeInTheDocument();
  expect(screen.getByText("eth0")).toBeInTheDocument();
  expect(screen.getAllByText("Intel I219-LM")).toHaveLength(2);
  expect(screen.getAllByText("ethernet | 1,000 Mbps | full | MTU 1,500")).toHaveLength(2);
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
 * Asserts the disk telemetry route content rendered by the mocked provider data.
 */
const assertDiskTelemetryPage = (): void => {
  expect(screen.getByTestId("monitoring-settings-provider-disk-metrics")).toBeInTheDocument();
  expect(screen.getByTestId("disk-metric-provider")).toBeInTheDocument();
  expect(screen.getByTestId("system-telemetry-title")).toHaveTextContent("Disk");
  expect(screen.getByTestId("disk-telemetry-sections")).toBeInTheDocument();
  expect(screen.getByTestId("disk-metric-search-control")).toBeInTheDocument();
  expect(screen.getByTestId("disk-devices-panel-title")).toHaveTextContent("Devices");
  expect(screen.getByText("Ubuntu Linux 24.")).toBeInTheDocument();
  expect(screen.getByText("sda | ATA | Ubuntu Linux 24.")).toBeInTheDocument();
  expect(screen.getByText("SATA | disk | solid-state")).toBeInTheDocument();
  expect(screen.getByText("Partitions: sda1, sda2")).toBeInTheDocument();
  expect(screen.getByTestId("disk-devices-panel")).toHaveTextContent("loop7");
  expect(screen.getByText("Partitions: loop7p1")).toBeInTheDocument();
  expect(screen.getByTestId("disk-auxiliary-devices-panel-title")).toHaveTextContent(
    "Loop & other block devices",
  );
  expect(screen.getByText("loop0")).toBeInTheDocument();
  expect(screen.getByText("snap image")).toBeInTheDocument();
  expect(screen.getByTestId("disk-other-block-devices-section")).toHaveTextContent(
    "No secondary block devices observed.",
  );
  expect(screen.getByTestId("disk-mounts-panel-title")).toHaveTextContent("Mounts & filesystems");
  expect(screen.getByTestId("disk-mounts-table")).toHaveTextContent("sda");
  expect(screen.queryByText("squashfs")).not.toBeInTheDocument();
  expect(screen.getByText("/testpool")).toBeInTheDocument();
  expect(screen.getAllByTestId("value-line-chart")).toHaveLength(4);
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

/**
 * Builds one complete network snapshot fixture around the shared defaults.
 */
function buildNetworkSnapshot(
  overrides: Partial<NetworkMetricSnapshotDTO>,
): NetworkMetricSnapshotDTO {
  return {
    interfaces: [],
    kernel_pressure: {
      softirqs: {
        net_rx_per_cpu: [],
        net_rx_total: 0,
        net_tx_per_cpu: [],
        net_tx_total: 0,
      },
    },
    protocols: {
      icmp: {},
      ip: {},
      ip_ext: {},
      tcp: {},
      tcp_ext: {},
      udp: {},
      udplite: {},
    },
    sockets: {
      by_state: {},
      sockstat: buildSockstat({}),
      top_local_ports: [],
      top_remote_ips: [],
      total: { all: 0, tcp: 0, tcp6: 0, udp: 0, udp6: 0 },
    },
    timestamp: "2026-06-15T13:00:00Z",
    ...overrides,
  };
}

/**
 * Builds one reusable ethernet interface fixture for telemetry page tests.
 */
function buildEthernetInterface(
  overrides: Partial<NetworkMetricInterfaceSnapshotDTO>,
): NetworkMetricInterfaceSnapshotDTO {
  return {
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
    statistics: buildNetworkStatistics({}),
    ...overrides,
  };
}

/**
 * Builds one network statistics block with zero defaults for concise interface fixtures.
 */
function buildNetworkStatistics(
  overrides: Partial<NetworkMetricInterfaceSnapshotDTO["statistics"]>,
): NetworkMetricInterfaceSnapshotDTO["statistics"] {
  return {
    rx_bytes: 0,
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
    rx_packets: 0,
    tx_aborted_errors: 0,
    tx_bytes: 0,
    tx_carrier_errors: 0,
    tx_collisions: 0,
    tx_compressed: 0,
    tx_dropped: 0,
    tx_errors: 0,
    tx_fifo_errors: 0,
    tx_heartbeat_errors: 0,
    tx_packets: 0,
    tx_window_errors: 0,
    ...overrides,
  };
}

/**
 * Builds one shared socket-stat fixture with zero defaults.
 */
function buildSockstat(overrides: Partial<NetworkMetricSnapshotDTO["sockets"]["sockstat"]>) {
  return {
    frag_inuse: 0,
    frag_memory: 0,
    raw_inuse: 0,
    sockets_used: 0,
    tcp_alloc: 0,
    tcp_inuse: 0,
    tcp_mem: 0,
    tcp_orphan: 0,
    tcp_time_wait: 0,
    udp_inuse: 0,
    udp_mem: 0,
    udplite_inuse: 0,
    ...overrides,
  };
}

/**
 * Builds one complete disk snapshot fixture around the shared defaults.
 */
function buildDiskSnapshot(overrides: Partial<DiskMetricSnapshotDTO>): DiskMetricSnapshotDTO {
  return {
    devices: [],
    filesystems: {
      ext4: { mount_count: 2, type: "local_block" },
      tmpfs: { mount_count: 3, type: "memory" },
      zfs: { mount_count: 1, type: "local_block" },
    },
    mounts: {},
    timestamp: "2026-06-16T10:57:17.000Z",
    ...overrides,
  };
}

/**
 * Builds one reusable disk device fixture for telemetry page tests.
 */
function buildDiskDevice(
  overrides: Partial<DiskMetricDeviceSnapshotDTO>,
): DiskMetricDeviceSnapshotDTO {
  return {
    connection_type: "SATA",
    io: buildDiskIoCounters({}),
    kind: "disk",
    major: 8,
    minor: 0,
    model: "Ubuntu Linux 24.",
    name: "Ubuntu Linux 24.",
    node: "sda",
    partitions: ["sda1", "sda2"],
    removable: false,
    rotational: false,
    size_bytes: 64 * 1024 ** 3,
    vendor: "ATA",
    ...overrides,
  };
}

/**
 * Builds one reusable loop-backed block-device fixture for telemetry page tests.
 */
function buildDiskLoopDevice(
  overrides: Partial<DiskMetricDeviceSnapshotDTO>,
): DiskMetricDeviceSnapshotDTO {
  return {
    connection_type: "LOOP",
    io: buildDiskIoCounters({
      io_time_ms: 20,
      reads_completed: 10,
      read_time_ms: 5,
      sectors_read: 100,
      sectors_written: 0,
      weighted_io_time_ms: 20,
      writes_completed: 0,
      write_time_ms: 0,
    }),
    kind: "loop",
    major: 7,
    minor: 0,
    node: "loop0",
    rotational: false,
    size_bytes: 1024,
    ...overrides,
  };
}

/**
 * Builds one reusable partition-container loop fixture for telemetry page tests.
 */
function buildDiskPartitionedLoopDevice(
  overrides: Partial<DiskMetricDeviceSnapshotDTO>,
): DiskMetricDeviceSnapshotDTO {
  return {
    connection_type: "LOOP",
    io: buildDiskIoCounters({
      io_time_ms: 80,
      reads_completed: 24,
      read_time_ms: 10,
      sectors_read: 1200,
      sectors_written: 2400,
      weighted_io_time_ms: 80,
      writes_completed: 40,
      write_time_ms: 14,
    }),
    kind: "loop",
    major: 7,
    minor: 7,
    node: "loop7",
    partitions: ["loop7p1"],
    rotational: false,
    size_bytes: 256 * 1024 ** 2,
    ...overrides,
  };
}

/**
 * Builds one reusable disk partition fixture for telemetry page tests.
 */
function buildDiskPartition(
  overrides: Partial<DiskMetricDeviceSnapshotDTO>,
): DiskMetricDeviceSnapshotDTO {
  return {
    connection_type: "SATA",
    io: buildDiskIoCounters({
      io_in_progress: 0,
      io_time_ms: 20,
      read_time_ms: 10,
      reads_completed: 30,
      sectors_read: 1000,
      sectors_written: 2000,
      weighted_io_time_ms: 20,
      write_time_ms: 10,
      writes_completed: 40,
    }),
    kind: "partition",
    major: 8,
    minor: 2,
    mounts: ["/", "/tmp"],
    node: "sda2",
    parent: "sda",
    size_bytes: 63 * 1024 ** 3,
    ...overrides,
  };
}

/**
 * Builds one disk I/O counter block with stable defaults for telemetry fixtures.
 */
function buildDiskIoCounters(
  overrides: Partial<DiskMetricDeviceIOCountersDTO>,
): DiskMetricDeviceIOCountersDTO {
  return {
    io_in_progress: 0,
    io_time_ms: 200,
    reads_completed: 120,
    reads_merged: 0,
    read_time_ms: 30,
    sectors_read: 4000,
    sectors_written: 8000,
    weighted_io_time_ms: 200,
    writes_completed: 180,
    writes_merged: 0,
    write_time_ms: 20,
    ...overrides,
  };
}

/**
 * Builds one disk mount fixture with concise defaults for telemetry page tests.
 */
function buildDiskMount(
  overrides: Partial<DiskMetricMountSnapshotDTO> &
    Pick<DiskMetricMountSnapshotDTO, "mountpoint" | "source">,
): DiskMetricMountSnapshotDTO {
  const { mountpoint, source, ...rest } = overrides;

  return {
    available_bytes: 10 * 1024 ** 3,
    device: "sda2",
    filesystem: "ext4",
    free_bytes: 12 * 1024 ** 3,
    memory_backed: false,
    mountpoint,
    readonly: false,
    remote: false,
    source,
    total_bytes: 64 * 1024 ** 3,
    used_bytes: 52 * 1024 ** 3,
    used_percent: 81.25,
    ...rest,
  };
}
