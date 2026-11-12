import { PollingResourceContext } from "@contexts/polling-resource-context";
import { usePollingQuery } from "@domain/monitoring/hooks/usePollingQuery";
import type {
  PollingResourceContextValue,
  PollingResponseParser,
} from "@dto/monitoring/polling-resource";
import type { PropsWithChildren, ReactElement } from "react";
import { useMemo } from "react";

type PollingResourceProviderProps<
  TSnapshot,
  TActions extends object = Record<string, never>,
> = PropsWithChildren<{
  /**
   * Provider-owned action methods exposed alongside polling state.
   */
  actions: TActions;
  /**
   * Whether the polling query should execute.
   */
  enabled?: boolean;
  /**
   * Error message used when the transport request fails.
   */
  errorMessage: string;
  /**
   * Parses and validates one snapshot transport response.
   */
  parseResponse: PollingResponseParser<TSnapshot>;
  /**
   * Snapshot endpoint polled by the provider.
   */
  path: string;
  /**
   * Stable TanStack Query key for the polling resource.
   */
  queryKey: readonly unknown[];
  /**
   * Active polling interval in milliseconds, or `false` to disable polling.
   */
  refetchIntervalMs?: number | false;
}>;

/**
 * Provides one generic snapshot polling slice backed by one active polling query.
 */
export const PollingResourceProvider = <
  TSnapshot,
  TActions extends object = Record<string, never>,
>({
  actions,
  children,
  enabled,
  errorMessage,
  parseResponse,
  path,
  queryKey,
  refetchIntervalMs,
}: PollingResourceProviderProps<TSnapshot, TActions>): ReactElement => {
  const query = usePollingQuery<TSnapshot>({
    errorMessage,
    parser: parseResponse,
    path,
    queryKey,
    ...(enabled === undefined ? {} : { enabled }),
    ...(refetchIntervalMs === undefined ? {} : { refetchIntervalMs }),
  });

  const value = useMemo<PollingResourceContextValue<TSnapshot, TActions>>(() => {
    return {
      ...actions,
      error: query.error ?? null,
      isError: query.isError,
      isFetching: query.isFetching,
      isLoading: query.isLoading,
      refetch: async (): Promise<unknown> => query.refetch(),
      snapshot: query.data ?? null,
    };
  }, [actions, query]);

  return (
    <PollingResourceContext.Provider value={value}>{children}</PollingResourceContext.Provider>
  );
};
