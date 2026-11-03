import { highlightCatalogText } from "@helpers/catalog-search";
import { useRbac } from "@hooks/useRbac";
import { useLanguage } from "@lite-nas/language";
import RefreshIcon from "@mui/icons-material/Refresh";
import Alert from "@mui/material/Alert";
import IconButton from "@mui/material/IconButton";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import Tooltip from "@mui/material/Tooltip";
import Typography from "@mui/material/Typography";
import { useAutomatedAudit } from "@providers/AutomatedAuditProvider";
import type { ReactElement } from "react";

/** Shows configured Log Analysis detections without mutation controls. */
// biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The read-only catalog shows loading, empty, and matching states together.
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: The rule card and search controls are one read-only view.
export const AutomatedAuditPage = (): ReactElement | null => {
  const { requireSecurity } = useRbac();
  const { getLabel } = useLanguage();
  const { rules, visibleRules, loading, error, search, setSearch, refresh } = useAutomatedAudit();
  const allowed = requireSecurity();

  if (!allowed) return null;

  return (
    <Stack data-testid="automated-audit-page" spacing={3}>
      <Typography variant="h1" data-testid="automated-audit-title">
        {getLabel("navigation.Automated Audit")}
      </Typography>
      <Typography>{getLabel("admin.logAudit.summary")}</Typography>
      <Stack direction={{ xs: "column", sm: "row" }} alignItems="center" spacing={1}>
        <TextField
          data-testid="automated-audit-search"
          label={getLabel("admin.logAudit.search")}
          value={search}
          onChange={(event) => setSearch(event.target.value)}
          fullWidth
          sx={{ flex: 1 }}
        />
        <Tooltip title={getLabel("admin.logAudit.refresh")}>
          <span>
            <IconButton
              data-testid="automated-audit-refresh"
              aria-label={getLabel("admin.logAudit.refresh")}
              disabled={loading}
              onClick={() => void refresh()}
              color="primary"
            >
              <RefreshIcon />
            </IconButton>
          </span>
        </Tooltip>
      </Stack>
      {loading && <Typography>{getLabel("common.loading")}</Typography>}
      {error && <Alert severity="error">{getLabel("admin.logAudit.loadFailed")}</Alert>}
      {!loading && !error && rules.length === 0 && (
        <Typography>{getLabel("admin.logAudit.empty")}</Typography>
      )}
      {!loading && !error && rules.length > 0 && visibleRules.length === 0 && (
        <Typography>{getLabel("admin.logAudit.noMatch")}</Typography>
      )}
      {!error &&
        // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: One card displays the catalog fields directly.
        visibleRules.map((rule) => (
          <Paper
            key={rule.id}
            variant="outlined"
            sx={{
              p: 2,
              borderLeft: 4,
              borderLeftColor: rule.enabled ? "success.main" : "warning.main",
            }}
            data-test-class="automated-audit-rule"
            data-test-name={rule.id}
          >
            <Stack spacing={1}>
              <Typography
                color={rule.enabled ? "success.main" : "warning.main"}
                fontWeight={600}
                variant="h6"
              >
                {highlightCatalogText(rule.name, search)}
              </Typography>
              <Typography variant="body2">
                {highlightCatalogText(rule.explanation, search)}
              </Typography>
              <Typography variant="body2">
                {getLabel("admin.automation.ruleId")}: {highlightCatalogText(rule.id, search)}
              </Typography>
              <Typography variant="body2">
                {getLabel(rule.enabled ? "admin.automation.enabled" : "admin.automation.disabled")}
                {" · "}
                {highlightCatalogText(rule.source, search)}
                {" · "}
                {highlightCatalogText(rule.category, search)}
                {" · "}
                {highlightCatalogText(rule.severity, search)}
              </Typography>
              <Typography variant="body2">
                {getLabel("admin.logAudit.instance")}: {highlightCatalogText(rule.instance, search)}
                {" · "}
                {getLabel("admin.logAudit.threshold")}:{" "}
                {highlightCatalogText(rule.threshold, search)}
                {" · "}
                {getLabel("admin.logAudit.window")}: {highlightCatalogText(rule.window, search)}
              </Typography>
              <Typography variant="body2">
                {getLabel("admin.logAudit.emailOnCritical")}:{" "}
                {getLabel(rule.email_on_critical ? "admin.logAudit.yes" : "admin.logAudit.no")}
                {" · "}
                {getLabel("admin.logAudit.lockdown")}:{" "}
                {getLabel(rule.lockdown ? "admin.logAudit.yes" : "admin.logAudit.no")}
              </Typography>
              {rule.possible_reactions.length > 0 && (
                <Typography variant="body2">
                  {getLabel("admin.logAudit.possibleReactions")}:{" "}
                  {highlightCatalogText(rule.possible_reactions.join(", "), search)}
                </Typography>
              )}
            </Stack>
          </Paper>
        ))}
    </Stack>
  );
};
