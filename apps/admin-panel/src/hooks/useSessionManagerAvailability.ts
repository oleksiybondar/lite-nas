import { useSystemServiceAvailability } from "@hooks/useSystemServiceAvailability";

/** Reports whether the host session manager is running and able to serve loginctl requests. */
export const useSessionManagerAvailability = (): boolean => {
  return useSystemServiceAvailability("lite-nas-session-manager.service");
};
