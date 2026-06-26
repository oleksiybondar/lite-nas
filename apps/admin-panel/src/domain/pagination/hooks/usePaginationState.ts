import type { PaginationStateController } from "@dto/pagination";
import { useMemo, useState } from "react";

export type UsePaginationStateInput = {
  initialPage?: number;
  initialPageSize?: number;
};

const defaultInitialPage = 1;
const defaultInitialPageSize = 20;

/**
 * Creates the mutable pagination state shared by client and server adapters.
 */
export const usePaginationState = ({
  initialPage = defaultInitialPage,
  initialPageSize = defaultInitialPageSize,
}: UsePaginationStateInput = {}): PaginationStateController => {
  const normalizedInitialPage = normalizePage(initialPage);
  const normalizedInitialPageSize = normalizePageSize(initialPageSize);
  const [page, setPageValue] = useState(normalizedInitialPage);
  const [pageSize, setPageSizeValue] = useState(normalizedInitialPageSize);

  return useMemo(() => {
    return {
      nextPage: () => {
        setPageValue((currentPage) => currentPage + 1);
      },
      page,
      pageSize,
      previousPage: () => {
        setPageValue((currentPage) => Math.max(defaultInitialPage, currentPage - 1));
      },
      resetPage: () => {
        setPageValue(normalizedInitialPage);
      },
      resetPagination: () => {
        setPageValue(normalizedInitialPage);
        setPageSizeValue(normalizedInitialPageSize);
      },
      setPage: (nextPage: number) => {
        setPageValue(normalizePage(nextPage));
      },
      setPageSize: (nextPageSize: number) => {
        setPageSizeValue(normalizePageSize(nextPageSize));
        setPageValue(normalizedInitialPage);
      },
    };
  }, [normalizedInitialPage, normalizedInitialPageSize, page, pageSize]);
};

/**
 * Normalizes one page number to the supported positive integer range.
 */
const normalizePage = (value: number): number => {
  if (!Number.isFinite(value)) {
    return defaultInitialPage;
  }

  return Math.max(defaultInitialPage, Math.trunc(value));
};

/**
 * Normalizes one page-size value to the supported positive integer range.
 */
const normalizePageSize = (value: number): number => {
  if (!Number.isFinite(value)) {
    return defaultInitialPageSize;
  }

  return Math.max(1, Math.trunc(value));
};
