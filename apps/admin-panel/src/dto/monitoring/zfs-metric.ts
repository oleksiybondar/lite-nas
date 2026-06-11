/**
 * One ZFS read/write sample group returned by the transport.
 */
export type ZFSMetricIOStatValuesDTO = {
  Read: number;
  Write: number;
};

/**
 * ZFS I/O error counts returned by the transport.
 */
export type ZFSMetricIOErrorsDTO = {
  Checksum: number;
  Read: number;
  Write: number;
};

/**
 * ZFS I/O statistics returned for one pool.
 */
export type ZFSMetricIOStatDTO = {
  Bandwidth: ZFSMetricIOStatValuesDTO;
  Operations: ZFSMetricIOStatValuesDTO;
};

/**
 * ZFS usage metrics returned for one pool.
 */
export type ZFSMetricUsageDTO = {
  AllocatedBytes: number;
  CapacityPct: number;
  FreeBytes: number;
  SizeBytes: number;
};

/**
 * Recursive ZFS vdev snapshot returned by the transport.
 */
export type ZFSMetricVdevSnapshotDTO = {
  Children: ZFSMetricVdevSnapshotDTO[] | null;
  Errors: ZFSMetricIOErrorsDTO;
  Name: string;
  Path: string;
  Type: string;
};

/**
 * One ZFS pool snapshot item returned by the transport.
 */
export type ZFSMetricPoolSnapshotDTO = {
  Errors: string;
  Health: string;
  IOStat: ZFSMetricIOStatDTO;
  Name: string;
  Root: ZFSMetricVdevSnapshotDTO;
  Scan: string;
  Usage: ZFSMetricUsageDTO;
};

/**
 * One timestamped ZFS metrics snapshot item.
 */
export type ZFSMetricSnapshotDTO = {
  Pools: ZFSMetricPoolSnapshotDTO[] | null;
  Timestamp: string;
};

/**
 * Response envelope returned by ZFS metrics history endpoints.
 */
export type ZFSMetricHistoryResponseDTO = {
  code?: string;
  data: ZFSMetricSnapshotDTO[] | null;
  message?: string;
  request_id?: string;
  success: boolean;
  timestamp: string;
  trace_id?: string;
};

/**
 * Response envelope returned by ZFS metrics snapshot endpoints.
 */
export type ZFSMetricSnapshotResponseDTO = {
  code?: string;
  data: ZFSMetricSnapshotDTO;
  message?: string;
  request_id?: string;
  success: boolean;
  timestamp: string;
  trace_id?: string;
};
