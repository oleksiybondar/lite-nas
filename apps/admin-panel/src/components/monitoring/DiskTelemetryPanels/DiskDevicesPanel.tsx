import type { DiskDevicePanelItem } from "@helpers/disk-metric-panel";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import { DiskDeviceCard } from "./DiskDeviceCard";

type DiskDevicesPanelProps = {
  /**
   * Maximum number of values represented by the fixed X scale.
   */
  capacity: number;
  /**
   * Latest browser-facing disk device rows rendered in the devices panel.
   */
  devices: DiskDevicePanelItem[];
};

/**
 * Renders the first draft of the disk devices panel with full device cards.
 */
export const DiskDevicesPanel = ({ capacity, devices }: DiskDevicesPanelProps): ReactElement => {
  return (
    <Paper data-testid="disk-devices-panel" sx={{ p: 2, width: "100%" }}>
      <Stack spacing={1.5}>
        <Typography data-testid="disk-devices-panel-title" variant="h2">
          Devices
        </Typography>
        {devices.length === 0 ? (
          <Typography color="text.secondary" data-testid="disk-devices-panel-empty" variant="body2">
            Disk device data will appear after telemetry points are loaded.
          </Typography>
        ) : (
          <Stack data-testid="disk-devices-subrow" spacing={1.5}>
            {devices.map((device) => {
              return <DiskDeviceCard capacity={capacity} device={device} key={device.name} />;
            })}
          </Stack>
        )}
      </Stack>
    </Paper>
  );
};
