import { useState } from "react";

/**
 * Shared search-highlight state owned by one alerts route slice.
 */
export type AlertsSearchState = {
  search: string;
  setSearch: (value: string) => void;
};

/**
 * Creates the shared search-highlight state used by the alerts provider.
 */
export const useAlertsSearchState = (): AlertsSearchState => {
  const [search, setSearch] = useState("");

  return { search, setSearch };
};
