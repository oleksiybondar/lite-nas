/**
 * Extracts the shared pagination members exposed by local pagination adapters.
 */
export const buildPaginationContext = <
  TPagination extends {
    hasNextPage: boolean;
    hasPreviousPage: boolean;
    nextPage: () => void;
    page: number;
    pageSize: number;
    previousPage: () => void;
    resetPage: () => void;
    resetPagination: () => void;
    setPage: (page: number) => void;
    setPageSize: (pageSize: number) => void;
    totalPages: number;
  },
>(
  pagination: TPagination,
) => {
  return {
    hasNextPage: pagination.hasNextPage,
    hasPreviousPage: pagination.hasPreviousPage,
    nextPage: pagination.nextPage,
    page: pagination.page,
    pageSize: pagination.pageSize,
    previousPage: pagination.previousPage,
    resetPage: pagination.resetPage,
    resetPagination: pagination.resetPagination,
    setPage: pagination.setPage,
    setPageSize: pagination.setPageSize,
    totalPages: pagination.totalPages,
  };
};
