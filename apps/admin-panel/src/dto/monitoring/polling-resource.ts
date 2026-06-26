/**
 * Generic parser used to validate and extract one browser-facing polling payload.
 */
export type PollingResponseParser<TResult> = (value: unknown) => TResult;

/**
 * Generic runtime state exposed by one snapshot polling provider.
 */
export type PollingResourceContextValue<
  TSnapshot,
  TActions extends object = Record<string, never>,
> = {
  /**
   * Latest successfully parsed polling snapshot.
   */
  snapshot: TSnapshot | null;
  /**
   * Latest query error observed by the polling provider.
   */
  error: Error | null;
  /**
   * Whether the current provider session has produced an error.
   */
  isError: boolean;
  /**
   * Whether one polling request is currently in flight.
   */
  isFetching: boolean;
  /**
   * Whether the provider is still waiting for its initial polling result.
   */
  isLoading: boolean;
  /**
   * Refetches the currently active polling query.
   */
  refetch: () => Promise<unknown>;
} & TActions;
