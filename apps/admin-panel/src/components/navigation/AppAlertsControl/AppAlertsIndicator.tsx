import type { AlertDomain } from "@dto/alerts/alerts";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

/**
 * Props accepted by one alerts counter badge.
 */
type AppAlertsIndicatorProps = {
  /**
   * Counter value currently represented by the badge.
   */
  count: number;
  /**
   * Alert domain represented by the badge.
   */
  domain: AlertDomain;
  /**
   * Vertical corner used by the badge on the shared alerts icon.
   */
  position: "bottom" | "top";
};

/**
 * Counter badge layered over the shared header alerts icon.
 *
 * The badge models one domain-specific alert counter and stays visually tied
 * to the parent alerts control through absolute corner positioning.
 */
export const AppAlertsIndicator = ({
  count,
  domain,
  position,
}: AppAlertsIndicatorProps): ReactElement => {
  const isSecurityDomain = domain === "security";

  return (
    <Box
      alignItems="center"
      data-test-class="alerts-indicator"
      data-test-name={domain}
      data-test-position={position}
      display="inline-flex"
      justifyContent="center"
      position="absolute"
      right={-2}
      sx={{
        [position]: -2,
        bgcolor: isSecurityDomain ? "error.main" : "warning.main",
        borderRadius: "999px",
        boxShadow: 1,
        color: isSecurityDomain ? "error.contrastText" : "warning.contrastText",
        height: 23,
        minWidth: 23,
        width: 23,
      }}
    >
      <Typography component="span" fontSize="0.8rem" lineHeight={1}>
        {formatAlertCount(count)}
      </Typography>
    </Box>
  );
};

/**
 * Formats one alert counter for compact badge display.
 */
const formatAlertCount = (count: number): string => {
  if (count > 20) {
    return "20+";
  }

  return String(count);
};
