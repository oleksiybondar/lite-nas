import { useClientPagination } from "@domain/pagination/hooks/useClientPagination";
import { usePaginationState } from "@domain/pagination/hooks/usePaginationState";
import { useServerPagination } from "@domain/pagination/hooks/useServerPagination";
import { act, renderHook } from "@testing-library/react";

describe("shared pagination hooks", () => {
  test("resets the page when the page size changes", () => {
    const { result } = renderHook(() =>
      usePaginationState({
        initialPage: 1,
        initialPageSize: 10,
      }),
    );

    act(() => {
      result.current.setPage(3);
    });

    expect(result.current.page).toBe(3);

    act(() => {
      result.current.setPageSize(25);
    });

    expect(result.current.page).toBe(1);
    expect(result.current.pageSize).toBe(25);
  });
});

describe("shared client pagination adapter", () => {
  test("paginates in-memory records with the shared client adapter", () => {
    const { result } = renderHook(() => {
      const pagination = usePaginationState({
        initialPage: 2,
        initialPageSize: 2,
      });

      return useClientPagination({
        pagination,
        records: ["alpha", "beta", "gamma", "delta", "epsilon"],
      });
    });

    expect(result.current.mode).toBe("client");
    expect(result.current.records).toEqual(["gamma", "delta"]);
    expect(result.current.totalCount).toBe(5);
    expect(result.current.totalPages).toBe(3);
    expect(result.current.hasPreviousPage).toBe(true);
    expect(result.current.hasNextPage).toBe(true);
  });
});

describe("shared server pagination adapter", () => {
  test("derives backend pagination metadata with the shared server adapter", () => {
    const { result } = renderHook(() => {
      const pagination = usePaginationState({
        initialPage: 2,
        initialPageSize: 20,
      });

      return useServerPagination({
        data: {
          records: ["alert-3", "alert-4"],
          totalCount: 4,
          totalPages: 2,
        },
        pagination,
      });
    });

    expect(result.current.mode).toBe("server");
    expect(result.current.records).toEqual(["alert-3", "alert-4"]);
    expect(result.current.totalCount).toBe(4);
    expect(result.current.totalPages).toBe(2);
    expect(result.current.hasPreviousPage).toBe(true);
    expect(result.current.hasNextPage).toBe(false);
  });
});
