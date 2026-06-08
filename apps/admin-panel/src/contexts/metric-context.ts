import type { MetricContextValue } from "@dto/monitoring/metric";
import { createContext } from "react";

/**
 * Context for one concrete metrics polling slice.
 */
export const MetricContext = createContext<MetricContextValue<unknown> | undefined>(undefined);
