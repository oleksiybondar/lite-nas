import type {
  NetworkMetricHistoryResponseDTO,
  NetworkMetricSnapshotDTO,
  NetworkMetricSnapshotResponseDTO,
} from "@dto/monitoring/network-metric";
import { buildMonitoringResponseEnvelopeSchema } from "@schemas/monitoring/response-envelope";
import { z } from "zod";

/**
 * Runtime schema for one named network counter group.
 */
export const networkMetricCounterGroupSchema = z.record(z.string(), z.number());

/**
 * Runtime schema for one interface adapter identity block.
 */
export const networkMetricInterfaceAdapterSchema = z.object({
  description: z.string().nullable().optional(),
  device_id: z.string(),
  device_name: z.string().nullable().optional(),
  vendor_id: z.string(),
  vendor_name: z.string().nullable().optional(),
});

/**
 * Runtime schema for one interface statistics block.
 */
export const networkMetricInterfaceStatisticsSchema = z.object({
  rx_bytes: z.number(),
  rx_compressed: z.number(),
  rx_crc_errors: z.number(),
  rx_dropped: z.number(),
  rx_errors: z.number(),
  rx_fifo_errors: z.number(),
  rx_frame_errors: z.number(),
  rx_length_errors: z.number(),
  rx_missed_errors: z.number(),
  rx_multicast: z.number(),
  rx_nohandler: z.number(),
  rx_over_errors: z.number(),
  rx_packets: z.number(),
  tx_aborted_errors: z.number(),
  tx_bytes: z.number(),
  tx_carrier_errors: z.number(),
  tx_collisions: z.number(),
  tx_compressed: z.number(),
  tx_dropped: z.number(),
  tx_errors: z.number(),
  tx_fifo_errors: z.number(),
  tx_heartbeat_errors: z.number(),
  tx_packets: z.number(),
  tx_window_errors: z.number(),
});

/**
 * Runtime schema for one network interface snapshot item.
 */
export const networkMetricInterfaceSnapshotSchema = z.object({
  adapter: networkMetricInterfaceAdapterSchema.nullable().optional(),
  address: z.string(),
  bus: z.string().nullable().optional(),
  carrier_up: z.boolean().nullable().optional(),
  duplex: z.string(),
  ifindex: z.number().nullable().optional(),
  kind: z.string(),
  mtu: z.number().nullable().optional(),
  name: z.string(),
  oper_state: z.string(),
  speed_mbps: z.number().nullable().optional(),
  statistics: networkMetricInterfaceStatisticsSchema,
  tx_queue_len: z.number().nullable().optional(),
});

/**
 * Runtime schema for grouped protocol counters.
 */
export const networkMetricProtocolSnapshotSchema = z.object({
  icmp: networkMetricCounterGroupSchema,
  ip: networkMetricCounterGroupSchema,
  ip_ext: networkMetricCounterGroupSchema,
  tcp: networkMetricCounterGroupSchema,
  tcp_ext: networkMetricCounterGroupSchema,
  udp: networkMetricCounterGroupSchema,
  udplite: networkMetricCounterGroupSchema,
});

/**
 * Runtime schema for one summarized local-port count item.
 */
export const networkMetricSocketPortCountSchema = z.object({
  count: z.number(),
  port: z.number().int().nonnegative(),
  protocol: z.string(),
});

/**
 * Runtime schema for one summarized remote-IP count item.
 */
export const networkMetricSocketIPCountSchema = z.object({
  count: z.number(),
  ip: z.string(),
});

/**
 * Runtime schema for aggregate socket totals.
 */
export const networkMetricSocketTotalsSchema = z.object({
  all: z.number(),
  tcp: z.number(),
  tcp6: z.number(),
  udp: z.number(),
  udp6: z.number(),
});

/**
 * Runtime schema for `/proc/net/sockstat` totals.
 */
export const networkMetricSockStatSchema = z.object({
  frag_inuse: z.number(),
  frag_memory: z.number(),
  raw_inuse: z.number(),
  sockets_used: z.number(),
  tcp_alloc: z.number(),
  tcp_inuse: z.number(),
  tcp_mem: z.number(),
  tcp_orphan: z.number(),
  tcp_time_wait: z.number(),
  udp_inuse: z.number(),
  udp_mem: z.number(),
  udplite_inuse: z.number(),
});

/**
 * Runtime schema for summarized socket metrics.
 */
export const networkMetricSocketSnapshotSchema = z.object({
  by_state: z.record(z.string(), z.number()),
  sockstat: networkMetricSockStatSchema,
  top_local_ports: z.array(networkMetricSocketPortCountSchema).nullable(),
  top_remote_ips: z.array(networkMetricSocketIPCountSchema).nullable(),
  total: networkMetricSocketTotalsSchema,
});

/**
 * Runtime schema for summarized softirq counters.
 */
export const networkMetricSoftIRQSnapshotSchema = z.object({
  net_rx_per_cpu: z.array(z.number()).nullable(),
  net_rx_total: z.number(),
  net_tx_per_cpu: z.array(z.number()).nullable(),
  net_tx_total: z.number(),
});

/**
 * Runtime schema for kernel-side pressure indicators.
 */
export const networkMetricKernelPressureSnapshotSchema = z.object({
  softirqs: networkMetricSoftIRQSnapshotSchema,
});

/**
 * Runtime schema for one timestamped network metrics snapshot item.
 */
export const networkMetricSnapshotSchema = z.object({
  interfaces: z.array(networkMetricInterfaceSnapshotSchema).nullable(),
  kernel_pressure: networkMetricKernelPressureSnapshotSchema,
  protocols: networkMetricProtocolSnapshotSchema,
  sockets: networkMetricSocketSnapshotSchema,
  timestamp: z.string(),
});

/**
 * Runtime schema for a network metrics history response envelope.
 */
export const networkMetricHistoryResponseSchema = buildMonitoringResponseEnvelopeSchema(
  z.array(networkMetricSnapshotSchema).nullable(),
);

/**
 * Runtime schema for a network metrics snapshot response envelope.
 */
export const networkMetricSnapshotResponseSchema = buildMonitoringResponseEnvelopeSchema(
  networkMetricSnapshotSchema,
);

/**
 * Parses a network metrics history transport response into browser-facing items.
 */
export const parseNetworkMetricHistoryResponse = (value: unknown): NetworkMetricSnapshotDTO[] => {
  const response = networkMetricHistoryResponseSchema.parse(
    value,
  ) as NetworkMetricHistoryResponseDTO;

  return response.data ?? [];
};

/**
 * Parses a network metrics snapshot transport response into one browser-facing item.
 */
export const parseNetworkMetricSnapshotResponse = (value: unknown): NetworkMetricSnapshotDTO => {
  const response = networkMetricSnapshotResponseSchema.parse(
    value,
  ) as NetworkMetricSnapshotResponseDTO;

  return response.data;
};
