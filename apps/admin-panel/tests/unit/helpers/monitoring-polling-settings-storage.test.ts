import {
  buildMonitoringPollingSettingsStorageKey,
  defaultMonitoringPollingSettings,
  loadMonitoringPollingSettings,
  normalizeMonitoringPollingSettings,
  resolveDefaultMonitoringPollingSettings,
  saveMonitoringPollingSettings,
} from "@helpers/monitoring-polling-settings-storage";

const monitoringPollingSettingsStorageKey =
  buildMonitoringPollingSettingsStorageKey("system-metrics");

describe("monitoring polling settings storage loading", () => {
  beforeEach(() => {
    window.localStorage.clear();
  });

  test("loads the shared monitoring defaults when storage is empty", () => {
    expect(loadMonitoringPollingSettings("system-metrics")).toEqual(
      defaultMonitoringPollingSettings,
    );
  });

  test("loads scope-specific defaults for process and service telemetry", () => {
    expect(loadMonitoringPollingSettings("process-metrics")).toEqual({
      historyIntervalMs: 1,
      historyResetGapMs: 1,
      maxRecords: 1,
      mode: "snapshot",
      snapshotIntervalMs: 5000,
    });
    expect(loadMonitoringPollingSettings("service-metrics")).toEqual({
      historyIntervalMs: 1,
      historyResetGapMs: 1,
      maxRecords: 1,
      mode: "snapshot",
      snapshotIntervalMs: 15000,
    });
  });

  test("loads saved settings for one source scope", () => {
    saveMonitoringPollingSettings("system-metrics", {
      historyIntervalMs: 20000,
      historyResetGapMs: 12000,
      maxRecords: 180,
      mode: "snapshot",
      snapshotIntervalMs: 1500,
    });

    expect(loadMonitoringPollingSettings("system-metrics")).toEqual({
      historyIntervalMs: 20000,
      historyResetGapMs: 12000,
      maxRecords: 180,
      mode: "snapshot",
      snapshotIntervalMs: 1500,
    });
  });

  test("uses defaults when saved settings are not valid JSON", () => {
    window.localStorage.setItem(monitoringPollingSettingsStorageKey, "{");

    expect(loadMonitoringPollingSettings("system-metrics")).toEqual(
      defaultMonitoringPollingSettings,
    );
  });
});

describe("monitoring polling settings storage normalization", () => {
  beforeEach(() => {
    window.localStorage.clear();
  });

  test("normalizes invalid fields individually instead of discarding valid values", () => {
    window.localStorage.setItem(
      monitoringPollingSettingsStorageKey,
      JSON.stringify({
        historyIntervalMs: 20000,
        historyResetGapMs: -1,
        maxRecords: 180,
        mode: "invalid",
        snapshotIntervalMs: 1500,
      }),
    );

    expect(loadMonitoringPollingSettings("system-metrics")).toEqual({
      historyIntervalMs: 20000,
      historyResetGapMs: defaultMonitoringPollingSettings.historyResetGapMs,
      maxRecords: 180,
      mode: defaultMonitoringPollingSettings.mode,
      snapshotIntervalMs: 1500,
    });
  });

  test("uses scope-specific defaults while normalizing unsupported input values", () => {
    const serviceDefaults = resolveDefaultMonitoringPollingSettings("service-metrics");

    expect(normalizeMonitoringPollingSettings(null)).toEqual(defaultMonitoringPollingSettings);
    expect(normalizeMonitoringPollingSettings("history", serviceDefaults)).toEqual(serviceDefaults);
  });
});
