import type { MonitoringPollingSettingsContextValue } from "@dto/monitoring/monitoring-polling-settings";
import { createContext } from "react";

/**
 * Context for source-scoped monitoring polling settings shared across UI and providers.
 */
export const MonitoringPollingSettingsContext = createContext<
  MonitoringPollingSettingsContextValue | undefined
>(undefined);
