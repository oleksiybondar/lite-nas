import { SystemMetricsCard } from "@components/monitoring/SystemMetricsCard";
import type { SystemMetricsCardData } from "@helpers/system-metric-chart";
import { render, screen } from "@testing-library/react";

const mockMetrics: SystemMetricsCardData = {
  availableRam: "500 MiB",
  cpuUsageColor: "success.main",
  cpuUsageLabel: "25.5%",
  cpuUsagePct: 25.5,
  perCoreCpuSeries: {
    stamps: ["2026-06-08T12:00:00Z"],
    valuesByKey: {
      CPU1: [20],
      CPU2: [31],
    },
  },
  ramSeries: {
    stamps: ["2026-06-08T12:00:00Z"],
    values: [40],
  },
  ramUsageColor: "success.main",
  ramUsageLabel: "40.0%",
  ramUsagePct: 40,
  totalCores: 2,
  totalCpuSeries: {
    stamps: ["2026-06-08T12:00:00Z"],
    values: [25.5],
  },
  totalRam: "2 GiB",
  usedRam: "1.5 GiB",
};

describe("SystemMetricsCard", () => {
  test("renders all rows with correct data", () => {
    render(<SystemMetricsCard capacity={300} metrics={mockMetrics} />);

    // Row 1
    expect(screen.getByTestId("system-metrics-cpu-percent")).toHaveTextContent("CPU 25.5%");
    expect(screen.getByTestId("system-metrics-ram-percent")).toHaveTextContent("RAM 40.0%");

    // Row 2
    expect(screen.getByTestId("system-metrics-total-cores")).toHaveTextContent("Total cores: 2");

    // Row 3
    expect(screen.getByText("Total RAM: 2 GiB")).toBeInTheDocument();
    expect(screen.getByText("Used RAM: 1.5 GiB")).toBeInTheDocument();
    expect(screen.getByText("Available RAM: 500 MiB")).toBeInTheDocument();

    // Row 4
    expect(screen.getByTestId("system-metrics-total-cpu-chart-title")).toHaveTextContent(
      "Total CPU %",
    );
    expect(screen.getByTestId("system-metrics-ram-chart-title")).toHaveTextContent("RAM %");
    expect(screen.getAllByTestId("percent-gradient-chart")).toHaveLength(2);

    // Row 5
    expect(screen.getByTestId("system-metrics-per-core-chart-title")).toHaveTextContent(
      "Per-core CPU %",
    );
    expect(screen.getByTestId("percent-gradient-multi-chart")).toBeInTheDocument();
  });
});
