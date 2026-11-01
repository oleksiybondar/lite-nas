import { CategoryLandingPage } from "@components/navigation/CategoryLandingPage";
import { useRbac } from "@hooks/useRbac";
import { useLanguage } from "@lite-nas/language";
import SecurityIcon from "@mui/icons-material/Security";
import SupervisorAccountIcon from "@mui/icons-material/SupervisorAccount";
import UsbIcon from "@mui/icons-material/Usb";
import type { ReactElement } from "react";

/** Security overview available to security officers and administrators. */
export const SecurityLandingPage = (): ReactElement | null => {
  const { requireSecurity } = useRbac();
  const { getLabel } = useLanguage();

  if (!requireSecurity()) return null;

  return (
    <CategoryLandingPage
      cards={[
        {
          description: getLabel("admin.landing.securitySessions"),
          icon: <SupervisorAccountIcon />,
          path: "/security/sessions",
          title: getLabel("navigation.Sessions"),
        },
        {
          description: getLabel("admin.landing.securityUSBDevices"),
          icon: <UsbIcon />,
          path: "/security/usb-devices",
          title: getLabel("navigation.USBDevices"),
        },
        {
          description: getLabel("admin.securityTasks.integrityDescription"),
          icon: <SecurityIcon />,
          path: "/security/integrity",
          title: getLabel("navigation.Integrity"),
        },
        {
          description: getLabel("admin.securityTasks.antivirusDescription"),
          icon: <SecurityIcon />,
          path: "/security/antivirus",
          title: getLabel("navigation.Antivirus"),
        },
      ]}
      overline={getLabel("navigation.Security")}
      summary={getLabel("admin.landing.securitySummary")}
      title={getLabel("navigation.Security")}
    />
  );
};
