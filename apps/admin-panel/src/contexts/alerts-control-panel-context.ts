import type { AlertsControlPanelContextValue } from "@dto/alerts/alerts";
import { createContext } from "react";

/**
 * Context for alerts page control-panel state derived from the route alerts slice.
 */
export const AlertsControlPanelContext = createContext<AlertsControlPanelContextValue | undefined>(
  undefined,
);
