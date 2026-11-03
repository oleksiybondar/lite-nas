import {
  buildAlertOccurrenceSearchText,
  buildAlertOccurrencesSummary,
  buildSortedNumericAlertOccurrences,
  formatAlertOccurrenceValue,
  hasNumericAlertOccurrences,
} from "@components/alerts/AlertOccurrencesDialog/helpers";
import type { AlertOccurrenceItemDTO } from "@dto/alerts/alerts";

/** Creates an occurrence with explicit value variants for formatting tests. */
const occurrence = (overrides: Partial<AlertOccurrenceItemDTO> = {}): AlertOccurrenceItemDTO => ({
  EventID: "disk.temperature",
  EventRecID: 7,
  RecID: 9,
  Timestamp: "2026-11-02T10:00:00Z",
  ValueBool: null,
  ValueNum: null,
  ValueText: null,
  ValueType: "text",
  ValueUnit: null,
  ...overrides,
});

describe("alert occurrence formatting", () => {
  test.each([
    { value: occurrence({ ValueText: "overheated" }), expected: "overheated" },
    { value: occurrence({ ValueNum: 42, ValueUnit: "°C" }), expected: "42 °C" },
    { value: occurrence({ ValueBool: false }), expected: "false" },
    { value: occurrence(), expected: "-" },
  ])("formats $expected", ({ value, expected }) => {
    expect(formatAlertOccurrenceValue(value)).toBe(expected);
  });

  test.each([
    { values: [occurrence({ ValueNum: 1 })], expected: true },
    { values: [occurrence({ ValueNum: 1 }), occurrence()], expected: false },
    { values: [], expected: false },
  ])("detects numeric occurrence series", ({ values, expected }) => {
    expect(hasNumericAlertOccurrences(values)).toBe(expected);
  });

  test("summarizes a single chart occurrence", () => {
    expect(buildAlertOccurrencesSummary("disk.temperature", 1, true)).toBe(
      "Chart view. Event disk.temperature. 1 occurrence loaded.",
    );
  });

  test("summarizes multiple table occurrences", () => {
    expect(buildAlertOccurrencesSummary("disk.temperature", 2, false)).toBe(
      "Table view. Event disk.temperature. 2 occurrences loaded.",
    );
  });

  test("includes the event identifier and formatted value in search text", () => {
    expect(buildAlertOccurrenceSearchText(occurrence({ ValueNum: 42, ValueUnit: "°C" }))).toContain(
      "disk.temperature 7 42 °C",
    );
  });

  test("sorts numeric occurrences by timestamp", () => {
    const later = occurrence({ Timestamp: "2026-11-02T11:00:00Z", EventRecID: 2 });
    const earlier = occurrence({ Timestamp: "2026-11-02T10:00:00Z", EventRecID: 1 });
    expect(buildSortedNumericAlertOccurrences([later, earlier])).toEqual([earlier, later]);
  });
});
