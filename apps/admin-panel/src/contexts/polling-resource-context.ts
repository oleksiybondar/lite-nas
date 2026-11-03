import type { PollingResourceContextValue } from "@dto/monitoring/polling-resource";
import { createContext } from "react";

/**
 * Context for one concrete snapshot polling slice.
 */
export const PollingResourceContext = createContext<
  PollingResourceContextValue<unknown, Record<string, unknown>> | undefined
>(undefined);
