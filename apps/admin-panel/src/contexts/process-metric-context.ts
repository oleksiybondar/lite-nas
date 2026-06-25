import type { ProcessMetricContextValue } from "@dto/monitoring/process-metric";
import { createContext } from "react";

/**
 * Context for the process snapshot polling slice plus local search, sort, and pagination state.
 */
export const ProcessMetricContext = createContext<ProcessMetricContextValue | undefined>(undefined);
