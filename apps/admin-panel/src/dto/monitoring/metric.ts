import type { MonitoringPollingMode } from "@dto/monitoring/monitoring-polling-settings";

/**
 * Generic parser used to validate and extract one browser-facing metrics payload.
 */
export type MetricResponseParser<TResult> = (value: unknown) => TResult;

/**
 * Generic runtime state exposed by one metrics polling provider.
 */
export type MetricContextValue<TItem> = {
  /**
   * Current rolling metrics buffer after bootstrap and polling merges.
   */
  items: TItem[];
  /**
   * Most recent metrics item retained in the local rolling buffer.
   */
  latestItem: TItem | null;
  /**
   * Latest query error observed by bootstrap or active polling.
   */
  error: Error | null;
  /**
   * Whether the current provider session has produced an error.
   */
  isError: boolean;
  /**
   * Whether any bootstrap or active polling request is currently in flight.
   */
  isFetching: boolean;
  /**
   * Whether the provider is still waiting for its initial history bootstrap result.
   */
  isLoading: boolean;
  /**
   * Active polling mode used after the initial history bootstrap request.
   */
  mode: MonitoringPollingMode;
  /**
   * Refetches the currently active transport query.
   */
  refetch: () => Promise<unknown>;
};
