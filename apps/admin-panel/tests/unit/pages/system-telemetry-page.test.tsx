import { SystemTelemetryPage } from "@pages/SystemTelemetryPage/SystemTelemetryPage";
import { render, screen } from "@testing-library/react";
import { TestMemoryRouter } from "@tests/unit/test-utils/router";
import type { PropsWithChildren, ReactElement } from "react";
import { Route, Routes } from "react-router-dom";

vi.mock("@providers/MonitoringPollingSettingsProvider", () => ({
  MonitoringPollingSettingsProvider: ({ children }: PropsWithChildren): ReactElement => (
    <>{children}</>
  ),
}));

vi.mock("@providers/SystemMetricProvider", () => ({
  SystemMetricProvider: ({ children }: PropsWithChildren): ReactElement => <>{children}</>,
}));

vi.mock("@providers/ZFSMetricProvider", () => ({
  ZFSMetricProvider: ({ children }: PropsWithChildren): ReactElement => <>{children}</>,
}));

vi.mock("@hooks/useSystemMetric", () => ({
  useSystemMetric: vi.fn(() => ({
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
  })),
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

vi.mock("@hooks/useZFSMetric", () => ({
  useZFSMetric: vi.fn(() => ({
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
  })),
}));

describe("SystemTelemetryPage", () => {
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

  test("renders a placeholder state for unsupported telemetry routes", () => {
    renderSystemTelemetryPage("/system/sensors/temperature", "/system/sensors/:category");

    expect(screen.getByTestId("system-telemetry-overline")).toHaveTextContent("Sensors");
    expect(screen.getByTestId("system-telemetry-title")).toHaveTextContent("Temperature");
    expect(screen.getByTestId("system-telemetry-placeholder-title")).toHaveTextContent(
      "Route pending backend support",
    );
  });
});

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
