import type { AlertsContextValue } from "@dto/alerts/alerts";
import { createContext } from "react";

/**
 * Context for one concrete alerts route slice.
 *
 * Consumers should use `useAlerts` so missing provider wiring fails with a
 * clear local error instead of leaking `undefined` checks into feature code.
 */
export const AlertsContext = createContext<AlertsContextValue | undefined>(undefined);
