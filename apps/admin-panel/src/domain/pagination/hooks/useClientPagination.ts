import type { ClientPaginationController, PaginationStateController } from "@dto/pagination";
import { useEffect, useMemo } from "react";

type UseClientPaginationInput<TRecord> = {
  pagination: PaginationStateController;
  records: TRecord[];
};

/**
 * Adapts one in-memory record list to the shared pagination controller contract.
 */
export const useClientPagination = <TRecord>({
  pagination,
  records,
}: UseClientPaginationInput<TRecord>): ClientPaginationController<TRecord> => {
  const totalCount = records.length;
  const totalPages = totalCount === 0 ? 0 : Math.ceil(totalCount / pagination.pageSize);
  const maxPage = Math.max(totalPages, 1);
  const page = Math.min(pagination.page, maxPage);

  useEffect(() => {
    if (pagination.page !== page) {
      pagination.setPage(page);
    }
  }, [page, pagination]);

  return useMemo(() => {
    const startIndex = (page - 1) * pagination.pageSize;
    const pagedRecords = records.slice(startIndex, startIndex + pagination.pageSize);

    return {
      ...pagination,
      hasNextPage: totalPages > 0 && page < totalPages,
      hasPreviousPage: page > 1,
      mode: "client",
      page,
      records: pagedRecords,
      totalCount,
      totalPages,
    };
  }, [page, pagination, records, totalCount, totalPages]);
};
