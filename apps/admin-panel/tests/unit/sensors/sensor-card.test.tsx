import { SensorCard } from "@pages/SensorCard";
import type { SensorChart, SensorSnapshot } from "@schemas/monitoring/sensor";
import { screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { renderWithLanguage } from "../test-utils/language";

const reading = { key: "EXT5V_V", name: "5 V input supply", value: 5, unit: "V" };
const items: SensorSnapshot[] = [
  {
    timestamp: "2026-11-11T12:00:00Z",
    temperatures: [],
    voltages: [reading],
    clocks: [],
    fans: [],
    throttling: [],
  },
  {
    timestamp: "2026-11-11T12:01:00Z",
    temperatures: [],
    voltages: [reading],
    clocks: [],
    fans: [],
    throttling: [],
  },
];
const chart: SensorChart = {
  mode: "regular",
  stops: [{ value: 5, color: "#ff0000", label: "", showBar: true, showLabel: false }],
};

/** Renders the voltage card with saved regular-chart guide settings. */
const renderVoltageCard = (): void => {
  renderWithLanguage(
    <MemoryRouter>
      <SensorCard
        capacity={60}
        chart={chart}
        items={items}
        markerKey="voltages:EXT5V_V"
        onEdit={() => undefined}
        reading={reading}
        section="voltages"
      />
    </MemoryRouter>,
  );
};

test("shows the sensor chart directly on the Sensors card", () => {
  renderVoltageCard();
  expect(screen.getByTestId("value-line-chart")).toBeVisible();
});

test("draws a saved guide line on a regular sensor chart", () => {
  renderVoltageCard();
  expect(
    screen.getByTestId("value-line-chart").querySelector('line[stroke="#ff0000"]'),
  ).toBeInTheDocument();
});
