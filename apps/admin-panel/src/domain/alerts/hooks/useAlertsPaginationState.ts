import { getDefaultAlertsPage, getDefaultAlertsPageSize } from "@helpers/alerts";
import { useState } from "react";

/**
 * Shared pagination state owned by one alerts route slice.
 */
export type AlertsPaginationState = {
  nextPage: () => void;
  page: number;
  previousPage: () => void;
  setPage: (page: number) => void;
  setPageSize: (size: number) => void;
  pageSize: number;
  resetPage: () => void;
};

/**
 * Creates the shared pagination state used by the alerts provider.
 */
export const useAlertsPaginationState = (): AlertsPaginationState => {
  const [page, setPage] = useState(getDefaultAlertsPage);
  const [pageSize, setPageSizeValue] = useState(getDefaultAlertsPageSize);

  return {
    nextPage: () => {
      setPage((currentPage) => currentPage + 1);
    },
    page,
    pageSize,
    previousPage: () => {
      setPage((currentPage) => Math.max(getDefaultAlertsPage(), currentPage - 1));
    },
    resetPage: () => {
      setPage(getDefaultAlertsPage());
    },
    setPage: (nextPage: number) => {
      setPage(Math.max(getDefaultAlertsPage(), nextPage));
    },
    setPageSize: (nextPageSize: number) => {
      setPageSizeValue(nextPageSize);
      setPage(getDefaultAlertsPage());
    },
  };
};
