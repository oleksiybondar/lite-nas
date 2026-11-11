import { MetricProvider } from "@providers/MetricProvider";
import {
  parseSensorHistory,
  parseSensorSnapshot,
  type SensorSnapshot,
} from "@schemas/monitoring/sensor";
import type { PropsWithChildren, ReactElement } from "react";

/** Provides the complete sensor snapshot history to a sensor page or dashboard tile. */
export const SensorMetricProvider = ({ children }: PropsWithChildren): ReactElement => (
  <MetricProvider<SensorSnapshot>
    getTimestamp={(item) => item.timestamp}
    historyPath="/api/sensor-monitor/history"
    parseHistoryResponse={parseSensorHistory}
    parseSnapshotResponse={parseSensorSnapshot}
    snapshotPath="/api/sensor-monitor/snapshot"
    storageKey="sensor-monitor"
  >
    {children}
  </MetricProvider>
);
