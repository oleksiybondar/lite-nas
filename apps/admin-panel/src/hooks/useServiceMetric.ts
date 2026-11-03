import { ServiceMetricContext } from "@contexts/service-metric-context";
import type { ServiceMetricContextValue } from "@dto/monitoring/service-metric";
import { useContext } from "react";

/**
 * Reads the gateway-backed services snapshot polling slice.
 */
export const useServiceMetric = (): ServiceMetricContextValue => {
  const context = useContext(ServiceMetricContext);

  if (context === undefined) {
    throw new Error("useServiceMetric must be used inside ServiceMetricProvider");
  }

  return context;
};
