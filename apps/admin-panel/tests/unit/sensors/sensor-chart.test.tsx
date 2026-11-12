import { ValueLineChart } from "@components/monitoring/ValueLineChart";
import { render, screen } from "@testing-library/react";

test("colors the sensor line from configured value stops", () => {
  render(
    <ValueLineChart
      capacity={2}
      colorStops={[
        { value: 4.8, color: "#ff0000" },
        { value: 5, color: "#00ff00" },
      ]}
      stamps={["2026-11-11T12:00:00Z", "2026-11-11T12:01:00Z"]}
      valuesByKey={{ voltage: [4.8, 5] }}
    />,
  );
  const chart = screen.getByTestId("value-line-chart");
  expect(chart.querySelectorAll("linearGradient stop")).toHaveLength(2);
});

/** Renders a sensor chart with the guide settings saved by the chart editor. */
const renderChartWithGuides = () =>
  render(
    <ValueLineChart
      capacity={2}
      guideStops={[
        { value: 5, color: "#ff0000", label: "Maximum", showBar: true, showLabel: true },
      ]}
      stamps={["2026-11-11T12:00:00Z", "2026-11-11T12:01:00Z"]}
      valuesByKey={{ voltage: [4.8, 5] }}
    />,
  );

test("renders the edited sensor guide line", () => {
  renderChartWithGuides();
  const chart = screen.getByTestId("value-line-chart");
  expect(chart.querySelector('line[stroke="#ff0000"]')).toBeInTheDocument();
});

test("renders the edited sensor guide label", () => {
  renderChartWithGuides();
  expect(screen.getByText("Maximum")).toBeInTheDocument();
});

test("gives the sensor SVG an explicit height", () => {
  renderChartWithGuides();
  const chart = screen.getByTestId("value-line-chart");
  expect(chart.querySelector("svg")).toHaveAttribute("height", "280");
});

test("shows the sensor value as a point when only one sample is available", () => {
  render(
    <ValueLineChart
      capacity={60}
      stamps={["2026-11-11T12:00:00Z"]}
      valuesByKey={{ temperature: [47.95] }}
    />,
  );
  expect(screen.getByTestId("value-line-chart").querySelector("circle")).toBeInTheDocument();
});

test("draws a guide above a line at the same value", () => {
  render(
    <ValueLineChart
      capacity={2}
      guideStops={[{ value: 5, color: "#ff0000", label: "", showBar: true, showLabel: false }]}
      stamps={["2026-11-11T12:00:00Z", "2026-11-11T12:01:00Z"]}
      valuesByKey={{ voltage: [5, 5] }}
    />,
  );
  const chart = screen.getByTestId("value-line-chart");
  const series = chart.querySelector("path[stroke]");
  const guide = chart.querySelector('line[stroke="#ff0000"]');
  expect(
    series && guide && series.compareDocumentPosition(guide) & Node.DOCUMENT_POSITION_FOLLOWING,
  ).toBeTruthy();
});
