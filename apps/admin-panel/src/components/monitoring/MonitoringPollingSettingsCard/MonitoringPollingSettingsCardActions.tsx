import Button from "@mui/material/Button";
import Stack from "@mui/material/Stack";
import type { ReactElement } from "react";

type MonitoringPollingSettingsCardActionsProps = {
  /**
   * Handles reverting the editable card state to persisted settings.
   */
  onCancel: () => void;
  /**
   * Handles persisting the validated monitoring polling settings.
   */
  onApply: () => void;
  /**
   * Whether the current draft can be applied safely.
   */
  isApplyDisabled: boolean;
  /**
   * Source-scoped identifier used in stable test selectors.
   */
  storageKey: string;
};

/**
 * Action row rendered when one monitoring polling settings card has unsaved changes.
 */
export const MonitoringPollingSettingsCardActions = ({
  isApplyDisabled,
  onApply,
  onCancel,
  storageKey,
}: MonitoringPollingSettingsCardActionsProps): ReactElement => {
  return (
    <Stack
      data-test-class="monitoring-settings-actions"
      direction={{ sm: "row", xs: "column-reverse" }}
      justifyContent="flex-end"
      spacing={1.5}
    >
      <Button
        data-testid={`monitoring-settings-cancel-button-${storageKey}`}
        onClick={onCancel}
        variant="outlined"
      >
        Cancel
      </Button>
      <Button
        data-testid={`monitoring-settings-apply-button-${storageKey}`}
        disabled={isApplyDisabled}
        onClick={onApply}
        variant="contained"
      >
        Apply
      </Button>
    </Stack>
  );
};
