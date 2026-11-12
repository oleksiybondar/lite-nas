import { highlightCatalogText } from "@helpers/catalog-search";
import { useRbac } from "@hooks/useRbac";
import { useLanguage } from "@lite-nas/language";
import BlockIcon from "@mui/icons-material/Block";
import CheckCircleIcon from "@mui/icons-material/CheckCircle";
import PlayArrowIcon from "@mui/icons-material/PlayArrow";
import RefreshIcon from "@mui/icons-material/Refresh";
import RestartAltIcon from "@mui/icons-material/RestartAlt";
import TerminalIcon from "@mui/icons-material/Terminal";
import UndoIcon from "@mui/icons-material/Undo";
import Alert from "@mui/material/Alert";
import Button from "@mui/material/Button";
import Dialog from "@mui/material/Dialog";
import DialogActions from "@mui/material/DialogActions";
import DialogContent from "@mui/material/DialogContent";
import DialogTitle from "@mui/material/DialogTitle";
import IconButton from "@mui/material/IconButton";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import Tooltip from "@mui/material/Tooltip";
import Typography from "@mui/material/Typography";
import { useAutomationContext } from "@providers/AutomationProvider";
import type { ActiveAutomation, AutomationRule } from "@schemas/security-automation";
import { type ReactElement, useState } from "react";

type Action = "enable" | "disable" | "mitigate" | "rollback" | "normalize" | "reset";
type Selection = { action: Action; incident?: ActiveAutomation; rule?: AutomationRule };

/** Selects the visible startup or temporary state label for one rule. */
const ruleStateLabel = (rule: AutomationRule): string => {
  if (!rule.configured_enabled) return "admin.automation.disabledInConfig";
  return rule.enabled ? "admin.automation.enabled" : "admin.automation.disabled";
};

/** Colors the name and row edge according to effective and startup enablement. */
const ruleTone = (rule: AutomationRule): string => {
  if (rule.enabled) return "success.main";
  if (rule.configured_enabled) return "text.secondary";
  return "warning.main";
};

/** Selects the operator effect with an implemented and owned host restore path. */
const operatorEffectAvailable = (rule: AutomationRule): boolean =>
  rule.action_kind === "builtin" && rule.action_name === "block_source_ip";

/** Restricts manual restoration controls to effects with a recorded inverse. */
const restorableIncident = (incident: ActiveAutomation): boolean =>
  incident.mitigation === "block_source_ip" ||
  incident.mitigation === "block_scan_source_ip" ||
  incident.mitigation === "usb_block_device";

/** Shows the owning host inventory used to verify each effect target. */
const rawInventoryCommand = (rule: AutomationRule): string => {
  if (rule.action_name === "block_source_ip" || rule.action_name === "block_scan_source_ip")
    return "sudo ufw status numbered";
  if (rule.action_name === "usb_block_device") return "sudo usbguard list-devices";
  if (rule.action_name === "login_terminate_session") return "loginctl list-sessions";
  return "";
};

