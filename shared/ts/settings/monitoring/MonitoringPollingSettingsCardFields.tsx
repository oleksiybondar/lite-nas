import Stack from "@mui/material/Stack";
import type { ReactElement } from "react";
import type {
  MonitoringPollingSettingsCardProps,
  MonitoringPollingSettingsDraft,
  MonitoringPollingSettingsFieldErrors,
} from "./helpers";
import { MonitoringPollingSettingsCardHeader } from "./MonitoringPollingSettingsCardHeader";
import { MonitoringPollingSettingsModeField } from "./MonitoringPollingSettingsModeField";
import { MonitoringPollingSettingsNumericFields } from "./MonitoringPollingSettingsNumericFields";

type MonitoringPollingSettingsCardFieldsProps = MonitoringPollingSettingsCardProps & {
  /**
   * Editable draft state rendered by the resource form.
   */
  draft: MonitoringPollingSettingsDraft;
  /**
   * Field-level validation errors resolved from the monitoring settings schema.
   */
  fieldErrors: MonitoringPollingSettingsFieldErrors;
  /**
   * Updates one draft field in response to user edits.
   */
  onDraftChange: (name: keyof MonitoringPollingSettingsDraft, value: string) => void;
};

/**
 * Form fields rendered for one monitoring polling settings resource card.
 */
export const MonitoringPollingSettingsCardFields = ({
  description,
  draft,
  fieldErrors,
  onDraftChange,
  storageKey,
  title,
}: MonitoringPollingSettingsCardFieldsProps): ReactElement => {
  return (
    <Stack spacing={3}>
      <MonitoringPollingSettingsCardHeader
        description={description}
        storageKey={storageKey}
        title={title}
      />
      <MonitoringPollingSettingsModeField
        mode={draft.mode}
        onChange={(value) => {
          onDraftChange("mode", value);
        }}
        storageKey={storageKey}
      />
      <MonitoringPollingSettingsNumericFields
        draft={draft}
        fieldErrors={fieldErrors}
        onDraftChange={onDraftChange}
        storageKey={storageKey}
      />
    </Stack>
  );
};
