/**
 * Pagination backends currently supported by the admin panel.
 */
export type PaginationMode = "client" | "server";

/**
 * Minimal mutable pagination state shared by local and backend-backed lists.
 */
export type PaginationState = {
  page: number;
  pageSize: number;
};

/**
 * Derived pagination metadata used by reusable controls and list summaries.
 */
export type PaginationMeta = {
  hasNextPage: boolean;
  hasPreviousPage: boolean;
  totalCount: number;
  totalPages: number;
};

/**
 * Shared pagination commands exposed by list providers and pagination hooks.
 */
export type PaginationActions = {
  nextPage: () => void;
  previousPage: () => void;
  resetPage: () => void;
  resetPagination: () => void;
  setPage: (page: number) => void;
  setPageSize: (pageSize: number) => void;
};

/**
 * Complete pagination controller contract shared across local and server modes.
 */
export type PaginationController = PaginationState &
  PaginationMeta &
  PaginationActions & {
    mode: PaginationMode;
  };

/**
 * State-only pagination controller used as the mutable source for adapters.
 */
export type PaginationStateController = PaginationState & PaginationActions;

/**
 * Minimal pagination state required by the reusable pagination control surface.
 */
export type PaginationControlState = Pick<
  PaginationController,
  "page" | "setPage" | "totalCount" | "totalPages"
>;

/**
 * Client-backed pagination controller that slices records already present in memory.
 */
export type ClientPaginationController<TRecord> = PaginationController & {
  mode: "client";
  records: TRecord[];
};

/**
 * Snapshot payload shape consumed by the shared server-pagination adapter.
 */
export type ServerPaginationSnapshot<TRecord> = {
  records: TRecord[];
  totalCount: number;
  totalPages: number;
};

/**
 * Server-backed pagination controller that reflects backend-provided page metadata.
 */
export type ServerPaginationController<TRecord> = PaginationController & {
  mode: "server";
  records: TRecord[];
};
