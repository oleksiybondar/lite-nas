import { PercentGradientChart } from "@components/monitoring/PercentGradientChart";
import { PercentGradientMultiChart } from "@components/monitoring/PercentGradientMultiChart";
import { ValueLineChart } from "@components/monitoring/ValueLineChart";
import { fireEvent, render, screen, within } from "@testing-library/react";

const chartRect = {
  bottom: 220,
  height: 220,
  left: 0,
  right: 720,
  toJSON: () => ({}),
  top: 0,
  width: 720,
  x: 0,
  y: 0,
} as DOMRect;

/**
 * Shared test helper that simulates a hover event and verifies the tooltip content.
 */
const verifyChartTooltip = (testId: string, expectedLabels: string[]) => {
  const chart = screen.getByTestId(testId);

  vi.spyOn(chart, "getBoundingClientRect").mockReturnValue(chartRect);

  fireEvent.mouseMove(chart, { clientX: 700 });

  const tooltip = screen.getByTestId("monitoring-chart-tooltip");

  expect(tooltip).toBeInTheDocument();
  expect(within(tooltip).getByText("t2")).toBeInTheDocument();

  for (const label of expectedLabels) {
    expect(within(tooltip).getByText(label)).toBeInTheDocument();
  }

  return chart;
};

afterEach(() => {
  vi.restoreAllMocks();
});

describe("single-series chart hover tooltips", () => {
  test("renders a hover tooltip for the single-series percent chart", () => {
    render(<PercentGradientChart capacity={2} stamps={["t1", "t2"]} values={[20, 40]} />);

    const chart = verifyChartTooltip("percent-gradient-chart", ["Value: 40%"]);

    fireEvent.mouseLeave(chart);

    expect(screen.queryByTestId("monitoring-chart-tooltip")).toBeNull();
  });
});

describe("multi-series chart hover tooltips", () => {
  test("renders a hover tooltip for the multi-series percent chart", () => {
    render(
      <PercentGradientMultiChart
        capacity={2}
        stamps={["t1", "t2"]}
        valuesByKey={{ CPU1: [10, 30], CPU2: [20, 40] }}
      />,
    );

    verifyChartTooltip("percent-gradient-multi-chart", ["CPU1: 30%", "CPU2: 40%"]);
  });
});

describe("value chart hover tooltips", () => {
  test("renders a hover tooltip for the dynamic value chart", () => {
    render(
      <ValueLineChart
        capacity={2}
        formatValue={(value) => `${value} u`}
        stamps={["t1", "t2"]}
        valuesByKey={{ Read: [100, 2000], Write: [200, 4000] }}
      />,
    );

    verifyChartTooltip("value-line-chart", ["Read: 2000 u", "Write: 4000 u"]);
  });
});
