import { useLanguage } from "@lite-nas/language";
import { Paper, Stack, Switch, Typography } from "@mui/material";
import type { OperationalSafetyRule } from "@schemas/operational-safety-automation";
import type { ReactElement } from "react";

/** Displays one configured rule and its temporary enablement control. */
export const OperationalSafetyRuleCard = ({
  rule,
  onSetEnabled,
}: {
  rule: OperationalSafetyRule;
  onSetEnabled: (id: string, enabled: boolean) => void;
}): ReactElement => {
  const { getLabel } = useLanguage();
  const stateLabel = getLabel(
    rule.configured_enabled
      ? "admin.operationalSafety.enabled"
      : "admin.operationalSafety.disabled",
  );
  const toggleLabel = getLabel(
    rule.enabled ? "admin.operationalSafety.disableRule" : "admin.operationalSafety.enableRule",
  );
  return (
    <Paper
      variant="outlined"
      sx={{ p: 2, borderLeft: 4, borderLeftColor: rule.enabled ? "success.main" : "warning.main" }}
      data-test-class="operational-safety-rule"
      data-test-name={rule.id}
    >
      <Stack direction="row" alignItems="center" justifyContent="space-between" spacing={2}>
        <Stack spacing={0.5}>
          <Typography variant="h6">{rule.name}</Typography>
          <Typography variant="body2">
            {getLabel("admin.operationalSafety.ruleId")}: {rule.id}
          </Typography>
          <Typography variant="body2">
            {getLabel("admin.operationalSafety.startupState")}: {stateLabel} ·{" "}
            {getLabel("admin.operationalSafety.outputPrefix")}: {rule.output_id_prefix} ·{" "}
            {getLabel("admin.operationalSafety.cooldown")}: {rule.cooldown}
          </Typography>
        </Stack>
        <Switch
          checked={rule.enabled}
          disabled={!rule.configured_enabled}
          inputProps={{ "aria-label": toggleLabel }}
          onChange={(_, checked) => onSetEnabled(rule.id, checked)}
        />
      </Stack>
    </Paper>
  );
};
