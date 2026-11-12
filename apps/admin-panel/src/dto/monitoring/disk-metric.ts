/**
 * Raw disk I/O counters returned for one device when available.
 */
export type DiskMetricDeviceIOCountersDTO = {
  discards_completed?: number | null;
  discards_merged?: number | null;
  discard_time_ms?: number | null;
  flushes_completed?: number | null;
  flush_time_ms?: number | null;
  io_in_progress: number;
  io_time_ms: number;
  reads_completed: number;
  reads_merged: number;
  read_time_ms: number;
  sectors_discarded?: number | null;
  sectors_read: number;
  sectors_written: number;
  weighted_io_time_ms: number;
  writes_completed: number;
  writes_merged: number;
  write_time_ms: number;
};

/**
 * One block-device-like snapshot item returned by disk metrics transport.
 */
export type DiskMetricDeviceSnapshotDTO = {
  connection_type: string;
  description?: string | null;
  io?: DiskMetricDeviceIOCountersDTO | null;
  kind: string;
  major: number;
  minor: number;
  model?: string | null;
  mounts?: string[] | null;
  name?: string | null;
  node: string;
  parent?: string | null;
  partitions?: string[] | null;
  removable?: boolean | null;
  rotational?: boolean | null;
  serial?: string | null;
  size_bytes: number;
  vendor?: string | null;
  wwn?: string | null;
  zpool_memberships?: string[] | null;
};

/**
 * One filesystem summary entry returned by the disk metrics transport.
 */
export type DiskMetricFilesystemSummaryDTO = {
  mount_count: number;
  type: string;
};

/**
 * One active mountpoint snapshot returned by the disk metrics transport.
 */
export type DiskMetricMountSnapshotDTO = {
  available_bytes?: number | null;
  device?: string | null;
  filesystem: string;
  free_bytes?: number | null;
  memory_backed: boolean;
  mountpoint: string;
  readonly: boolean;
  remote: boolean;
  source: string;
  total_bytes?: number | null;
  used_bytes?: number | null;
  used_percent?: number | null;
};

/**
 * One timestamped disk metrics snapshot item.
 */
export type DiskMetricSnapshotDTO = {
  devices: DiskMetricDeviceSnapshotDTO[] | null;
  filesystems: Record<string, DiskMetricFilesystemSummaryDTO> | null;
  mounts: Record<string, DiskMetricMountSnapshotDTO> | null;
  timestamp: string;
};

/**
 * Response envelope returned by disk metrics history endpoints.
 */
export type DiskMetricHistoryResponseDTO = {
  code?: string;
  data: DiskMetricSnapshotDTO[] | null;
  message?: string;
  request_id?: string;
  success: boolean;
  timestamp: string;
  trace_id?: string;
};

/**
 * Response envelope returned by disk metrics snapshot endpoints.
 */
export type DiskMetricSnapshotResponseDTO = {
  code?: string;
  data: DiskMetricSnapshotDTO;
  message?: string;
  request_id?: string;
  success: boolean;
  timestamp: string;
  trace_id?: string;
};
