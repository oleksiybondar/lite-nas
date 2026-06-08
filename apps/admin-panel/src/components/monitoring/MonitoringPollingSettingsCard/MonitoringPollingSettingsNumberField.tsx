import TextField from "@mui/material/TextField";
import type { ReactElement } from "react";

type MonitoringPollingSettingsNumberFieldProps = {
  /**
   * Field-level validation message resolved for the current draft input.
   */
  errorMessage?: string;
  /**
   * Helper summary shown when the field has no validation error.
   */
  helperText: string;
  /**
   * Visible field label.
   */
  label: string;
  /**
   * Stable field name used for form inputs and tests.
   */
  name: string;
  /**
   * Draft update handler for the numeric input.
   */
  onChange: (value: string) => void;
  /**
   * Stable semantic selector name for repeated monitoring settings inputs.
   */
  testName: string;
  /**
   * Current string-backed numeric draft value.
   */
  value: string;
};

/**
 * Shared numeric text field used by monitoring polling settings forms.
 */
export const MonitoringPollingSettingsNumberField = ({
  errorMessage,
  helperText,
  label,
  name,
  onChange,
  testName,
  value,
}: MonitoringPollingSettingsNumberFieldProps): ReactElement => {
  return (
    <TextField
      data-test-class="monitoring-settings-input"
      data-test-name={testName}
      error={errorMessage !== undefined}
      fullWidth
      helperText={errorMessage ?? helperText}
      label={label}
      name={name}
      onChange={(event) => {
        onChange(event.target.value);
      }}
      slotProps={{ htmlInput: { min: 1 } }}
      type="number"
      value={value}
    />
  );
};
