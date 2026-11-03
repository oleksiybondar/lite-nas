import { useSystemServiceAvailability } from "@hooks/useSystemServiceAvailability";

/** Reports whether the installed Moments systemd service has a running process. */
export const useMomentsAvailability = (): boolean => {
  return useSystemServiceAvailability("lite-nas-moments.service");
};
