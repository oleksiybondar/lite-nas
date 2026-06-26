import { usePaginationState } from "@domain/pagination/hooks/usePaginationState";
import { getDefaultAlertsPage, getDefaultAlertsPageSize } from "@helpers/alerts";

/**
 * Shared pagination state owned by one alerts route slice.
 */
export type AlertsPaginationState = ReturnType<typeof usePaginationState>;

/**
 * Creates the shared pagination state used by the alerts provider.
 */
export const useAlertsPaginationState = (): AlertsPaginationState => {
  return usePaginationState({
    initialPage: getDefaultAlertsPage(),
    initialPageSize: getDefaultAlertsPageSize(),
  });
};
