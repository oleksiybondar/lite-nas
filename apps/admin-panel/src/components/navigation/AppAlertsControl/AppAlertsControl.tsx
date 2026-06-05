import { useAlertsCount } from "@domain/alerts/hooks/useAlertsCount";
import { useAuth } from "@hooks/useAuth";
import { useRbac } from "@hooks/useRbac";
import NotificationsRoundedIcon from "@mui/icons-material/NotificationsRounded";
import Box from "@mui/material/Box";
import Stack from "@mui/material/Stack";
import Tooltip from "@mui/material/Tooltip";
import type { ReactElement } from "react";
import { AppAlertsIndicator } from "./AppAlertsIndicator";

/**
 * Polling interval used by the header alerts control.
 *
 * This value is local to the current draft until application settings own the
 * user-configurable polling cadence for live dashboard counters.
 */
const alertsRefetchIntervalMilliseconds = 30_000;

/**
 * Header alerts control shown for authenticated sessions.
 *
 * The control renders a shared alerts icon plus separate security and system
 * counters. Security counters are shown only for principals that satisfy the
 * RBAC security guard. The surrounding tooltip summarizes the currently visible
 * counts without exposing implementation details to callers.
 */
export const AppAlertsControl = (): ReactElement | null => {
  const state = useAppAlertsControlState();

  if (!state.isVisible) {
    return null;
  }

  return renderAlertsControl(state);
};

/**
 * Reads and normalizes the alert-control state derived from auth, RBAC, and
 * alert-count queries.
 */
const useAppAlertsControlState = (): AppAlertsControlState => {
  const { isAuthenticated } = useAuth();
  const { requireOperator, requireSecurity } = useRbac();
  const showSystemAlerts = requireOperator();
  const showSecurityAlerts = requireSecurity();
  const systemAlerts = useAlertsCount("system", {
    enabled: isAuthenticated && showSystemAlerts,
    refetchInterval: alertsRefetchIntervalMilliseconds,
  });
  const securityAlerts = useAlertsCount("security", {
    enabled: isAuthenticated && showSecurityAlerts,
    refetchInterval: alertsRefetchIntervalMilliseconds,
  });

  return {
    isVisible: isAlertsControlVisible(isAuthenticated, showSystemAlerts, showSecurityAlerts),
    securityCount: securityAlerts.data?.count ?? 0,
    showSecurityAlerts,
    showSystemAlerts,
    systemCount: systemAlerts.data?.count ?? 0,
  };
};

/**
 * Derived state required to render the header alerts control.
 */
type AppAlertsControlState = {
  isVisible: boolean;
  securityCount: number;
  showSecurityAlerts: boolean;
  showSystemAlerts: boolean;
  systemCount: number;
};

/**
 * Returns whether the alerts control should be visible for the current session.
 */
const isAlertsControlVisible = (
  isAuthenticated: boolean,
  showSystemAlerts: boolean,
  showSecurityAlerts: boolean,
): boolean => {
  return isAuthenticated && (showSystemAlerts || showSecurityAlerts);
};

/**
 * Builds the visible header alerts control.
 */
const renderAlertsControl = ({
  securityCount,
  showSecurityAlerts,
  showSystemAlerts,
  systemCount,
}: AppAlertsControlState): ReactElement => {
  return (
    <Tooltip
      title={buildAlertsTooltipLabel(
        systemCount,
        securityCount,
        showSystemAlerts,
        showSecurityAlerts,
      )}
    >
      <Box
        alignItems="center"
        data-testid="alerts-control"
        display="inline-flex"
        justifyContent="center"
        position="relative"
        sx={{
          borderColor: "divider",
          color: "text.secondary",
          minHeight: 42,
          minWidth: 42,
        }}
      >
        <Stack alignItems="center" direction="row" spacing={1}>
          <Box
            alignItems="center"
            data-testid="alerts-control-icon"
            display="inline-flex"
            justifyContent="center"
            sx={{
              bgcolor: "action.hover",
              borderRadius: "15px",
              color: "text.primary",
              height: 42,
              width: 42,
            }}
          >
            <NotificationsRoundedIcon data-testid="alerts-control-icon-svg" sx={{ fontSize: 38 }} />
          </Box>
        </Stack>
        {showSecurityAlerts ? (
          <AppAlertsIndicator count={securityCount} domain="security" position="top" />
        ) : null}
        {showSystemAlerts ? (
          <AppAlertsIndicator count={systemCount} domain="system" position="bottom" />
        ) : null}
      </Box>
    </Tooltip>
  );
};

/**
 * Builds the user-facing tooltip label for the alerts control.
 */
const buildAlertsTooltipLabel = (
  systemCount: number,
  securityCount: number,
  showSystemAlerts: boolean,
  showSecurityAlerts: boolean,
): string => {
  if (showSystemAlerts && !showSecurityAlerts) {
    return `System alerts: ${systemCount}`;
  }

  if (!showSystemAlerts && showSecurityAlerts) {
    return `Security alerts: ${securityCount}`;
  }

  return `Security alerts: ${securityCount}. System alerts: ${systemCount}`;
};
