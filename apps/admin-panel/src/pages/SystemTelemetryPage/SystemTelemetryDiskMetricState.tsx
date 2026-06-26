import { DiskTelemetryPanels } from "@components/monitoring/DiskTelemetryPanels";
import type {
  DiskAuxiliaryDeviceItem,
  DiskDevicePanelItem,
  DiskFilesystemPanelRow,
  DiskMountPanelRow,
} from "@helpers/disk-metric-panel";
import {
  buildDiskAuxiliaryDevicesPanelData,
  buildDiskDevicePanelItems,
  buildDiskMountsPanelData,
} from "@helpers/disk-metric-panel";
import { useDiskMetric } from "@hooks/useDiskMetric";
import { type FilteredRecordSearchState, useFilteredRecords } from "@hooks/useFilteredRecords";
import { useMonitoringPollingSettings } from "@hooks/useMonitoringPollingSettings";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import { useMemo, useState } from "react";

type DiskTelemetryPanelsState = {
  auxiliaryDevices: ReturnType<typeof buildDiskAuxiliaryDevicesPanelData>;
  devices: DiskDevicePanelItem[];
  mounts: ReturnType<typeof buildDiskMountsPanelData>;
  totalDeviceCount: number;
  visibleDeviceCount: number;
};

type FilteredAuxiliaryDevices = {
  loopDevices: DiskAuxiliaryDeviceItem[];
  otherDevices: DiskAuxiliaryDeviceItem[];
};

type FilteredMountRows = {
  filesystemRows: DiskFilesystemPanelRow[];
  mountRows: DiskMountPanelRow[];
};

/**
 * Route state rendered when gateway-backed disk telemetry is available.
 */
export const SystemTelemetryDiskMetricState = (): ReactElement => {
  const { items } = useDiskMetric();
  const { maxRecords } = useMonitoringPollingSettings();
  const [search, setSearch] = useState("");
  const telemetryState = useDiskTelemetryPanelsState(items, { search, setSearch });

  return (
    <Stack spacing={3}>
      <Stack direction={{ md: "row", xs: "column" }} spacing={1.5}>
        <TextField
          data-testid="disk-metric-search-control"
          label="Search disks and mounts"
          name="diskMetricSearch"
          onChange={(event) => {
            setSearch(event.target.value);
          }}
          size="small"
          sx={{ minWidth: 240 }}
          value={search}
        />
        <Typography data-testid="disk-metric-search-summary" variant="body2">
          {telemetryState.visibleDeviceCount} of {telemetryState.totalDeviceCount} primary devices
          match the current search.
        </Typography>
      </Stack>
      <DiskTelemetryPanels
        auxiliaryDevices={telemetryState.auxiliaryDevices}
        capacity={maxRecords}
        devices={telemetryState.devices}
        mounts={telemetryState.mounts}
      />
    </Stack>
  );
};

/**
 * Derives the shared filtered disk telemetry panel data from the raw metric snapshot.
 */
const useDiskTelemetryPanelsState = (
  items: ReturnType<typeof useDiskMetric>["items"],
  searchState: FilteredRecordSearchState,
): DiskTelemetryPanelsState => {
  const auxiliaryDevices = useMemo(() => buildDiskAuxiliaryDevicesPanelData(items), [items]);
  const devices = useMemo(() => buildDiskDevicePanelItems(items), [items]);
  const mounts = useMemo(() => buildDiskMountsPanelData(items), [items]);
  const primaryDevices = useFilteredDiskDevices(devices, searchState);
  const filteredAuxiliaryDevices = useFilteredAuxiliaryDevices(auxiliaryDevices, searchState);
  const filteredMountRows = useFilteredMountRows(mounts, searchState);

  return {
    auxiliaryDevices: {
      ...auxiliaryDevices,
      ...filteredAuxiliaryDevices,
    },
    devices: primaryDevices.records,
    mounts: {
      ...mounts,
      ...filteredMountRows,
    },
    totalDeviceCount: primaryDevices.totalCount,
    visibleDeviceCount: primaryDevices.filteredCount,
  };
};

/**
 * Filters the primary device cards shown in the disk telemetry panels.
 */
const useFilteredDiskDevices = (
  devices: DiskDevicePanelItem[],
  searchState: FilteredRecordSearchState,
) => {
  return useFilteredRecords<DiskDevicePanelItem, string>({
    getSearchText: buildDiskDeviceSearchText,
    initialFilters: {},
    matchesFilter: () => true,
    records: devices,
    searchState,
  });
};

/**
 * Filters the auxiliary loop and non-loop device groups shown in the disk telemetry panels.
 */
const useFilteredAuxiliaryDevices = (
  auxiliaryDevices: ReturnType<typeof buildDiskAuxiliaryDevicesPanelData>,
  searchState: FilteredRecordSearchState,
): FilteredAuxiliaryDevices => {
  const { records: loopDevices } = useFilteredRecords<DiskAuxiliaryDeviceItem, string>({
    getSearchText: buildDiskAuxiliaryDeviceSearchText,
    initialFilters: {},
    matchesFilter: () => true,
    records: auxiliaryDevices.loopDevices,
    searchState,
  });
  const { records: otherDevices } = useFilteredRecords<DiskAuxiliaryDeviceItem, string>({
    getSearchText: buildDiskAuxiliaryDeviceSearchText,
    initialFilters: {},
    matchesFilter: () => true,
    records: auxiliaryDevices.otherDevices,
    searchState,
  });

  return { loopDevices, otherDevices };
};

/**
 * Filters the mount and filesystem table rows shown in the disk telemetry panels.
 */
const useFilteredMountRows = (
  mounts: ReturnType<typeof buildDiskMountsPanelData>,
  searchState: FilteredRecordSearchState,
): FilteredMountRows => {
  const { records: mountRows } = useFilteredRecords<DiskMountPanelRow, string>({
    getSearchText: buildDiskMountSearchText,
    initialFilters: {},
    matchesFilter: () => true,
    records: mounts.mountRows,
    searchState,
  });
  const { records: filesystemRows } = useFilteredRecords<DiskFilesystemPanelRow, string>({
    getSearchText: buildDiskFilesystemSearchText,
    initialFilters: {},
    matchesFilter: () => true,
    records: mounts.filesystemRows,
    searchState,
  });

  return { filesystemRows, mountRows };
};

/**
 * Builds the primary disk-device search blob used by the generic filtered-records hook.
 */
const buildDiskDeviceSearchText = (item: DiskDevicePanelItem): string => {
  return [item.name, item.subtitle, item.connectionLabel, item.mountsLabel, item.sizeLabel].join(
    " ",
  );
};

/**
 * Builds the auxiliary disk-device search blob used by the generic filtered-records hook.
 */
const buildDiskAuxiliaryDeviceSearchText = (item: DiskAuxiliaryDeviceItem): string => {
  return [
    item.name,
    item.subtitle,
    item.roleLabel,
    item.connectionLabel,
    item.mountsLabel,
    item.sizeLabel,
  ].join(" ");
};

/**
 * Builds the mount-row search blob used by the generic filtered-records hook.
 */
const buildDiskMountSearchText = (item: DiskMountPanelRow): string => {
  return [item.deviceLabel, item.mountpoint, item.source, item.filesystem].join(" ");
};

/**
 * Builds the filesystem-row search blob used by the generic filtered-records hook.
 */
const buildDiskFilesystemSearchText = (item: DiskFilesystemPanelRow): string => {
  return [item.filesystem, item.type].join(" ");
};
