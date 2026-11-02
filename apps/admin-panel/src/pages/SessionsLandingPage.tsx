import { CategoryLandingPage } from "@components/navigation/CategoryLandingPage";
import { useRbac } from "@hooks/useRbac";
import { useLanguage } from "@lite-nas/language";
import SupervisorAccountIcon from "@mui/icons-material/SupervisorAccount";
import VpnKeyIcon from "@mui/icons-material/VpnKey";
import Box from "@mui/material/Box";
import type { ReactElement } from "react";

/** Explains the host and application session inventories under Security. */
export const SessionsLandingPage = (): ReactElement | null => {
  const { requireSecurity, requireAdmin } = useRbac();
  const { getLabel } = useLanguage();
  if (!requireSecurity()) return null;

  return (
    <Box data-testid="sessions-landing-page">
      <CategoryLandingPage
        cards={[
          {
            description: getLabel("admin.landing.loginCtlSessions"),
            icon: <SupervisorAccountIcon />,
            path: "/security/sessions/loginctl",
            title: getLabel("navigation.loginctl"),
          },
          ...(requireAdmin()
            ? [
                {
                  description: getLabel("admin.landing.authSessions"),
                  icon: <VpnKeyIcon />,
                  path: "/security/sessions/litenas-auth",
                  title: getLabel("navigation.LiteNAS Auth"),
                },
              ]
            : []),
        ]}
        overline={getLabel("navigation.Security")}
        summary={getLabel("admin.landing.securitySessions")}
        title={getLabel("navigation.Sessions")}
      />
    </Box>
  );
};
