/**
 * CPU metrics sample returned by the system metrics transport.
 */
export type SystemMetricCPUSampleDTO = {
  PerCoreUsage: number[] | null;
  TotalUsagePct: number;
};

/**
 * Memory metrics sample returned by the system metrics transport.
 */
export type SystemMetricMemSampleDTO = {
  TotalBytes: number;
  UsedBytes: number;
  UsedPct: number;
};

/**
 * One timestamped system metrics snapshot item.
 */
export type SystemMetricSnapshotDTO = {
  CPU: SystemMetricCPUSampleDTO;
  Mem: SystemMetricMemSampleDTO;
  Timestamp: string;
};

/**
 * Response envelope returned by system metrics history endpoints.
 */
export type SystemMetricHistoryResponseDTO = {
  code?: string;
  data: SystemMetricSnapshotDTO[] | null;
  message?: string;
  request_id?: string;
  success: boolean;
  timestamp: string;
  trace_id?: string;
};

/**
 * Response envelope returned by system metrics snapshot endpoints.
 */
export type SystemMetricSnapshotResponseDTO = {
  code?: string;
  data: SystemMetricSnapshotDTO;
  message?: string;
  request_id?: string;
  success: boolean;
  timestamp: string;
  trace_id?: string;
};
