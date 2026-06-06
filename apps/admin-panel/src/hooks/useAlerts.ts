import { AlertsContext } from "@contexts/alerts-context";
import type { AlertsContextValue } from "@dto/alerts/alerts";
import { useContext } from "react";

/**
 * Reads shared alerts page state and commands from React context.
 */
export const useAlerts = (): AlertsContextValue => {
  const context = useContext(AlertsContext);

  if (context === undefined) {
    throw new Error("useAlerts must be used inside AlertsProvider");
  }

  return context;
};
