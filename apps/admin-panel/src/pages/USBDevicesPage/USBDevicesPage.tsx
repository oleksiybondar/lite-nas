import { USBDevicesPageContent } from "@pages/USBDevicesPage/USBDevicesPageContent";
import { USBDevicesProvider } from "@providers/USBDevicesProvider";
import type { ReactElement } from "react";

/** Security-only USBGuard inventory and device authorization page. */
export const USBDevicesPage = (): ReactElement => (
  <USBDevicesProvider>
    <USBDevicesPageContent />
  </USBDevicesProvider>
);
