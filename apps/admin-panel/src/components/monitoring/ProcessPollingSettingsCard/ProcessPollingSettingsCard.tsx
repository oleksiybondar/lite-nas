import { MonitoringPollingSettingsProvider } from "@providers/MonitoringPollingSettingsProvider";
import type { ReactElement } from "react";
import { ProcessPollingSettingsCardContent } from "./ProcessPollingSettingsCardContent";
import type { ProcessPollingSettingsCardProps } from "./types";

/**
 * Interval-only polling settings card used by process and service snapshot views.
 */
export const ProcessPollingSettingsCard = ({
  description,
  storageKey,
  title,
}: ProcessPollingSettingsCardProps): ReactElement => {
  return (
    <MonitoringPollingSettingsProvider storageKey={storageKey}>
      <ProcessPollingSettingsCardContent
        description={description}
        storageKey={storageKey}
        title={title}
      />
    </MonitoringPollingSettingsProvider>
  );
};
