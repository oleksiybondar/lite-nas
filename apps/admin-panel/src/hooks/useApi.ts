import { ApiContext } from "@contexts/api-context";
import { useContext } from "react";
import type { ApiContextValue } from "../dto/api/api";

/**
 * Reads the app API client from React context.
 *
 * Feature components should use this hook to make BFF requests so unauthorized
 * handling remains centralized in `ApiProvider`.
 */
export const useApi = (): ApiContextValue => {
  const context = useContext(ApiContext);

  if (context === undefined) {
    throw new Error("useApi must be used inside ApiProvider");
  }

  return context;
};
