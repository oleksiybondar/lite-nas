import { PollingResourceContext } from "@contexts/polling-resource-context";
import type { PollingResourceContextValue } from "@dto/monitoring/polling-resource";
import { useContext } from "react";

/**
 * Reads one concrete snapshot polling slice from the nearest polling provider.
 */
export const usePollingResource = <
  TSnapshot,
  TActions extends object = Record<string, never>,
>(): PollingResourceContextValue<TSnapshot, TActions> => {
  const context = useContext(PollingResourceContext);

  if (context === undefined) {
    throw new Error("usePollingResource must be used inside PollingResourceProvider");
  }

  return context as PollingResourceContextValue<TSnapshot, TActions>;
};
