import { USBDevicesContext } from "@contexts/usb-devices-context";
import type { USBDevicesContextValue } from "@dto/management/usb-devices";
import { useContext } from "react";

/** Reads USBGuard inventory and management actions from the page provider. */
export const useUSBDevices = (): USBDevicesContextValue => {
  const context = useContext(USBDevicesContext);
  if (!context) throw new Error("useUSBDevices must be used within USBDevicesProvider");
  return context;
};