/** Shows configured rules, current effects, and authenticated operator controls. */
// biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The page coordinates rule and incident controls in one view.
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: The page keeps the row controls and confirmation dialog together.
export const AutomationPage = (): ReactElement | null => {
  const { requireSecurity } = useRbac();
  const { getLabel } = useLanguage();
  const {
    rules,
    active,
    visibleRules,
    visibleActive,
    search,
    setSearch,
    loading,
    error,
    outcome,
    refresh,
    rollback,
    setEnabled,
    forceMitigate,
    resetCooldown,
  } = useAutomationContext();
  const [selection, setSelection] = useState<Selection | null>(null);
  const [busy, setBusy] = useState(false);
  const [typed, setTyped] = useState("");
  const [operands, setOperands] = useState<Record<string, string>>({});
  const [correlationKey, setCorrelationKey] = useState("");
  const [actionError, setActionError] = useState(false);
  const [detailsRule, setDetailsRule] = useState<AutomationRule | null>(null);
  if (!requireSecurity()) return null;

  const request = (selected: Selection) => {
    setTyped("");
    setOperands({});
    setCorrelationKey("");
    setActionError(false);
    setSelection(selected);
  };
  // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: Each action uses its specific validated provider operation.
  const confirm = async () => {
    if (!selection || typed !== "yes" || (selection.action === "reset" && !correlationKey)) return;
    setBusy(true);
    setActionError(false);
    try {
      if (selection.action === "enable" || selection.action === "disable") {
        if (selection.rule) await setEnabled(selection.rule.id, selection.action === "enable");
      } else if (selection.action === "mitigate") {
        if (selection.rule) await forceMitigate(selection.rule.id, operands);
      } else if (selection.action === "reset") {
        if (selection.rule) await resetCooldown(selection.rule.id, correlationKey);
      } else if (selection.incident) {
        await rollback(selection.incident.incident_id, selection.action === "normalize");
      }
      setSelection(null);
    } catch {
      setActionError(true);
    } finally {
      setBusy(false);
    }
  };

  return (
    <Stack data-testid="automation-page" spacing={3}>
      <Typography variant="h1" data-testid="automation-title">
        {getLabel("navigation.Automation")}
      </Typography>
      <Typography>{getLabel("admin.automation.summary")}</Typography>
      <Stack direction={{ xs: "column", sm: "row" }} alignItems="center" spacing={1}>
        <TextField
          data-testid="automation-search"
          label={getLabel("admin.automation.search")}
          value={search}
          onChange={(event) => setSearch(event.target.value)}
          fullWidth
          sx={{ flex: 1 }}
        />
        <Tooltip title={getLabel("admin.automation.refresh")}>
          <span>
            <IconButton
              data-testid="automation-refresh"
              aria-label={getLabel("admin.automation.refresh")}
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
      {error && <Alert severity="error">{getLabel("admin.automation.loadFailed")}</Alert>}
      {actionError && <Alert severity="error">{getLabel("admin.automation.actionFailed")}</Alert>}
      {outcome && <Alert severity="info">{outcome}</Alert>}
      {!loading && !error && rules.length === 0 && (
        <Typography>{getLabel("admin.automation.empty")}</Typography>
      )}
      {!loading && !error && rules.length > 0 && visibleRules.length === 0 && (
        <Typography>{getLabel("admin.automation.noRulesMatch")}</Typography>
      )}
      {
        // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: Rule controls reflect the existing action contract.
        // biome-ignore lint/complexity/noExcessiveLinesPerFunction: The rule card keeps its controls and visible fields together.
        visibleRules.map((rule) => (
          <Paper
            key={rule.id}
            variant="outlined"
            sx={{
              p: 2,
              borderLeft: 4,
              borderLeftColor: ruleTone(rule),
            }}
            data-test-class="automation-rule"
            data-test-name={rule.id}
          >
            <Stack
              direction={{ xs: "column", md: "row" }}
              justifyContent="space-between"
              spacing={2}
            >
              <Stack>
                <Typography color={ruleTone(rule)} fontWeight={600} variant="h6">
                  {highlightCatalogText(rule.name, search)}
                </Typography>
                <Typography variant="body2">
                  {highlightCatalogText(rule.description, search)}
                </Typography>
                {rule.explanation && (
                  <Typography variant="body2">
                    {highlightCatalogText(rule.explanation, search)}
                  </Typography>
                )}
                <Typography variant="body2">
                  {getLabel("admin.automation.action")}:{" "}
                  {highlightCatalogText(
                    [rule.action_kind, rule.action_name, ...(rule.required_operands ?? [])]
                      .filter(Boolean)
                      .join(" · "),
                    search,
                  )}
                </Typography>
                <Typography variant="body2">
                  {getLabel("admin.automation.ruleId")}: {highlightCatalogText(rule.id, search)}
                </Typography>
                <Typography variant="body2">
                  {getLabel(ruleStateLabel(rule))} ·{" "}
                  {getLabel(`admin.automation.state.${rule.state}`)} ·{" "}
                  {getLabel("admin.automation.activeCount")}: {rule.active_count}
                </Typography>
                {rule.last_outcome && (
                  <Typography variant="body2">
                    {getLabel("admin.automation.lastOutcome")}:{" "}
                    {highlightCatalogText(rule.last_outcome, search)}
                  </Typography>
                )}
              </Stack>
              <Stack direction="row" alignItems="center">
                <Tooltip title={getLabel("admin.automation.detailsAndCli")}>
                  <IconButton
                    color="info"
                    aria-label={getLabel("admin.automation.detailsAndCli")}
                    onClick={() => setDetailsRule(rule)}
                    data-testid={`automation-details-${rule.id}`}
                  >
                    <TerminalIcon />
                  </IconButton>
                </Tooltip>
                <Tooltip
                  title={getLabel(
                    rule.enabled
                      ? "admin.automation.disableTemporary"
                      : "admin.automation.enableTemporary",
                  )}
                >
                  <span>
                    <IconButton
                      color={rule.enabled ? "warning" : "success"}
                      aria-label={getLabel(
                        rule.enabled
                          ? "admin.automation.disableTemporary"
                          : "admin.automation.enableTemporary",
                      )}
                      data-testid={`automation-toggle-${rule.id}`}
                      disabled={busy || (!rule.enabled && !rule.configured_enabled)}
                      onClick={() => request({ action: rule.enabled ? "disable" : "enable", rule })}
                    >
                      {rule.enabled ? <BlockIcon /> : <CheckCircleIcon />}
                    </IconButton>
                  </span>
                </Tooltip>
                {operatorEffectAvailable(rule) && (
                  <Tooltip title={getLabel("admin.automation.forceMitigate")}>
                    <span>
                      <IconButton
                        color="error"
                        aria-label={getLabel("admin.automation.forceMitigate")}
                        data-testid={`automation-mitigate-${rule.id}`}
                        disabled={busy || !rule.enabled}
                        onClick={() => request({ action: "mitigate", rule })}
                      >
                        <PlayArrowIcon />
                      </IconButton>
                    </span>
                  </Tooltip>
                )}
                {rule.configured_enabled && (
                  <Tooltip title={getLabel("admin.automation.resetCooldown")}>
                    <span>
                      <IconButton
                        aria-label={getLabel("admin.automation.resetCooldown")}
                        data-testid={`automation-reset-${rule.id}`}
                        disabled={busy}
                        onClick={() => request({ action: "reset", rule })}
                      >
                        <RestartAltIcon />
                      </IconButton>
                    </span>
                  </Tooltip>
                )}
              </Stack>
            </Stack>
          </Paper>
        ))
      }
      <Typography variant="h2">{getLabel("admin.automation.activeTitle")}</Typography>
      {!loading && active.length === 0 && (
        <Typography>{getLabel("admin.automation.noActive")}</Typography>
      )}
      {!loading && active.length > 0 && visibleActive.length === 0 && (
        <Typography>{getLabel("admin.automation.noActiveMatch")}</Typography>
      )}
      {visibleActive.map((incident) => (
        <Paper
          key={incident.incident_id}
          variant="outlined"
          sx={{ p: 2 }}
          data-test-class="automation-incident"
          data-test-name={incident.incident_id}
        >
          <Stack direction={{ xs: "column", md: "row" }} justifyContent="space-between" spacing={2}>
            <Stack>
              <Typography fontWeight={600}>
                {highlightCatalogText(incident.rule.name, search)} ·{" "}
                {highlightCatalogText(incident.target, search)}
              </Typography>
              <Typography variant="body2">
                {highlightCatalogText(incident.rule.description, search)}
              </Typography>
              {incident.rule.explanation && (
                <Typography variant="body2">
                  {highlightCatalogText(incident.rule.explanation, search)}
                </Typography>
              )}
              <Typography variant="body2">
                {getLabel("admin.automation.ruleId")}:{" "}
                {highlightCatalogText(incident.rule.id, search)} ·{" "}
                {highlightCatalogText(incident.incident_id, search)}
              </Typography>
              <Typography variant="body2">
                {highlightCatalogText(incident.mitigation, search)} ·{" "}
                {getLabel(`admin.automation.state.${incident.state}`)}
              </Typography>
              <Typography variant="body2">
                {getLabel("admin.automation.started")}: {incident.started_at} ·{" "}
                {getLabel("admin.automation.expires")}: {incident.expires_at || "—"}
              </Typography>
              <Typography variant="body2">
                {getLabel("admin.automation.lastOutcome")}:{" "}
                {highlightCatalogText(incident.last_outcome, search)}
              </Typography>
              {restorableIncident(incident) && (
                <Typography component="code" variant="body2" sx={{ overflowWrap: "anywhere" }}>
                  {`security-automation-cli rollback ${incident.incident_id}`}
                </Typography>
              )}
            </Stack>
            {restorableIncident(incident) && (
              <Stack direction="row">
                <Tooltip title={getLabel("admin.automation.rollback")}>
                  <span>
                    <IconButton
                      color="warning"
                      aria-label={getLabel("admin.automation.rollback")}
                      onClick={() => request({ action: "rollback", incident })}
                      disabled={busy}
                    >
                      <UndoIcon />
                    </IconButton>
                  </span>
                </Tooltip>
                <Tooltip title={getLabel("admin.automation.forceNormalize")}>
                  <span>
                    <IconButton
                      color="success"
                      aria-label={getLabel("admin.automation.forceNormalize")}
                      onClick={() => request({ action: "normalize", incident })}
                      disabled={busy}
                    >
                      <CheckCircleIcon />
                    </IconButton>
                  </span>
                </Tooltip>
              </Stack>
            )}
          </Stack>
        </Paper>
      ))}
      <Dialog
        open={detailsRule !== null}
        onClose={() => setDetailsRule(null)}
        data-testid="automation-rule-details"
        fullWidth
      >
        <DialogTitle>{getLabel("admin.automation.detailsAndCli")}</DialogTitle>
        <DialogContent>
          {detailsRule && (
            <Stack spacing={1.5} sx={{ pt: 1 }}>
              <Typography color={ruleTone(detailsRule)} variant="h6">
                {detailsRule.name}
              </Typography>
              <Typography>{detailsRule.description}</Typography>
              {detailsRule.explanation && <Typography>{detailsRule.explanation}</Typography>}
              <Typography variant="subtitle2">
                {getLabel(
                  detailsRule.action_kind === "alert" || detailsRule.action_kind === "alert_only"
                    ? "admin.automation.actionBehavior"
                    : "admin.automation.applySteps",
                )}
              </Typography>
              <Typography component="pre" variant="body2" sx={{ overflowX: "auto", m: 0 }}>
                {detailsRule.action_kind === "alert_only" && getLabel("admin.automation.alertOnly")}
                {detailsRule.action_kind === "alert" && getLabel("admin.automation.reportAction")}
                {detailsRule.action_kind !== "alert_only" &&
                  detailsRule.action_kind !== "alert" &&
                  (operatorEffectAvailable(detailsRule)
                    ? `security-automation-cli force-mitigate ${detailsRule.id} ${detailsRule.required_operands?.map((name) => `${name}=VALUE`).join(" ") ?? ""} --yes`
                    : getLabel("admin.automation.automaticOnly"))}
              </Typography>
              {rawInventoryCommand(detailsRule) && (
                <Typography component="pre" variant="body2" sx={{ overflowX: "auto", m: 0 }}>
                  {rawInventoryCommand(detailsRule)}
                </Typography>
              )}
              <Typography variant="subtitle2">{getLabel("admin.automation.undoSteps")}</Typography>
              <Typography component="pre" variant="body2" sx={{ overflowX: "auto", m: 0 }}>
                {detailsRule.action_kind === "alert_only" ||
                detailsRule.action_kind === "alert" ||
                detailsRule.action_name === "login_terminate_session" ||
                detailsRule.action_name === "revoke_web_sessions"
                  ? getLabel("admin.automation.noEffect")
                  : "security-automation-cli active\nsecurity-automation-cli rollback <incident-id>"}
              </Typography>
              <Typography variant="body2">
                {getLabel("admin.automation.ruleId")}: {detailsRule.id}
              </Typography>
              <Typography variant="body2">
                {getLabel("admin.automation.configuredState")}:{" "}
                {getLabel(
                  detailsRule.configured_enabled
                    ? "admin.automation.enabled"
                    : "admin.automation.disabled",
                )}{" "}
                · {getLabel("admin.automation.effectiveState")}:{" "}
                {getLabel(
                  detailsRule.enabled ? "admin.automation.enabled" : "admin.automation.disabled",
                )}
              </Typography>
            </Stack>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDetailsRule(null)}>{getLabel("admin.automation.close")}</Button>
        </DialogActions>
      </Dialog>
      <Dialog
        open={selection !== null}
        onClose={() => !busy && setSelection(null)}
        data-testid="automation-confirmation"
      >
        <DialogTitle>
          {selection ? getLabel(`admin.automation.${selection.action}`) : ""}
        </DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ pt: 1 }}>
            <Typography>
              {getLabel(
                selection?.action === "reset"
                  ? "admin.automation.resetConfirm"
                  : "admin.automation.confirm",
              )}
            </Typography>
            {selection?.action === "mitigate" &&
              selection.rule?.required_operands?.map((name) => (
                <TextField
                  key={name}
                  label={getLabel(`admin.automation.operand.${name}`)}
                  value={operands[name] ?? ""}
                  onChange={(event) => setOperands({ ...operands, [name]: event.target.value })}
                  required
                  inputProps={{ "data-testid": `automation-operand-${name}` }}
                />
              ))}
            {selection?.action === "reset" && (
              <TextField
                label={getLabel("admin.automation.correlationKey")}
                value={correlationKey}
                onChange={(event) => setCorrelationKey(event.target.value)}
                required
                inputProps={{ "data-testid": "automation-correlation-key" }}
              />
            )}
            <TextField
              data-testid="automation-type-yes"
              label={getLabel("admin.automation.typeYes")}
              value={typed}
              onChange={(event) => setTyped(event.target.value)}
              required
            />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setSelection(null)} disabled={busy}>
            {getLabel("common.cancel")}
          </Button>
          <Button
            onClick={() => void confirm()}
            disabled={
              busy ||
              typed !== "yes" ||
              (selection?.action === "mitigate" &&
                (selection.rule?.required_operands?.some((name) => !operands[name]) ?? false)) ||
              (selection?.action === "reset" && !correlationKey)
            }
            color={selection?.action === "normalize" ? "error" : "primary"}
          >
            {getLabel("common.continue")}
          </Button>
        </DialogActions>
      </Dialog>
    </Stack>
  );
};
