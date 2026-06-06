import {
  buildCategoryOptions,
  buildPriorityOptions,
  buildSeverityOptions,
  buildSourceOptions,
  formatFilterSelection,
} from "@components/alerts/AlertsControlPanel/helpers";

describe("alerts control-panel helpers", () => {
  test("returns fixed priority options from one through five", () => {
    expect(buildPriorityOptions()).toEqual([
      { label: "1", value: 1 },
      { label: "2", value: 2 },
      { label: "3", value: 3 },
      { label: "4", value: 4 },
      { label: "5", value: 5 },
    ]);
  });

  test("returns the fixed supported severity options", () => {
    expect(buildSeverityOptions()).toEqual([
      { label: "Critical", value: "critical" },
      { label: "Error", value: "error" },
      { label: "Info", value: "info" },
      { label: "Warning", value: "warning" },
    ]);
  });

  test("returns the configured system source options", () => {
    expect(buildSourceOptions("system")).toEqual([
      { label: "resource-monitor", value: "resource-monitor" },
    ]);
  });

  test("returns empty category and security source options when no constants are configured", () => {
    expect(buildCategoryOptions("system")).toEqual([]);
    expect(buildCategoryOptions("security")).toEqual([]);
    expect(buildSourceOptions("security")).toEqual([]);
  });

  test("formats empty and populated filter selections", () => {
    expect(formatFilterSelection([])).toBe("All");
    expect(formatFilterSelection(["resource-monitor", "warning"])).toBe(
      "resource-monitor, warning",
    );
  });
});
