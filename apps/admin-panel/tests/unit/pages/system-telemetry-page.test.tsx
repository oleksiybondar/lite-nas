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
    maxRecords: 300,
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
    items: [{ Timestamp: "2026-06-07T13:00:00Z" }],
    latestItem: { Timestamp: "2026-06-07T13:00:00Z" },
    mode: "snapshot",
    refetch: vi.fn(),
  })),
}));

describe("SystemTelemetryPage", () => {
  test("renders gateway-backed system metrics state on the system performance route", () => {
    renderSystemTelemetryPage("/system/performance/system", "/system/performance/:category");

    expect(screen.getByTestId("system-telemetry-overline")).toHaveTextContent("Performance");
    expect(screen.getByTestId("system-telemetry-title")).toHaveTextContent("System (CPU & RAM)");
    expect(screen.getByTestId("system-telemetry-total-cpu-title")).toHaveTextContent("Total CPU");
    expect(screen.getAllByTestId("percent-gradient-chart")).toHaveLength(2);
    expect(screen.getByTestId("system-telemetry-ram-title")).toHaveTextContent("RAM");
    expect(screen.getByText("Total: 1000 B")).toBeInTheDocument();
    expect(screen.getByText("Used: 420 B")).toBeInTheDocument();
    expect(screen.getByTestId("system-telemetry-per-core-cpu-title")).toHaveTextContent(
      "Per-core CPU",
    );
    expect(screen.getByTestId("percent-gradient-multi-chart")).toBeInTheDocument();
  });

  test("renders gateway-backed zfs metrics state on the zfs performance route", () => {
    renderSystemTelemetryPage("/system/performance/zfs", "/system/performance/:category");

    expect(screen.getByTestId("system-telemetry-overline")).toHaveTextContent("Performance");
    expect(screen.getByTestId("system-telemetry-title")).toHaveTextContent("Zfs");
    expect(screen.getByTestId("system-telemetry-metric-title")).toHaveTextContent("ZFS metrics");
    expect(screen.getByTestId("system-telemetry-metric-mode")).toHaveTextContent("snapshot");
    expect(screen.getByTestId("system-telemetry-metric-fetching")).toHaveTextContent("true");
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
