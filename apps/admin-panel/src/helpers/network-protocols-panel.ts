import type {
  NetworkMetricSnapshotDTO,
  NetworkMetricSocketIPCountDTO,
  NetworkMetricSocketPortCountDTO,
} from "@dto/monitoring/network-metric";
import type { MetricChartLabel, MetricMultiChartSeries } from "@helpers/system-metric-chart";

/**
 * One browser-facing local-port row rendered in the protocols detail table.
 */
export type NetworkProtocolsPanelPortRow = {
  app: string;
  connection: string;
  count: string;
};

/**
 * One browser-facing remote-IP row rendered in the protocols detail table.
 */
export type NetworkProtocolsPanelRemoteIPRow = {
  count: string;
  ip: string;
};

/**
 * Browser-facing data rendered by the protocols-and-sockets network telemetry panel.
 */
export type NetworkProtocolsPanelData = {
  portRows: NetworkProtocolsPanelPortRow[];
  protocolTotalsSeries: MetricMultiChartSeries;
  protocolTotalsLabels: MetricChartLabel[];
  remoteIPRows: NetworkProtocolsPanelRemoteIPRow[];
  socketStateLabels: MetricChartLabel[];
  summaryLabels: MetricChartLabel[];
};

const commonServiceNamesByProtocolAndPort: Record<string, Record<number, string>> = {
  tcp: {
    20: "FTP data",
    21: "FTP control",
    22: "SSH",
    25: "SMTP",
    53: "DNS",
    80: "HTTP",
    110: "POP3",
    111: "RPCbind",
    123: "NTP",
    135: "MS RPC",
    139: "NetBIOS session",
    143: "IMAP",
    161: "SNMP",
    389: "LDAP",
    443: "HTTPS",
    445: "SMB",
    465: "SMTPS",
    514: "Syslog",
    587: "Submission",
    631: "IPP",
    636: "LDAPS",
    873: "rsync",
    993: "IMAPS",
    995: "POP3S",
    1433: "MSSQL",
    1521: "Oracle DB",
    2049: "NFS",
    2375: "Docker",
    2376: "Docker TLS",
    3000: "Web dev server",
    3306: "MySQL",
    3389: "RDP",
    4222: "NATS",
    5000: "Registry / dev app",
    5173: "Vite dev server",
    5432: "PostgreSQL",
    5900: "VNC",
    5985: "WinRM HTTP",
    5986: "WinRM HTTPS",
    6379: "Redis",
    6443: "Kubernetes API",
    8000: "Web app",
    8080: "HTTP alt",
    8443: "HTTPS alt",
    9000: "App / admin",
    9090: "Metrics / admin",
    9100: "Node exporter",
    9200: "Elasticsearch",
    9418: "Git",
    10000: "Webmin",
    11211: "Memcached",
    27017: "MongoDB",
  },
  tcp6: {
    22: "SSH",
    25: "SMTP",
    53: "DNS",
    80: "HTTP",
    111: "RPCbind",
    443: "HTTPS",
    445: "SMB",
    631: "IPP",
    2049: "NFS",
    4222: "NATS",
    5173: "Vite dev server",
    5432: "PostgreSQL",
    8080: "HTTP alt",
    8443: "HTTPS alt",
    9090: "Metrics / admin",
  },
  udp: {
    53: "DNS",
    67: "DHCP server",
    68: "DHCP client",
    69: "TFTP",
    111: "RPCbind",
    123: "NTP",
    137: "NetBIOS name",
    138: "NetBIOS datagram",
    161: "SNMP",
    162: "SNMP trap",
    389: "LDAP",
    443: "QUIC / HTTPS",
    514: "Syslog",
    631: "IPP",
    1194: "OpenVPN",
    1900: "SSDP",
    2049: "NFS",
    3478: "STUN",
    4789: "VXLAN",
    5353: "mDNS",
    6443: "Kubernetes API",
    8472: "Flannel VXLAN",
  },
  udp6: {
    53: "DNS",
    123: "NTP",
    443: "QUIC / HTTPS",
    5353: "mDNS",
  },
};

/**
 * Builds the latest browser-facing protocol and socket panel data.
 */
export const buildNetworkProtocolsPanelData = (
  items: NetworkMetricSnapshotDTO[],
): NetworkProtocolsPanelData => {
  const latestSockets = items.at(-1)?.sockets;
  const stateCounts = latestSockets?.by_state ?? {};

  return {
    portRows: buildPortRows(latestSockets?.top_local_ports),
    protocolTotalsLabels: buildProtocolTotalsLabels(latestSockets),
    protocolTotalsSeries: buildProtocolTotalsSeries(items),
    remoteIPRows: buildRemoteIPRows(latestSockets?.top_remote_ips),
    socketStateLabels: buildSocketStateLabels(stateCounts),
    summaryLabels: buildSummaryLabels(latestSockets, stateCounts),
  };
};

