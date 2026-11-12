import { SessionsContext } from "@contexts/sessions-context";
import type { SessionsContextValue } from "@dto/management/sessions";
import { useContext } from "react";

/** Reads host session inventory and management actions from the sessions provider. */
export const useSessions = (): SessionsContextValue => {
  const context = useContext(SessionsContext);

  if (context === undefined) {
    throw new Error("useSessions must be used inside SessionsProvider");
  }

  return context;
};
