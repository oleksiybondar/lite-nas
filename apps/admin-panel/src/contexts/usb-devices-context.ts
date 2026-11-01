import type { USBDevicesContextValue } from "@dto/management/usb-devices";
import { createContext } from "react";

/** Context for USBGuard inventory and management state scoped to its page. */
export const USBDevicesContext = createContext<USBDevicesContextValue | null>(null);
