import type {
  DiskAuxiliaryDeviceItem,
  DiskAuxiliaryDevicesPanelData,
} from "@helpers/disk-metric-panel";
import type { MetricChartLabel } from "@helpers/system-metric-chart";
import Chip from "@mui/material/Chip";
import Divider from "@mui/material/Divider";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

const auxiliaryDeviceSectionHeight = 760;

type DiskAuxiliaryDevicesPanelProps = {
  /**
   * Browser-facing categorized loop and non-primary block-device data.
   */
  devices: DiskAuxiliaryDevicesPanelData;
};

/**
 * Renders lightweight visibility sections for loop-backed and other non-primary block devices.
 */
export const DiskAuxiliaryDevicesPanel = ({
  devices,
}: DiskAuxiliaryDevicesPanelProps): ReactElement | null => {
  if (devices.loopDevices.length === 0 && devices.otherDevices.length === 0) {
    return null;
  }

  return (
    <Paper data-testid="disk-auxiliary-devices-panel" sx={{ p: 2, width: "100%" }}>
      <Stack spacing={1.25}>
        <Typography data-testid="disk-auxiliary-devices-panel-title" variant="h4">
          Loop & other block devices
        </Typography>
        {renderSummaryRow(devices.summaryLabels)}
        <Divider />
        <Stack direction="row" flexWrap="wrap" gap={1.5} useFlexGap>
          {renderDeviceSection(
            "Loop devices",
            "disk-loop-devices-section",
            devices.loopDevices,
            "No loop-backed virtual devices observed.",
          )}
          {renderDeviceSection(
            "Other block devices",
            "disk-other-block-devices-section",
            devices.otherDevices,
            "No secondary block devices observed.",
          )}
        </Stack>
      </Stack>
    </Paper>
  );
};

/**
 * Renders the summary chip row shown above the auxiliary device lists.
 */
const renderSummaryRow = (summaryLabels: MetricChartLabel[]): ReactElement => {
  return (
    <Stack
      data-testid="disk-auxiliary-devices-summary-row"
      direction="row"
      flexWrap="wrap"
      gap={1}
      useFlexGap
    >
      {summaryLabels.map((label) => {
        return (
          <Chip
            key={label.key}
            label={`${label.key}: ${label.value}`}
            size="small"
            variant="outlined"
          />
        );
      })}
    </Stack>
  );
};

/**
 * Renders one categorized auxiliary device section.
 */
const renderDeviceSection = (
  title: string,
  testId: string,
  devices: DiskAuxiliaryDeviceItem[],
  emptyLabel: string,
): ReactElement => {
  return (
    <Stack data-testid={testId} spacing={1} sx={{ flex: "1 1 320px", minWidth: 0 }}>
      <Typography variant="body2">{title}</Typography>
      <Stack
        spacing={1}
        sx={{
          maxHeight: auxiliaryDeviceSectionHeight,
          minHeight: auxiliaryDeviceSectionHeight,
          overflowY: "auto",
          pr: 0.5,
        }}
      >
        {devices.length > 0 ? (
          devices.map((device) => {
            return renderDeviceRow(device);
          })
        ) : (
          <Typography color="text.secondary" variant="body2">
            {emptyLabel}
          </Typography>
        )}
      </Stack>
    </Stack>
  );
};

/**
 * Renders one compact auxiliary device row with role and source metadata.
 */
const renderDeviceRow = (device: DiskAuxiliaryDeviceItem): ReactElement => {
  return (
    <Paper
      data-test-class="disk-auxiliary-device-row"
      data-test-name={device.name}
      key={device.name}
      sx={{ p: 1.25 }}
      variant="outlined"
    >
      <Stack spacing={0.5}>
        <Stack alignItems="center" direction="row" justifyContent="space-between" spacing={1}>
          <Typography
            data-test-class="disk-auxiliary-device-name"
            data-test-name={device.name}
            variant="body2"
          >
            {device.name}
          </Typography>
          <Chip label={device.roleLabel} size="small" variant="outlined" />
        </Stack>
        <Typography variant="body2">{device.subtitle}</Typography>
        <Typography variant="body2">{device.connectionLabel}</Typography>
        <Typography variant="body2">{device.mountsLabel}</Typography>
        <Typography variant="body2">Size: {device.sizeLabel}</Typography>
      </Stack>
    </Paper>
  );
};
