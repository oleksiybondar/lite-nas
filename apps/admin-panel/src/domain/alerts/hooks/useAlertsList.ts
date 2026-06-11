import type {
  AlertCategory,
  AlertDomain,
  AlertListDTO,
  AlertListResponseDTO,
  AlertsFilterState,
} from "@dto/alerts/alerts";
import { buildAlertsListPath, buildAlertsQueryKey } from "@helpers/alerts";
import { useApi } from "@hooks/useApi";
import { alertListResponseSchema } from "@schemas/alerts/alert-list";
import { type UseQueryResult, useQuery } from "@tanstack/react-query";

/**
 * Query state accepted by `useAlertsList`.
 */
export type UseAlertsListInput = AlertsFilterState & {
  category: AlertCategory;
  domain: AlertDomain;
  page: number;
  pageSize: number;
};

/**
 * Reads one paginated alerts slice from the gateway-backed browser API.
 */
export const useAlertsList = ({
  category,
  categoryFilter,
  domain,
  page,
  pageSize,
  priorityFilter,
  severityFilter,
  sourceFilter,
}: UseAlertsListInput): UseQueryResult<AlertListDTO, Error> => {
  const { get } = useApi();
  const queryState = {
    category,
    categoryFilter,
    domain,
    page,
    pageSize,
    priorityFilter,
    severityFilter,
    sourceFilter,
  };

  return useQuery<AlertListDTO, Error>({
    placeholderData: (previousData) => previousData,
    queryFn: async () => {
      const response = await get(buildAlertsListPath(queryState)).execute();

      if (!response.ok) {
        throw new Error(`Failed to load ${domain} ${category} alerts.`);
      }

      const responseJson = (await response.json()) as AlertListResponseDTO;
      return alertListResponseSchema.parse(responseJson).data;
    },
    queryKey: buildAlertsQueryKey(queryState),
  });
};
