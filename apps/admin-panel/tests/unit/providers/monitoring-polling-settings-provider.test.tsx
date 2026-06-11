import { buildMonitoringPollingSettingsStorageKey } from "@helpers/monitoring-polling-settings-storage";
import { useMonitoringPollingSettings } from "@hooks/useMonitoringPollingSettings";
import { MonitoringPollingSettingsProvider } from "@providers/MonitoringPollingSettingsProvider";
import { fireEvent, render, screen } from "@testing-library/react";
import type { ReactElement } from "react";

beforeEach(() => {
  window.localStorage.clear();
});

describe("MonitoringPollingSettingsProvider", () => {
  test("exposes source-scoped monitoring polling settings and setters", () => {
    renderMonitoringPollingSettingsProvider();

    expectDefaultMonitoringPollingSettings();
  });
});

describe("MonitoringPollingSettingsProvider persistence", () => {
  test("persists individual setter updates", () => {
    renderMonitoringPollingSettingsProvider();

    fireEvent.click(screen.getByTestId("set-snapshot-mode"));
    fireEvent.click(screen.getByTestId("set-history-interval"));
    fireEvent.click(screen.getByTestId("set-snapshot-interval"));
    fireEvent.click(screen.getByTestId("set-max-records"));
    fireEvent.click(screen.getByTestId("set-history-reset-gap"));

    expect(screen.getByTestId("monitoring-mode")).toHaveTextContent("snapshot");
    expect(screen.getByTestId("monitoring-history-interval")).toHaveTextContent("20000");
    expect(screen.getByTestId("monitoring-snapshot-interval")).toHaveTextContent("1500");
    expect(screen.getByTestId("monitoring-max-records")).toHaveTextContent("180");
    expect(screen.getByTestId("monitoring-history-reset-gap")).toHaveTextContent("12000");
    expect(
      window.localStorage.getItem(buildMonitoringPollingSettingsStorageKey("system-metrics")),
    ).toBe(
      JSON.stringify({
        historyIntervalMs: 20000,
        historyResetGapMs: 12000,
        maxRecords: 180,
        mode: "snapshot",
        snapshotIntervalMs: 1500,
      }),
    );
  });
});

describe("MonitoringPollingSettingsProvider resets", () => {
  test("replaces the full settings object and resets to defaults", () => {
    renderMonitoringPollingSettingsProvider();

    fireEvent.click(screen.getByTestId("replace-settings"));

    expect(screen.getByTestId("monitoring-mode")).toHaveTextContent("snapshot");
    expect(screen.getByTestId("monitoring-history-interval")).toHaveTextContent("30000");
    expect(screen.getByTestId("monitoring-snapshot-interval")).toHaveTextContent("2000");
    expect(screen.getByTestId("monitoring-max-records")).toHaveTextContent("240");
    expect(screen.getByTestId("monitoring-history-reset-gap")).toHaveTextContent("8000");

    fireEvent.click(screen.getByTestId("reset-settings"));

    expectDefaultMonitoringPollingSettings();
  });
});

/**
 * Asserts the default monitoring polling settings exposed by the shared test probe.
 */
const expectDefaultMonitoringPollingSettings = (): void => {
  expect(screen.getByTestId("monitoring-mode")).toHaveTextContent("history");
  expect(screen.getByTestId("monitoring-history-interval")).toHaveTextContent("15000");
  expect(screen.getByTestId("monitoring-snapshot-interval")).toHaveTextContent("1000");
  expect(screen.getByTestId("monitoring-max-records")).toHaveTextContent("180");
  expect(screen.getByTestId("monitoring-history-reset-gap")).toHaveTextContent("10000");
};

/**
 * Test-only consumer that exposes monitoring polling settings through stable selectors.
 */
const MonitoringPollingSettingsProbe = (): ReactElement => {
  const context = useMonitoringPollingSettings();

  return (
    <>
      <span data-testid="monitoring-mode">{context.mode}</span>
      <span data-testid="monitoring-history-interval">{context.historyIntervalMs}</span>
      <span data-testid="monitoring-snapshot-interval">{context.snapshotIntervalMs}</span>
      <span data-testid="monitoring-max-records">{context.maxRecords}</span>
      <span data-testid="monitoring-history-reset-gap">{context.historyResetGapMs}</span>
      <button
        data-testid="set-snapshot-mode"
        onClick={() => context.setMode("snapshot")}
        type="button"
      />
      <button
        data-testid="set-history-interval"
        onClick={() => context.setHistoryIntervalMs(20000)}
        type="button"
      />
      <button
        data-testid="set-snapshot-interval"
        onClick={() => context.setSnapshotIntervalMs(1500)}
        type="button"
      />
      <button
        data-testid="set-max-records"
        onClick={() => context.setMaxRecords(180)}
        type="button"
      />
      <button
        data-testid="set-history-reset-gap"
        onClick={() => context.setHistoryResetGapMs(12000)}
        type="button"
      />
      <button
        data-testid="replace-settings"
        onClick={() => {
          context.setSettings({
            historyIntervalMs: 30000,
            historyResetGapMs: 8000,
            maxRecords: 240,
            mode: "snapshot",
            snapshotIntervalMs: 2000,
          });
        }}
        type="button"
      />
      <button data-testid="reset-settings" onClick={() => context.resetSettings()} type="button" />
    </>
  );
};

/**
 * Renders the monitoring polling settings provider around the shared test probe.
 */
const renderMonitoringPollingSettingsProvider = (): void => {
  render(
    <MonitoringPollingSettingsProvider storageKey="system-metrics">
      <MonitoringPollingSettingsProbe />
    </MonitoringPollingSettingsProvider>,
  );
};
