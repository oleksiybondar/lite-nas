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
