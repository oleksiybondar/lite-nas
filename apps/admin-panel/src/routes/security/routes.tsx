import { AuthSessionsPage } from "@pages/AuthSessionsPage";
import { JournalPage } from "@pages/JournalPage";
import { SecurityLandingPage } from "@pages/SecurityLandingPage";
import { AntivirusPage } from "@pages/SecurityTasks/AntivirusPage";
import { IntegrityPage } from "@pages/SecurityTasks/IntegrityPage";
import { SessionsLandingPage } from "@pages/SessionsLandingPage";
import { SessionsPage } from "@pages/SessionsPage";
import { USBDevicesPage } from "@pages/USBDevicesPage";
import type { RouteObject } from "react-router-dom";

/** Routes for security administration. Each page enforces security access. */
export const securityRoutes: RouteObject[] = [
  { element: <SecurityLandingPage />, path: "/security" },
  { element: <SessionsLandingPage />, path: "/security/sessions" },
  { element: <SessionsPage />, path: "/security/sessions/loginctl" },
  { element: <AuthSessionsPage />, path: "/security/sessions/litenas-auth" },
  { element: <JournalPage />, path: "/security/journal" },
  { element: <USBDevicesPage />, path: "/security/usb-devices" },
  { element: <IntegrityPage />, path: "/security/integrity" },
  { element: <AntivirusPage />, path: "/security/antivirus" },
];
