import { AlertsControlPanel } from "@components/alerts/AlertsControlPanel";
import { AlertsControlPanelPagination } from "@components/alerts/AlertsControlPanel/AlertsControlPanelPagination";
import { AlertsTable } from "@components/alerts/AlertsTable";
import { buildAlertsPageSummary, buildAlertsPageTitle, formatAlertsLabel } from "@helpers/alerts";
import { useAlerts } from "@hooks/useAlerts";
import { useBulkAlertAcknowledge } from "@hooks/useBulkAlertAcknowledge";
import { useLanguage } from "@lite-nas/language";
import Alert from "@mui/material/Alert";
import Button from "@mui/material/Button";
import Snackbar from "@mui/material/Snackbar";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

/**
 * Draft page for status-specific alerts dashboards.
 */
export const AlertsDashboardPage = (): ReactElement => {
  const { category, domain, items, acknowledgeMany, isAcknowledging } = useAlerts();
  const { getLabel } = useLanguage();
  const title = buildAlertsPageTitle(category);
  const groupTitle = formatAlertsLabel(domain);
  const unacknowledgedIds = items.filter((item) => !item.Acknowledged).map((item) => item.EventID);
  const { notifications, isBulkAcknowledging, acknowledgeAll, dismiss } = useBulkAlertAcknowledge(
    unacknowledgedIds,
    acknowledgeMany,
  );

  return (
    <Stack data-testid="alerts-dashboard-page" minWidth={0} spacing={1} width="100%">
      <Typography color="primary" data-testid="alerts-dashboard-overline" variant="overline">
        {groupTitle}
      </Typography>
      <Typography data-testid="alerts-dashboard-title" variant="h1">
        {title}
      </Typography>
      <Typography color="text.secondary" data-testid="alerts-dashboard-summary" variant="body1">
        {buildAlertsPageSummary(domain, category)}
      </Typography>
      <AlertsControlPanel />
      <AlertsControlPanelPagination />
      <Button
        data-testid="alerts-acknowledge-all-button"
        disabled={isAcknowledging || isBulkAcknowledging || unacknowledgedIds.length === 0}
        onClick={() => {
          void acknowledgeAll();
        }}
        sx={{ alignSelf: "flex-end" }}
        variant="contained"
      >
        {getLabel("admin.alerts.acknowledgeAllOnPage")}
      </Button>
      <AlertsTable />
      <AlertsControlPanelPagination />
      <Stack
        data-testid="alerts-notifications"
        spacing={1}
        sx={{
          bottom: 16,
          flexDirection: "column-reverse",
          position: "fixed",
          right: 16,
          zIndex: (theme) => theme.zIndex.snackbar,
        }}
      >
        {notifications.map((notification) => (
          <Snackbar
            autoHideDuration={3000}
            key={notification.id}
            onClose={() => dismiss(notification.id)}
            open
            sx={{ position: "static", transform: "none" }}
          >
            <Alert severity={notification.severity} variant="filled">
              {notification.message}
            </Alert>
          </Snackbar>
        ))}
      </Stack>
    </Stack>
  );
};
