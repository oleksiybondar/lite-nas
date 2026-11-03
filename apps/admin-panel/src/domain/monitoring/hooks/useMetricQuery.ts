import {
  type UsePollingQueryInput,
  usePollingQuery,
} from "@domain/monitoring/hooks/usePollingQuery";
import type { UseQueryResult } from "@tanstack/react-query";

/**
 * Query state accepted by the shared metrics transport hook.
 */
export type UseMetricQueryInput<TResult> = UsePollingQueryInput<TResult>;

/**
 * Executes one generic gateway-backed metrics request through TanStack Query.
 */
export const useMetricQuery = <TResult>(
  input: UseMetricQueryInput<TResult>,
): UseQueryResult<TResult, Error> => {
  return usePollingQuery(input);
};