/**
 * Builds human-readable totals for the latest protocol socket counts.
 */
const buildProtocolTotalsLabels = (
  latestSockets: NetworkMetricSnapshotDTO["sockets"] | null | undefined,
): MetricChartLabel[] => {
  return [
    { key: "All sockets", value: formatInteger(latestSockets?.total.all ?? 0) },
    { key: "TCP", value: formatInteger(latestSockets?.total.tcp ?? 0) },
    { key: "TCP6", value: formatInteger(latestSockets?.total.tcp6 ?? 0) },
    { key: "UDP", value: formatInteger(latestSockets?.total.udp ?? 0) },
    { key: "UDP6", value: formatInteger(latestSockets?.total.udp6 ?? 0) },
  ];
};

/**
 * Builds the protocol socket totals history chart from all snapshots.
 */
const buildProtocolTotalsSeries = (items: NetworkMetricSnapshotDTO[]): MetricMultiChartSeries => {
  return {
    stamps: items.map((item) => item.timestamp),
    valuesByKey: {
      TCP: items.map((item) => item.sockets.total.tcp),
      TCP6: items.map((item) => item.sockets.total.tcp6),
      UDP: items.map((item) => item.sockets.total.udp),
      UDP6: items.map((item) => item.sockets.total.udp6),
    },
  };
};

/**
 * Builds the latest connection-state rows.
 */
const buildSocketStateLabels = (stateCounts: Record<string, number>): MetricChartLabel[] => {
  return [
    { key: "Established", value: formatInteger(stateCounts.ESTABLISHED ?? 0) },
    { key: "Listen", value: formatInteger(stateCounts.LISTEN ?? 0) },
    { key: "Time wait", value: formatInteger(stateCounts.TIME_WAIT ?? 0) },
    { key: "Close wait", value: formatInteger(stateCounts.CLOSE_WAIT ?? 0) },
    { key: "Close", value: formatInteger(stateCounts.CLOSE ?? 0) },
  ];
};

/**
 * Builds the latest high-level summary rows.
 */
const buildSummaryLabels = (
  latestSockets: NetworkMetricSnapshotDTO["sockets"] | null | undefined,
  stateCounts: Record<string, number>,
): MetricChartLabel[] => {
  return [
    { key: "Sockets used", value: formatInteger(latestSockets?.sockstat.sockets_used ?? 0) },
    { key: "Active TCP", value: formatInteger(stateCounts.ESTABLISHED ?? 0) },
    { key: "Listening", value: formatInteger(stateCounts.LISTEN ?? 0) },
    { key: "Time wait", value: formatInteger(stateCounts.TIME_WAIT ?? 0) },
  ];
};

/**
 * Formats one integer counter for compact telemetry metadata rows.
 */
const formatInteger = (value: number): string => {
  return new Intl.NumberFormat().format(value);
};

/**
 * Builds one browser-facing local-port detail row with a best-effort common service name.
 */
const buildPortRow = (value: NetworkMetricSocketPortCountDTO): NetworkProtocolsPanelPortRow => {
  return {
    app: resolveCommonServiceName(value.protocol, value.port),
    connection: `${value.protocol.toUpperCase()} ${value.port}`,
    count: formatInteger(value.count),
  };
};

/**
 * Builds all browser-facing local-port rows.
 */
const buildPortRows = (
  topLocalPorts: NetworkMetricSocketPortCountDTO[] | null | undefined,
): NetworkProtocolsPanelPortRow[] => {
  return (topLocalPorts ?? []).map(buildPortRow);
};

/**
 * Builds one browser-facing remote-IP detail row.
 */
const buildRemoteIPRow = (
  value: NetworkMetricSocketIPCountDTO,
): NetworkProtocolsPanelRemoteIPRow => {
  return {
    count: formatInteger(value.count),
    ip: value.ip,
  };
};

/**
 * Builds all browser-facing remote-IP detail rows.
 */
const buildRemoteIPRows = (
  topRemoteIPs: NetworkMetricSocketIPCountDTO[] | null | undefined,
): NetworkProtocolsPanelRemoteIPRow[] => {
  return (topRemoteIPs ?? []).map(buildRemoteIPRow);
};

/**
 * Resolves a human-readable service name from one protocol and port when a common mapping exists.
 */
const resolveCommonServiceName = (protocol: string, port: number): string => {
  return commonServiceNamesByProtocolAndPort[protocol.toLowerCase()]?.[port] ?? "Unknown app";
};
