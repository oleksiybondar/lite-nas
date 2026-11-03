import { ProcessMetricContext } from "@contexts/process-metric-context";
import type { ProcessMetricContextValue } from "@dto/monitoring/process-metric";
import { useContext } from "react";

/**
 * Reads the gateway-backed process snapshot polling slice.
 */
export const useProcessMetric = (): ProcessMetricContextValue => {
  const context = useContext(ProcessMetricContext);

  if (context === undefined) {
    throw new Error("useProcessMetric must be used inside ProcessMetricProvider");
  }

  return context;
};
