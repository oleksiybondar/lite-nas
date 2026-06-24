import type {
  AlertDomain,
  AlertOccurrenceItemDTO,
  AlertOccurrencesResponseDTO,
} from "@dto/alerts/alerts";
import { buildAlertOccurrencesPath, buildAlertOccurrencesQueryKey } from "@helpers/alerts";
import { useApi } from "@hooks/useApi";
import { alertOccurrencesResponseSchema } from "@schemas/alerts/alert-occurrences";
import { type UseQueryResult, useQuery } from "@tanstack/react-query";

type UseAlertOccurrencesInput = {
  /**
   * Alert domain owning the occurrences endpoint.
   */
  domain: AlertDomain;
  /**
   * Stable alert event identifier used by the occurrences endpoints.
   */
  eventId: string;
  /**
   * Controls whether the query should execute.
   */
  enabled: boolean;
};

/**
 * Reads the full occurrence history for one alert event identifier.
 */
export const useAlertOccurrences = ({
  domain,
  enabled,
  eventId,
}: UseAlertOccurrencesInput): UseQueryResult<AlertOccurrenceItemDTO[], Error> => {
  const { get } = useApi();

  return useQuery<AlertOccurrenceItemDTO[], Error>({
    enabled,
    placeholderData: (previousData) => previousData,
    queryFn: async () => {
      const response = await get(buildAlertOccurrencesPath(domain, eventId)).execute();

      if (!response.ok) {
        throw new Error(`Failed to load ${domain} alert occurrences.`);
      }

      const responseJson = (await response.json()) as AlertOccurrencesResponseDTO;
      return alertOccurrencesResponseSchema.parse(responseJson).data;
    },
    queryKey: buildAlertOccurrencesQueryKey(domain, eventId),
  });
};
