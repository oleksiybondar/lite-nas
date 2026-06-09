/**
 * Alert domains exposed by the gateway alerts API.
 */
export type AlertDomain = "security" | "system";

/**
 * Alert list categories currently exposed by the browser routes.
 */
export type AlertCategory = "active" | "all" | "unacknowledged";

/**
 * Alert severity values supported by current LiteNAS alert records.
 */
export type AlertSeverity = "critical" | "error" | "info" | "warning";

/**
 * One browser-facing alerts filter encoded for the gateway query string.
 */
export type AlertFilterDTO = {
  /**
   * Gateway filter key.
   */
  key: string;
  /**
   * Gateway filter condition.
   */
  condition: "eq" | "in";
  /**
   * Gateway filter values.
   */
  values: string[];
};

/**
 * Count payload returned by alert-count endpoints.
 */
export type AlertCountDTO = {
  /**
   * Total matching alert count.
   */
  count: number;
};

/**
 * Response envelope returned by alert-count endpoints.
 */
export type AlertCountResponseDTO = {
  /**
   * Common response metadata.
   */
  code?: string;
  /**
   * Count payload for the requested alert filter.
   */
  data: AlertCountDTO;
  /**
   * Common response metadata.
   */
  message?: string;
  /**
   * Common response metadata.
   */
  request_id?: string;
  /**
   * Whether the request completed successfully.
   */
  success: boolean;
  /**
   * Response creation timestamp.
   */
  timestamp: string;
  /**
   * Common response metadata.
   */
  trace_id?: string;
};

/**
 * One alert row returned by gateway-backed alert list endpoints.
 *
 * The field names intentionally mirror the current gateway response contract.
 */
export type AlertListItemDTO = {
  Acknowledged: boolean;
  AcknowledgedAt: string;
  AcknowledgedBy: string;
  Category: string;
  CreatedAt: string;
  EventID: string;
  EventRecID: number;
  LastEventID: string | null;
  LastEventRecID: number | null;
  LastRecID: number | null;
  LastTimestamp: string | null;
  LastValueBool: boolean | null;
  LastValueNum: number | null;
  LastValueText: string | null;
  LastValueType: string | null;
  LastValueUnit: string | null;
  Message: string;
  Meta?: Record<string, string> | undefined;
  Muted: boolean;
  MutedAt: string;
  MutedBy: string;
  Priority: number;
  RecID: number;
  Severity: AlertSeverity;
  Source: string;
  Status: string;
};

/**
 * Browser-facing pagination metadata returned with alert list pages.
 */
export type AlertListMetadataDTO = {
  page: number;
  size: number;
  total_count: number;
  total_pages: number;
};

/**
 * Browser-facing alert list payload returned by list endpoints.
 */
export type AlertListDTO = {
  items: AlertListItemDTO[];
  metadata: AlertListMetadataDTO;
};

/**
 * Response envelope returned by browser-facing alert list endpoints.
 */
export type AlertListResponseDTO = {
  code?: string;
  data: AlertListDTO;
  message?: string;
  request_id?: string;
  success: boolean;
  timestamp: string;
  trace_id?: string;
};

/**
 * Browser-facing alert action response returned by acknowledge endpoints.
 */
export type AlertActionResponseDTO = {
  code?: string;
  data?: Record<string, never>;
  message?: string;
  request_id?: string;
  success: boolean;
  timestamp: string;
  trace_id?: string;
};

/**
 * Separate UI-managed filter states exposed by the shared alerts context.
 */
export type AlertsFilterState = {
  categoryFilter: string[];
  priorityFilter: number[];
  severityFilter: AlertSeverity[];
  sourceFilter: string[];
};

/**
 * Shared filter setters and page controls used by alerts list views and control-panel state.
 */
export type AlertsFilterControls = {
  clearFilters: () => void;
  page: number;
  search: string;
  setCategoryFilter: (value: string[]) => void;
  setPage: (page: number) => void;
  setPriorityFilter: (value: number[]) => void;
  setSearch: (value: string) => void;
  setSeverityFilter: (value: AlertSeverity[]) => void;
  setSourceFilter: (value: string[]) => void;
  totalCount: number;
  totalPages: number;
};

/**
 * Browser-facing alerts page state and commands shared across one route slice.
 */
export type AlertsContextValue = AlertsFilterState &
  AlertsFilterControls & {
    acknowledge: (id: string) => Promise<void>;
    acknowledgeMany: (ids: string[]) => Promise<void>;
    apiPath: string;
    category: AlertCategory;
    domain: AlertDomain;
    error: Error | null;
    isAcknowledging: boolean;
    isError: boolean;
    isFetching: boolean;
    isLoading: boolean;
    items: AlertListItemDTO[];
    nextPage: () => void;
    pageSize: number;
    previousPage: () => void;
    queryKey: readonly unknown[];
    refetch: () => Promise<unknown>;
    routePath: string;
    setPageSize: (size: number) => void;
  };

/**
 * One selectable option rendered by alerts control-panel filter inputs.
 */
export type AlertsControlPanelOption<T extends string | number> = {
  label: string;
  value: T;
};

/**
 * Focused UI contract exposed by the alerts control-panel provider.
 */
export type AlertsControlPanelContextValue = AlertsFilterState &
  AlertsFilterControls & {
    availableCategoryOptions: AlertsControlPanelOption<string>[];
    availablePriorityOptions: AlertsControlPanelOption<number>[];
    availableSeverityOptions: AlertsControlPanelOption<AlertSeverity>[];
    availableSourceOptions: AlertsControlPanelOption<string>[];
  };

/**
 * Input contract used to derive one alerts control-panel context value.
 *
 * This keeps the provider/helper boundary explicit without repeating the same
 * state and setter signatures inline at the call site.
 */
export type BuildAlertsControlPanelValueInput = AlertsFilterState &
  AlertsFilterControls & {
    domain: AlertDomain;
  };
