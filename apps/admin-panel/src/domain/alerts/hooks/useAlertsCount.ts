import type { AlertCountDTO, AlertCountResponseDTO, AlertDomain } from "@dto/alerts/alerts";
import { useApi } from "@hooks/useApi";
import { alertCountResponseSchema } from "@schemas/alerts/alert-count";
import { type UseQueryResult, useQuery } from "@tanstack/react-query";

/**
 * Polling and enablement options accepted by `useAlertsCount`.
 */
export type UseAlertsCountOptions = {
  /**
   * Whether the query is allowed to execute.
   */
  enabled?: boolean;
  /**
   * Polling interval in milliseconds, or `false` to disable polling.
   */
  refetchInterval?: false | number;
};

/**
 * Reads the unacknowledged alert count for the supplied domain.
 *
 * The hook owns the shared endpoint shape for system and security alert
 * counters while leaving permission gating to callers through `enabled`.
 */
export const useAlertsCount = (
  alertDomain: AlertDomain,
  options: UseAlertsCountOptions = {},
): UseQueryResult<AlertCountDTO> => {
  const { get } = useApi();
  const { enabled = true, refetchInterval = false } = options;

  return useQuery({
    enabled,
    queryFn: async () => {
      const response = await get(buildAlertCountPath(alertDomain)).execute();

      if (!response.ok) {
        throw new Error(`Failed to load ${alertDomain} alerts count.`);
      }

      const responseJson = (await response.json()) as AlertCountResponseDTO;
      return alertCountResponseSchema.parse(responseJson).data;
    },
    queryKey: getAlertsCountQueryKey(alertDomain),
    refetchInterval,
  });
};

/**
 * Returns the query key for an alert-count request.
 */
export const getAlertsCountQueryKey = (alertDomain: AlertDomain): string[] => {
  return ["alerts", alertDomain, "unacknowledged", "count"];
};

/**
 * Builds the unacknowledged-count endpoint path for the supplied alert domain.
 */
const buildAlertCountPath = (alertDomain: AlertDomain): string => {
  return `/api/alerts/${alertDomain}/unacknowledged/count`;
};
