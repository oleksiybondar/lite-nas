import type { USBDevice } from "@schemas/management/usb-devices";

/** A pending USBGuard authorization decision. */
export type USBDeviceConfirmation = { action: "allow" | "block"; device: USBDevice } | null;

/** State and operations exposed by the USB devices page provider. */
export type USBDevicesContextValue = {
  actionError: string;
  allowed: boolean;
  closeConfirmation: () => void;
  confirmation: USBDeviceConfirmation;
  devices: USBDevice[];
  isError: boolean;
  isLoading: boolean;
  page: number;
  pageCount: number;
  pending: string;
  permanent: boolean;
  search: string;
  setPage: (page: number) => void;
  setPermanent: (permanent: boolean) => void;
  setSearch: (search: string) => void;
  totalDevices: number;
  visibleDevices: USBDevice[];
  requestAction: (device: USBDevice, action: "allow" | "block") => void;
  runAction: (device: USBDevice, action: "allow" | "block") => Promise<void>;
};

/** Actions attached to the polled USBGuard inventory. */
export type USBDevicesPollingActions = {
  actionError: string;
  runUSBDeviceAction: (
    device: USBDevice,
    action: "allow" | "block",
    permanent: boolean,
  ) => Promise<void>;
};
