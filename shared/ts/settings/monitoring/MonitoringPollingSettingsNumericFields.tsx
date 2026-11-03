import Stack from "@mui/material/Stack";
import type { ReactElement } from "react";
import type {
  MonitoringPollingSettingsDraft,
  MonitoringPollingSettingsFieldErrors,
} from "./helpers";
import { MonitoringPollingSettingsNumberField } from "./MonitoringPollingSettingsNumberField";

type MonitoringPollingSettingsNumericFieldsProps = {
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
  /**
   * Source-scoped identifier used in stable input names.
   */
  storageKey: string;
};

const numericFieldDefinitions = [
  {
    errorKey: "historyIntervalMs",
    helperText: "Polling interval used while history mode is active.",
    label: "History interval (ms)",
    namePrefix: "monitoring-history-interval",
    testName: "history-interval",
  },
  {
    errorKey: "snapshotIntervalMs",
    helperText: "Polling interval used while snapshot mode is active.",
    label: "Snapshot interval (ms)",
    namePrefix: "monitoring-snapshot-interval",
    testName: "snapshot-interval",
  },
  {
    errorKey: "maxRecords",
    helperText: "Maximum number of points retained in the browser cache.",
    label: "Max records",
    namePrefix: "monitoring-max-records",
    testName: "max-records",
  },
  {
    errorKey: "historyResetGapMs",
    helperText: "Gap threshold after which the next history result replaces the local cache.",
    label: "History reset gap (ms)",
    namePrefix: "monitoring-history-reset-gap",
    testName: "history-reset-gap",
  },
] as const satisfies ReadonlyArray<{
  errorKey: Exclude<keyof MonitoringPollingSettingsDraft, "mode">;
  helperText: string;
  label: string;
  namePrefix: string;
  testName: string;
}>;

/**
 * Numeric field group rendered for one monitoring polling settings card.
 */
export const MonitoringPollingSettingsNumericFields = ({
  draft,
  fieldErrors,
  onDraftChange,
  storageKey,
}: MonitoringPollingSettingsNumericFieldsProps): ReactElement => {
  return (
    <Stack spacing={3}>
      {numericFieldDefinitions.map((field) => {
        return (
          <MonitoringPollingSettingsNumberField
            errorMessage={fieldErrors[field.errorKey]?.[0]}
            helperText={field.helperText}
            key={field.errorKey}
            label={field.label}
            name={`${field.namePrefix}-${storageKey}`}
            onChange={(value) => {
              onDraftChange(field.errorKey, value);
            }}
            testName={field.testName}
            value={draft[field.errorKey]}
          />
        );
      })}
    </Stack>
  );
};
