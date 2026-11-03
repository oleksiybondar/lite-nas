import { AuthSessionsPage } from "@pages/AuthSessionsPage";
import { AutomatedAuditPage } from "@pages/AutomatedAuditPage";
import { AutomationPage } from "@pages/AutomationPage";
import { JournalPage } from "@pages/JournalPage";
import { LogsLandingPage } from "@pages/LogsLandingPage";
import { SecurityLandingPage } from "@pages/SecurityLandingPage";
import { AntivirusPage } from "@pages/SecurityTasks/AntivirusPage";
import { IntegrityPage } from "@pages/SecurityTasks/IntegrityPage";
import { SessionsLandingPage } from "@pages/SessionsLandingPage";
import { SessionsPage } from "@pages/SessionsPage";
import { USBDevicesPage } from "@pages/USBDevicesPage";
import { AutomatedAuditProvider } from "@providers/AutomatedAuditProvider";
import { AutomationProvider } from "@providers/AutomationProvider";
import type { RouteObject } from "react-router-dom";

/** Routes for security administration. Each page enforces security access. */
export const securityRoutes: RouteObject[] = [
  { element: <SecurityLandingPage />, path: "/security" },
  { element: <SessionsLandingPage />, path: "/security/sessions" },
  { element: <SessionsPage />, path: "/security/sessions/loginctl" },
  { element: <AuthSessionsPage />, path: "/security/sessions/litenas-auth" },
  { element: <JournalPage />, path: "/security/journal" },
  { element: <LogsLandingPage />, path: "/security/logs" },
  { element: <JournalPage />, path: "/security/logs/journal" },
  {
    element: (
      <AutomatedAuditProvider>
        <AutomatedAuditPage />
      </AutomatedAuditProvider>
    ),
    path: "/security/logs/automated-audit",
  },
  { element: <USBDevicesPage />, path: "/security/usb-devices" },
  { element: <IntegrityPage />, path: "/security/integrity" },
  { element: <AntivirusPage />, path: "/security/antivirus" },
  {
    element: (
      <AutomationProvider>
        <AutomationPage />
      </AutomationProvider>
    ),
    path: "/security/automation",
  },
];
