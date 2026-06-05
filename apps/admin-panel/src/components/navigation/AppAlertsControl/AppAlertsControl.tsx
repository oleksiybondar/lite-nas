import { useAlertsCount } from "@domain/alerts/hooks/useAlertsCount";
import { useAuth } from "@hooks/useAuth";
import { useRbac } from "@hooks/useRbac";
import NotificationsRoundedIcon from "@mui/icons-material/NotificationsRounded";
import Box from "@mui/material/Box";
import IconButton from "@mui/material/IconButton";
import ListItemText from "@mui/material/ListItemText";
import Menu from "@mui/material/Menu";
import MenuItem from "@mui/material/MenuItem";
import Stack from "@mui/material/Stack";
import Tooltip from "@mui/material/Tooltip";
import type { MouseEvent, ReactElement } from "react";
import { useState } from "react";
import { useNavigate } from "react-router-dom";
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
  const navigate = useNavigate();
  const [anchorElement, setAnchorElement] = useState<HTMLElement | null>(null);

  if (!state.isVisible) {
    return null;
  }

  const closeMenu = (): void => {
    setAnchorElement(null);
  };

  return renderAlertsControl({
    ...state,
    anchorElement,
    onCloseMenu: closeMenu,
    onOpenMenu: (event) => {
      setAnchorElement(event.currentTarget);
    },
    onSelectSecurityAlerts: () => {
      closeMenu();
      void navigate("/alerts/security/unacknowledged");
    },
    onSelectSystemAlerts: () => {
      closeMenu();
      void navigate("/alerts/system/unacknowledged");
    },
  });
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
 * State and commands required to render the interactive alerts control.
 */
type AppAlertsControlRenderOptions = AppAlertsControlState & {
  anchorElement: HTMLElement | null;
  onCloseMenu: () => void;
  onOpenMenu: (event: MouseEvent<HTMLButtonElement>) => void;
  onSelectSecurityAlerts: () => void;
  onSelectSystemAlerts: () => void;
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
  anchorElement,
  onCloseMenu,
  onOpenMenu,
  onSelectSecurityAlerts,
  onSelectSystemAlerts,
  securityCount,
  showSecurityAlerts,
  showSystemAlerts,
  systemCount,
}: AppAlertsControlRenderOptions): ReactElement => {
  return (
    <>
      {renderAlertsTrigger({
        onOpenMenu,
        securityCount,
        showSecurityAlerts,
        showSystemAlerts,
        systemCount,
      })}
      {renderAlertsMenu({
        anchorElement,
        onClose: onCloseMenu,
        onSelectSecurityAlerts,
        onSelectSystemAlerts,
        securityCount,
        showSecurityAlerts,
        showSystemAlerts,
        systemCount,
      })}
    </>
  );
};

/**
 * State and commands required to render the clickable alerts trigger.
 */
type AlertsTriggerRenderOptions = {
  onOpenMenu: (event: MouseEvent<HTMLButtonElement>) => void;
  securityCount: number;
  showSecurityAlerts: boolean;
  showSystemAlerts: boolean;
  systemCount: number;
};

/**
 * Builds the clickable top-bar trigger for the alerts dropdown.
 */
const renderAlertsTrigger = ({
  onOpenMenu,
  securityCount,
  showSecurityAlerts,
  showSystemAlerts,
  systemCount,
}: AlertsTriggerRenderOptions): ReactElement => {
  return (
    <Tooltip
      title={buildAlertsTooltipLabel(
        systemCount,
        securityCount,
        showSystemAlerts,
        showSecurityAlerts,
      )}
    >
      <IconButton
        aria-label="Alerts menu"
        color="inherit"
        data-testid="alerts-control"
        data-test-class="alerts-control-button"
        onClick={onOpenMenu}
        sx={{
          borderColor: "divider",
          borderRadius: "15px",
          color: "text.secondary",
          position: "relative",
          minHeight: 42,
          minWidth: 42,
          p: 0,
        }}
      >
        {renderAlertsTriggerIcon()}
        {showSecurityAlerts ? (
          <AppAlertsIndicator count={securityCount} domain="security" position="top" />
        ) : null}
        {showSystemAlerts ? (
          <AppAlertsIndicator count={systemCount} domain="system" position="bottom" />
        ) : null}
      </IconButton>
    </Tooltip>
  );
};

/**
 * Builds the static bell icon surface inside the alerts trigger.
 */
const renderAlertsTriggerIcon = (): ReactElement => {
  return (
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
  );
};

/**
 * State and commands required to render the alerts dropdown menu.
 */
type AlertsMenuRenderOptions = {
  anchorElement: HTMLElement | null;
  onClose: () => void;
  onSelectSecurityAlerts: () => void;
  onSelectSystemAlerts: () => void;
  securityCount: number;
  showSecurityAlerts: boolean;
  showSystemAlerts: boolean;
  systemCount: number;
};

/**
 * Builds the alerts dropdown menu opened from the header control.
 */
const renderAlertsMenu = ({
  anchorElement,
  onClose,
  onSelectSecurityAlerts,
  onSelectSystemAlerts,
  securityCount,
  showSecurityAlerts,
  showSystemAlerts,
  systemCount,
}: AlertsMenuRenderOptions): ReactElement => {
  return (
    <Menu
      anchorEl={anchorElement}
      anchorOrigin={{ horizontal: "right", vertical: "bottom" }}
      data-testid="alerts-control-menu"
      onClose={onClose}
      open={anchorElement !== null}
      transformOrigin={{ horizontal: "right", vertical: "top" }}
    >
      {showSystemAlerts ? (
        <MenuItem data-testid="alerts-control-system-button" onClick={onSelectSystemAlerts}>
          <ListItemText primary={buildAlertsMenuLabel(systemCount, "system")} />
        </MenuItem>
      ) : null}
      {showSecurityAlerts ? (
        <MenuItem data-testid="alerts-control-security-button" onClick={onSelectSecurityAlerts}>
          <ListItemText primary={buildAlertsMenuLabel(securityCount, "security")} />
        </MenuItem>
      ) : null}
    </Menu>
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

/**
 * Builds the visible label for one alerts dropdown action.
 */
const buildAlertsMenuLabel = (count: number, domain: "security" | "system"): string => {
  return `${count} ${domain} alerts`;
};
