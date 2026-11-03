import type { ServiceMetricContextValue } from "@dto/monitoring/service-metric";
import { createContext } from "react";

/**
 * Context for the service snapshot polling slice plus local search and filter state.
 */
export const ServiceMetricContext = createContext<ServiceMetricContextValue | undefined>(undefined);
