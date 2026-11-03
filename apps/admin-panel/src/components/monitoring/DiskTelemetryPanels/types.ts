import type {
  DiskAuxiliaryDevicesPanelData,
  DiskDevicePanelItem,
  DiskMountsPanelData,
} from "@helpers/disk-metric-panel";

/**
 * Props consumed by the disk telemetry workspace renderer.
 */
export type DiskTelemetryPanelsProps = {
  /**
   * Browser-facing categorized loop and non-primary block-device data.
   */
  auxiliaryDevices: DiskAuxiliaryDevicesPanelData;
  /**
   * Maximum number of values represented by the fixed X scale.
   */
  capacity: number;
  /**
   * Browser-facing disk device cards derived from metric history.
   */
  devices: DiskDevicePanelItem[];
  /**
   * Browser-facing mounts and filesystem data derived from the latest snapshot.
   */
  mounts: DiskMountsPanelData;
};
