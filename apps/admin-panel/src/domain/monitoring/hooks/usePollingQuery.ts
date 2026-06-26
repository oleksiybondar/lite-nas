import type { PollingResponseParser } from "@dto/monitoring/polling-resource";
import { useApi } from "@hooks/useApi";
import { type UseQueryResult, useQuery } from "@tanstack/react-query";

/**
 * Query state accepted by the shared polling transport hook.
 */
export type UsePollingQueryInput<TResult> = {
  enabled?: boolean;
  errorMessage: string;
  parser: PollingResponseParser<TResult>;
  path: string;
  queryKey: readonly unknown[];
  refetchIntervalMs?: number | false;
};

/**
 * Executes one generic gateway-backed polling request through TanStack Query.
 */
export const usePollingQuery = <TResult>({
  enabled = true,
  errorMessage,
  parser,
  path,
  queryKey,
  refetchIntervalMs,
}: UsePollingQueryInput<TResult>): UseQueryResult<TResult, Error> => {
  const { get } = useApi();
  const queryOptions = {
    enabled,
    queryFn: async (): Promise<TResult> => {
      const response = await get(path).execute();

      if (!response.ok) {
        throw new Error(errorMessage);
      }

      const responseJson = (await response.json()) as unknown;
      return parser(responseJson);
    },
    queryKey,
    ...(refetchIntervalMs === undefined ? {} : { refetchInterval: refetchIntervalMs }),
  };

  return useQuery<TResult, Error, TResult>(queryOptions);
};
