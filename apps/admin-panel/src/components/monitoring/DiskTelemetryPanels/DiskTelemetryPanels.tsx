import Stack from "@mui/material/Stack";
import type { ReactElement } from "react";
import { DiskAuxiliaryDevicesPanel } from "./DiskAuxiliaryDevicesPanel";
import { DiskDevicesPanel } from "./DiskDevicesPanel";
import { DiskMountsPanel } from "./DiskMountsPanel";
import type { DiskTelemetryPanelsProps } from "./types";

/**
 * Renders the disk telemetry sections as full-width workspace rows.
 */
export const DiskTelemetryPanels = ({
  auxiliaryDevices,
  capacity,
  devices,
  mounts,
}: DiskTelemetryPanelsProps): ReactElement => {
  return (
    <Stack data-testid="disk-telemetry-sections" spacing={1.5} width="100%">
      <DiskDevicesPanel capacity={capacity} devices={devices} />
      <DiskMountsPanel mounts={mounts} />
      <DiskAuxiliaryDevicesPanel devices={auxiliaryDevices} />
    </Stack>
  );
};
