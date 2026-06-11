import { MonitoringPollingSettingsProvider } from "@providers/MonitoringPollingSettingsProvider";
import type { ReactElement } from "react";
import type { MonitoringPollingSettingsCardProps } from "./helpers";
import { MonitoringPollingSettingsCardContent } from "./MonitoringPollingSettingsCardContent";

/**
 * Editable monitoring polling settings card for one monitoring resource scope.
 */
export const MonitoringPollingSettingsCard = ({
  description,
  storageKey,
  title,
}: MonitoringPollingSettingsCardProps): ReactElement => {
  return (
    <MonitoringPollingSettingsProvider storageKey={storageKey}>
      <MonitoringPollingSettingsCardContent
        description={description}
        storageKey={storageKey}
        title={title}
      />
    </MonitoringPollingSettingsProvider>
  );
};
