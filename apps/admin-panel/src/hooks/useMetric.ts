import { MetricContext } from "@contexts/metric-context";
import type { MetricContextValue } from "@dto/monitoring/metric";
import { useContext } from "react";

/**
 * Reads one concrete metrics polling slice from the nearest metrics provider.
 */
export const useMetric = <TItem>(): MetricContextValue<TItem> => {
  const context = useContext(MetricContext);

  if (context === undefined) {
    throw new Error("useMetric must be used inside MetricProvider");
  }

  return context as MetricContextValue<TItem>;
};
