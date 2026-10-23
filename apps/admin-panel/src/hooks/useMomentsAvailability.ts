import { usePollingQuery } from "@domain/monitoring/hooks/usePollingQuery";
import type { ServiceMetricSnapshotDTO } from "@dto/monitoring/service-metric";
import { parseServiceMetricSnapshotResponse } from "@schemas/monitoring/service-metric";

/** Reports whether the installed Moments systemd service has a running process. */
export const useMomentsAvailability = (): boolean => {
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
      (service) => service.name === "lite-nas-moments.service" && service.sub_state === "running",
    ) === true
  );
};
