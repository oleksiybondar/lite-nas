import type {
  PaginationStateController,
  ServerPaginationController,
  ServerPaginationSnapshot,
} from "@dto/pagination";
import { useEffect, useMemo } from "react";

type UseServerPaginationInput<TRecord> = {
  data: ServerPaginationSnapshot<TRecord> | null | undefined;
  pagination: PaginationStateController;
};

/**
 * Adapts one backend-provided paginated slice to the shared pagination contract.
 */
export const useServerPagination = <TRecord>({
  data,
  pagination,
}: UseServerPaginationInput<TRecord>): ServerPaginationController<TRecord> => {
  const totalCount = data?.totalCount ?? 0;
  const totalPages = data?.totalPages ?? 0;
  const maxPage = Math.max(totalPages, 1);
  const page = Math.min(pagination.page, maxPage);

  useEffect(() => {
    if (pagination.page !== page) {
      pagination.setPage(page);
    }
  }, [page, pagination]);

  return useMemo(() => {
    return {
      ...pagination,
      hasNextPage: totalPages > 0 && page < totalPages,
      hasPreviousPage: page > 1,
      mode: "server",
      page,
      records: data?.records ?? [],
      totalCount,
      totalPages,
    };
  }, [data?.records, page, pagination, totalCount, totalPages]);
};
