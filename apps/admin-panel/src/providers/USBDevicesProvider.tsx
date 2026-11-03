import { USBDevicesContext } from "@contexts/usb-devices-context";
import type { USBDevicesContextValue, USBDevicesPollingActions } from "@dto/management/usb-devices";
import { useApi } from "@hooks/useApi";
import { usePollingResource } from "@hooks/usePollingResource";
import { useRbac } from "@hooks/useRbac";
import { useLanguage } from "@lite-nas/language";
import { PollingResourceProvider } from "@providers/PollingResourceProvider";
import { managementActionResponseSchema } from "@schemas/management/management-action";
import {
  parseUSBDevices,
  type USBDevice,
  usbDeviceActionSchema,
} from "@schemas/management/usb-devices";
import { useQueryClient } from "@tanstack/react-query";
import type { PropsWithChildren, ReactElement } from "react";
import { useMemo, useState } from "react";

const usbDevicesQueryKey = ["polling-resource", "usb-devices", "snapshot"];

/** Provides the validated USBGuard inventory and authorization operations. */
export const USBDevicesProvider = ({ children }: PropsWithChildren): ReactElement => {
  const { getLabel } = useLanguage();
  const { post } = useApi();
  const { requireSecurity } = useRbac();
  const queryClient = useQueryClient();
  const [actionError, setActionError] = useState("");
  const allowed = requireSecurity();
  const actions = useMemo<USBDevicesPollingActions>(() => {
    const runUSBDeviceAction = async (
      device: USBDevice,
      action: "allow" | "block",
      permanent: boolean,
      // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The provider keeps device actions and their state updates together.
    ): Promise<void> => {
      setActionError("");
      try {
        const response = await post(
          `/api/usbguard/devices/${encodeURIComponent(device.id)}/${action}`,
          usbDeviceActionSchema.parse({ permanent }),
        ).execute();
        if (!response.ok) throw new Error(getLabel("admin.usbDevices.actionFailed"));
        managementActionResponseSchema.parse(await response.json());
        await queryClient.invalidateQueries({ queryKey: usbDevicesQueryKey });
      } catch (caught) {
        setActionError(getLabel("admin.usbDevices.actionFailed"));
        throw caught;
      }
    };
    return { actionError, runUSBDeviceAction };
  }, [actionError, getLabel, post, queryClient]);

  return (
    <PollingResourceProvider
      actions={actions}
      enabled={allowed}
      errorMessage={getLabel("admin.usbDevices.loadFailed")}
      parseResponse={parseUSBDevices}
      path="/api/usbguard/devices"
      queryKey={usbDevicesQueryKey}
    >
      <USBDevicesStateProvider allowed={allowed}>{children}</USBDevicesStateProvider>
    </PollingResourceProvider>
  );
};

/** Adds local filtering, pagination, and confirmation state to the USB snapshot. */
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: The provider keeps device actions and their state updates together.
const USBDevicesStateProvider = ({
  allowed,
  children,
}: PropsWithChildren<{ allowed: boolean }>): ReactElement => {
  const base = usePollingResource<USBDevice[], USBDevicesPollingActions>();
  const [search, setSearchValue] = useState("");
  const [page, setPage] = useState(1);
  const [pending, setPending] = useState("");
  const [permanent, setPermanent] = useState(false);
  const [confirmation, setConfirmation] = useState<NonNullable<
    USBDevicesContextValue["confirmation"]
  > | null>(null);
  const allDevices = base.snapshot ?? [];
  const normalizedSearch = search.toLocaleLowerCase();
  const devices = allDevices.filter((device) =>
    [device.id, device.product, device.hash, device.name, device.state, device.attributes]
      .join(" ")
      .toLocaleLowerCase()
      .includes(normalizedSearch),
  );
  const pageCount = Math.max(1, Math.ceil(devices.length / 20));
  const currentPage = Math.min(page, pageCount);
  const visibleDevices = devices.slice((currentPage - 1) * 20, currentPage * 20);
  const runAction = async (device: USBDevice, action: "allow" | "block"): Promise<void> => {
    setPending(device.id);
    try {
      await base.runUSBDeviceAction(device, action, permanent);
      setConfirmation(null);
    } finally {
      setPending("");
    }
  };
  const value: USBDevicesContextValue = {
    actionError: base.actionError,
    allowed,
    closeConfirmation: () => setConfirmation(null),
    confirmation,
    devices,
    isError: base.isError,
    isLoading: base.isLoading,
    page: currentPage,
    pageCount,
    pending,
    permanent,
    search,
    setPage,
    setPermanent,
    setSearch: (next) => {
      setSearchValue(next);
      setPage(1);
    },
    totalDevices: allDevices.length,
    visibleDevices,
    requestAction: (device, action) => setConfirmation({ device, action }),
    runAction,
  };
  return <USBDevicesContext.Provider value={value}>{children}</USBDevicesContext.Provider>;
};
