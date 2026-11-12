import {
  parseSensorHistory,
  readSensorCharts,
  sensorMarkersSchema,
} from "@schemas/monitoring/sensor";

test("parses a complete sensor snapshot from the history envelope", () => {
  const snapshot = {
    timestamp: "2026-11-11T12:00:00Z",
    temperatures: [{ key: "cpu", name: "CPU", value: 42, unit: "°C" }],
    voltages: [],
    fans: [],
    clocks: [],
    throttling: [],
  };
  expect(
    parseSensorHistory({ success: true, timestamp: snapshot.timestamp, data: [snapshot] }),
  ).toEqual([snapshot]);
});

test.each(["red", "#fff", "#1234567"])("rejects invalid marker color %s", (color) => {
  expect(sensorMarkersSchema.safeParse({ "voltages:core": [{ value: 5, color }] }).success).toBe(
    false,
  );
});

test("migrates legacy sensor stops without chart settings", () => {
  expect(
    readSensorCharts({ sensorMarkers: { "voltages:core": [{ value: 5, color: "#00ff00" }] } }),
  ).toEqual({
    "voltages:core": {
      mode: "gradient",
      stops: [{ value: 5, color: "#00ff00", label: "", showBar: false, showLabel: false }],
    },
  });
});

test("keeps valid legacy sensors when another entry is malformed", () => {
  expect(
    readSensorCharts({
      sensorMarkers: {
        "voltages:core": [{ value: 5, color: "#00ff00" }],
        "voltages:bad": [{ value: 5, color: "green" }],
      },
    }),
  ).toHaveProperty("voltages:core");
});

test("keeps valid legacy stops when the newer chart for the same sensor is malformed", () => {
  expect(
    readSensorCharts({
      sensorMarkers: { "voltages:core": [{ value: 5, color: "#00ff00" }] },
      sensorCharts: { "voltages:core": { mode: "gradient", stops: "invalid" } },
    })["voltages:core"]?.stops[0]?.color,
  ).toBe("#00ff00");
});
