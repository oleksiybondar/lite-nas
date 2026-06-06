import { AlertsControlPanelContext } from "@contexts/alerts-control-panel-context";
import type { AlertsControlPanelContextValue } from "@dto/alerts/alerts";
import { useContext } from "react";

/**
 * Reads alerts control-panel state from the nearest dedicated provider.
 */
export const useAlertsControlPanel = (): AlertsControlPanelContextValue => {
  const context = useContext(AlertsControlPanelContext);

  if (context === undefined) {
    throw new Error("useAlertsControlPanel must be used inside AlertsControlPanelProvider");
  }

  return context;
};
