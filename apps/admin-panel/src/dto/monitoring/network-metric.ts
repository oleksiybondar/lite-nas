/**
 * One named network counter group returned by the transport.
 */
export type NetworkMetricCounterGroupDTO = Record<string, number>;

/**
 * Stable adapter identity metadata returned for one network interface.
 */
export type NetworkMetricInterfaceAdapterDTO = {
  description?: string | null;
  device_id: string;
  device_name?: string | null;
  vendor_id: string;
  vendor_name?: string | null;
};

/**
 * Factual per-interface counters returned by the transport.
 */
export type NetworkMetricInterfaceStatisticsDTO = {
  rx_bytes: number;
  rx_compressed: number;
  rx_crc_errors: number;
  rx_dropped: number;
  rx_errors: number;
  rx_fifo_errors: number;
  rx_frame_errors: number;
  rx_length_errors: number;
  rx_missed_errors: number;
  rx_multicast: number;
  rx_nohandler: number;
  rx_over_errors: number;
  rx_packets: number;
  tx_aborted_errors: number;
  tx_bytes: number;
  tx_carrier_errors: number;
  tx_collisions: number;
  tx_compressed: number;
  tx_dropped: number;
  tx_errors: number;
  tx_fifo_errors: number;
  tx_heartbeat_errors: number;
  tx_packets: number;
  tx_window_errors: number;
};

/**
 * One host network interface snapshot item.
 */
export type NetworkMetricInterfaceSnapshotDTO = {
  adapter?: NetworkMetricInterfaceAdapterDTO | null;
  address: string;
  bus?: string | null;
  carrier_up?: boolean | null;
  duplex: string;
  ifindex?: number | null;
  kind: string;
  mtu?: number | null;
  name: string;
  oper_state: string;
  speed_mbps?: number | null;
  statistics: NetworkMetricInterfaceStatisticsDTO;
  tx_queue_len?: number | null;
};

/**
 * Grouped raw protocol counters returned by the transport.
 */
export type NetworkMetricProtocolSnapshotDTO = {
  icmp: NetworkMetricCounterGroupDTO;
  ip: NetworkMetricCounterGroupDTO;
  ip_ext: NetworkMetricCounterGroupDTO;
  tcp: NetworkMetricCounterGroupDTO;
  tcp_ext: NetworkMetricCounterGroupDTO;
  udp: NetworkMetricCounterGroupDTO;
  udplite: NetworkMetricCounterGroupDTO;
};

/**
 * One summarized local-port count item.
 */
export type NetworkMetricSocketPortCountDTO = {
  count: number;
  port: number;
  protocol: string;
};

/**
 * One summarized remote-IP count item.
 */
export type NetworkMetricSocketIPCountDTO = {
  count: number;
  ip: string;
};

/**
 * Aggregate socket totals returned by the transport.
 */
export type NetworkMetricSocketTotalsDTO = {
  all: number;
  tcp: number;
  tcp6: number;
  udp: number;
  udp6: number;
};

/**
 * Factual `/proc/net/sockstat` totals returned by the transport.
 */
export type NetworkMetricSockStatDTO = {
  frag_inuse: number;
  frag_memory: number;
  raw_inuse: number;
  sockets_used: number;
  tcp_alloc: number;
  tcp_inuse: number;
  tcp_mem: number;
  tcp_orphan: number;
  tcp_time_wait: number;
  udp_inuse: number;
  udp_mem: number;
  udplite_inuse: number;
};

/**
 * Summarized socket metrics returned by the transport.
 */
export type NetworkMetricSocketSnapshotDTO = {
  by_state: Record<string, number>;
  sockstat: NetworkMetricSockStatDTO;
  top_local_ports: NetworkMetricSocketPortCountDTO[] | null;
  top_remote_ips: NetworkMetricSocketIPCountDTO[] | null;
  total: NetworkMetricSocketTotalsDTO;
};

/**
 * Summarized network-related softirq counters returned by the transport.
 */
export type NetworkMetricSoftIRQSnapshotDTO = {
  net_rx_per_cpu: number[] | null;
  net_rx_total: number;
  net_tx_per_cpu: number[] | null;
  net_tx_total: number;
};

/**
 * Kernel-side pressure indicators returned by the transport.
 */
export type NetworkMetricKernelPressureSnapshotDTO = {
  softirqs: NetworkMetricSoftIRQSnapshotDTO;
};

/**
 * One timestamped network metrics snapshot item.
 */
export type NetworkMetricSnapshotDTO = {
  interfaces: NetworkMetricInterfaceSnapshotDTO[] | null;
  kernel_pressure: NetworkMetricKernelPressureSnapshotDTO;
  protocols: NetworkMetricProtocolSnapshotDTO;
  sockets: NetworkMetricSocketSnapshotDTO;
  timestamp: string;
};

/**
 * Response envelope returned by network metrics history endpoints.
 */
export type NetworkMetricHistoryResponseDTO = {
  code?: string;
  data: NetworkMetricSnapshotDTO[] | null;
  message?: string;
  request_id?: string;
  success: boolean;
  timestamp: string;
  trace_id?: string;
};

/**
 * Response envelope returned by network metrics snapshot endpoints.
 */
export type NetworkMetricSnapshotResponseDTO = {
  code?: string;
  data: NetworkMetricSnapshotDTO;
  message?: string;
  request_id?: string;
  success: boolean;
  timestamp: string;
  trace_id?: string;
};
