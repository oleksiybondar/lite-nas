import type { MetricResponseParser } from "@dto/monitoring/metric";
import { useApi } from "@hooks/useApi";
import { type UseQueryResult, useQuery } from "@tanstack/react-query";

/**
 * Query state accepted by the shared metrics transport hook.
 */
export type UseMetricQueryInput<TResult> = {
  enabled?: boolean;
  errorMessage: string;
  parser: MetricResponseParser<TResult>;
  path: string;
  queryKey: readonly unknown[];
  refetchIntervalMs?: number;
};

/**
 * Executes one generic gateway-backed metrics request through TanStack Query.
 */
export const useMetricQuery = <TResult>({
  enabled = true,
  errorMessage,
  parser,
  path,
  queryKey,
  refetchIntervalMs,
}: UseMetricQueryInput<TResult>): UseQueryResult<TResult, Error> => {
  const { get } = useApi();

  return useQuery<TResult, Error>({
    enabled,
    placeholderData: (previousData) => previousData,
    queryFn: async () => {
      const response = await get(path).execute();

      if (!response.ok) {
        throw new Error(errorMessage);
      }

      const responseJson = (await response.json()) as unknown;
      return parser(responseJson);
    },
    queryKey,
    refetchInterval: refetchIntervalMs,
  });
};
