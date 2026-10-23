import MenuItem from "@mui/material/MenuItem";
import TextField from "@mui/material/TextField";
import type { ReactElement } from "react";
import type { MonitoringPollingSettingsDraft } from "./helpers";

type MonitoringPollingSettingsModeFieldProps = {
  /**
   * Current draft polling mode.
   */
  mode: MonitoringPollingSettingsDraft["mode"];
  /**
   * Updates the draft polling mode.
   */
  onChange: (value: string) => void;
  /**
   * Source-scoped identifier used in stable selector names.
   */
  storageKey: string;
};

/**
 * Select field used to switch the active monitoring polling mode.
 */
export const MonitoringPollingSettingsModeField = ({
  mode,
  onChange,
  storageKey,
}: MonitoringPollingSettingsModeFieldProps): ReactElement => {
  return (
    <TextField
      data-test-class="monitoring-settings-input"
      data-test-name="mode"
      fullWidth
      label="Polling mode"
      name={`monitoring-polling-mode-${storageKey}`}
      onChange={(event) => {
        onChange(event.target.value);
      }}
      select
      value={mode}
    >
      <MenuItem
        data-test-class="monitoring-settings-mode-option"
        data-test-name="History"
        value="history"
      >
        History
      </MenuItem>
      <MenuItem
        data-test-class="monitoring-settings-mode-option"
        data-test-name="Snapshot"
        value="snapshot"
      >
        Snapshot
      </MenuItem>
    </TextField>
  );
};
