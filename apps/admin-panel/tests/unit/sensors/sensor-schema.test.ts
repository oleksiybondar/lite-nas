import { parseSensorHistory, sensorMarkersSchema } from "@schemas/monitoring/sensor";

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
