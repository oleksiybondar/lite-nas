import { usePollingQuery } from "@domain/monitoring/hooks/usePollingQuery";
import type { ServiceMetricSnapshotDTO } from "@dto/monitoring/service-metric";
import { parseServiceMetricSnapshotResponse } from "@schemas/monitoring/service-metric";

/** Reports whether a named systemd service is currently running. */
export const useSystemServiceAvailability = (unitName: string): boolean => {
  const { data, isError } = usePollingQuery<ServiceMetricSnapshotDTO>({
    errorMessage: "Failed to load service metrics snapshot.",
    parser: parseServiceMetricSnapshotResponse,
    path: "/api/service-metrics/snapshot",
    queryKey: ["polling-resource", "service-metrics", "snapshot"],
    refetchIntervalMs: 15_000,
  });

  return (
    !isError &&
    data?.services.some(
      (service) => service.name === unitName && service.sub_state === "running",
    ) === true
  );
};
