import type { MetricContextValue } from "@dto/monitoring/metric";
import { useMetric } from "@hooks/useMetric";
import type { SensorSnapshot } from "@schemas/monitoring/sensor";

/** Reads the complete sensor history from the nearest sensor provider. */
export const useSensorMetric = (): MetricContextValue<SensorSnapshot> =>
  useMetric<SensorSnapshot>();
