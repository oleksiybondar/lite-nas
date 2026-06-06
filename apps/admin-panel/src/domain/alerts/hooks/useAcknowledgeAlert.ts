import type { AlertActionResponseDTO, AlertDomain } from "@dto/alerts/alerts";
import { useApi } from "@hooks/useApi";
import { useAuth } from "@hooks/useAuth";
import { alertActionResponseSchema } from "@schemas/alerts/alert-action";
import { type UseMutationResult, useMutation } from "@tanstack/react-query";

/**
 * Input accepted by the shared acknowledge-alert mutation.
 */
export type AcknowledgeAlertInput = {
  id: string;
};

/**
 * Creates the shared acknowledge-alert mutation for one alert domain.
 */
export const useAcknowledgeAlert = (
  domain: AlertDomain,
): UseMutationResult<void, Error, AcknowledgeAlertInput> => {
  const { post } = useApi();
  const { me } = useAuth();

  return useMutation({
    mutationFn: async ({ id }: AcknowledgeAlertInput): Promise<void> => {
      const response = await post(`/api/alerts/${domain}/${id}/acknowledge`, {}).execute();

      if (!response.ok) {
        throw new Error(`Failed to acknowledge ${domain} alert.`);
      }

      const responseJson = (await response.json()) as AlertActionResponseDTO;
      alertActionResponseSchema.parse(responseJson);
    },
    meta: {
      actorLogin: me?.user.login ?? "",
    },
  });
};
