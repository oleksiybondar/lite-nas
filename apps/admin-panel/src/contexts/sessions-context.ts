import type { SessionsContextValue } from "@dto/management/sessions";
import { createContext } from "react";

/** Context for session inventory and management state scoped to the sessions page. */
export const SessionsContext = createContext<SessionsContextValue | undefined>(undefined);
