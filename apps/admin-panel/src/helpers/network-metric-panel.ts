import type {
  NetworkMetricInterfaceAdapterDTO,
  NetworkMetricInterfaceSnapshotDTO,
  NetworkMetricSnapshotDTO,
} from "@dto/monitoring/network-metric";
import { formatMetricBytes } from "@helpers/metric-display";
import type { MetricChartLabel, MetricMultiChartSeries } from "@helpers/system-metric-chart";

/**
 * One browser-facing network interface card rendered in the interfaces panel.
 */
export type NetworkInterfacePanelItem = {
  adapterLabel: string;
  errorAndDropSeries: MetricMultiChartSeries;
  linkDetailsLabel: string;
  name: string;
  packetRateSeries: MetricMultiChartSeries;
  status: string;
  throughputSeries: MetricMultiChartSeries;
  totalsLabels: MetricChartLabel[];
};

/**
 * Builds the latest interface cards rendered by the interfaces panel.
 */
export const buildNetworkInterfacePanelItems = (
  items: NetworkMetricSnapshotDTO[],
): NetworkInterfacePanelItem[] => {
  const latestItem = items.at(-1) ?? null;
  const latestInterfaces = latestItem?.interfaces ?? [];

  return latestInterfaces.map((item) => {
    return buildNetworkInterfacePanelItem(items, item);
  });
};

/**
 * Builds one browser-facing interface card from the full network metrics history.
 */
const buildNetworkInterfacePanelItem = (
  items: NetworkMetricSnapshotDTO[],
  latestInterface: NetworkMetricInterfaceSnapshotDTO,
): NetworkInterfacePanelItem => {
  const interfaceHistory = items.map((item) => {
    return (
      item.interfaces?.find((networkInterface) => networkInterface.name === latestInterface.name) ??
      null
    );
  });

  return {
    adapterLabel: resolveInterfaceAdapterLabel(latestInterface.adapter),
    errorAndDropSeries: buildInterfaceErrorAndDropSeries(items, interfaceHistory),
    linkDetailsLabel: resolveInterfaceLinkDetailsLabel(latestInterface),
    name: latestInterface.name,
    packetRateSeries: buildInterfacePacketRateSeries(items, interfaceHistory),
    status: latestInterface.oper_state,
    throughputSeries: buildInterfaceThroughputSeries(items, interfaceHistory),
    totalsLabels: [
      { key: "Total RX", value: formatMetricBytes(latestInterface.statistics.rx_bytes) },
      { key: "Total TX", value: formatMetricBytes(latestInterface.statistics.tx_bytes) },
    ],
  };
};

/**
 * Builds one RX/TX throughput history for the interface card.
 */
const buildInterfaceThroughputSeries = (
  items: NetworkMetricSnapshotDTO[],
  interfaceHistory: Array<NetworkMetricInterfaceSnapshotDTO | null>,
): MetricMultiChartSeries => {
  return {
    stamps: items.map((item) => item.timestamp),
    valuesByKey: {
      RX: interfaceHistory.map((item, index) => {
        return calculateInterfaceRate({
          currentItem: item,
          currentTimestamp: items[index]?.timestamp ?? null,
          metric: "rx_bytes",
          previousItem: interfaceHistory[index - 1] ?? null,
          previousTimestamp: items[index - 1]?.timestamp ?? null,
        });
      }),
      TX: interfaceHistory.map((item, index) => {
        return calculateInterfaceRate({
          currentItem: item,
          currentTimestamp: items[index]?.timestamp ?? null,
          metric: "tx_bytes",
          previousItem: interfaceHistory[index - 1] ?? null,
          previousTimestamp: items[index - 1]?.timestamp ?? null,
        });
      }),
    },
  };
};

/**
 * Builds one RX/TX packet-rate history for the interface card.
 */
const buildInterfacePacketRateSeries = (
  items: NetworkMetricSnapshotDTO[],
  interfaceHistory: Array<NetworkMetricInterfaceSnapshotDTO | null>,
): MetricMultiChartSeries => {
  return {
    stamps: items.map((item) => item.timestamp),
    valuesByKey: {
      "RX packets": interfaceHistory.map((item, index) => {
        return calculateInterfaceRate({
          currentItem: item,
          currentTimestamp: items[index]?.timestamp ?? null,
          metric: "rx_packets",
          previousItem: interfaceHistory[index - 1] ?? null,
          previousTimestamp: items[index - 1]?.timestamp ?? null,
        });
      }),
      "TX packets": interfaceHistory.map((item, index) => {
        return calculateInterfaceRate({
          currentItem: item,
          currentTimestamp: items[index]?.timestamp ?? null,
          metric: "tx_packets",
          previousItem: interfaceHistory[index - 1] ?? null,
          previousTimestamp: items[index - 1]?.timestamp ?? null,
        });
      }),
    },
  };
};

/**
 * Builds one RX/TX error-and-drop history for the interface card.
 */
const buildInterfaceErrorAndDropSeries = (
  items: NetworkMetricSnapshotDTO[],
  interfaceHistory: Array<NetworkMetricInterfaceSnapshotDTO | null>,
): MetricMultiChartSeries => {
  return {
    stamps: items.map((item) => item.timestamp),
    valuesByKey: {
      "RX errors+drops": interfaceHistory.map((item, index) => {
        return calculateCombinedInterfaceRate({
          currentItem: item,
          currentTimestamp: items[index]?.timestamp ?? null,
          metrics: ["rx_errors", "rx_dropped"],
          previousItem: interfaceHistory[index - 1] ?? null,
          previousTimestamp: items[index - 1]?.timestamp ?? null,
        });
      }),
      "TX errors+drops": interfaceHistory.map((item, index) => {
        return calculateCombinedInterfaceRate({
          currentItem: item,
          currentTimestamp: items[index]?.timestamp ?? null,
          metrics: ["tx_errors", "tx_dropped"],
          previousItem: interfaceHistory[index - 1] ?? null,
          previousTimestamp: items[index - 1]?.timestamp ?? null,
        });
      }),
    },
  };
};

type NetworkStatisticMetric = keyof NetworkMetricInterfaceSnapshotDTO["statistics"];

type CalculateInterfaceRateOptions = {
  currentItem: NetworkMetricInterfaceSnapshotDTO | null;
  currentTimestamp: string | null;
  metric: NetworkStatisticMetric;
  previousItem: NetworkMetricInterfaceSnapshotDTO | null;
  previousTimestamp: string | null;
};

/**
 * Calculates one per-second interface metric value from cumulative counters.
 */
const calculateInterfaceRate = ({
  currentItem,
  currentTimestamp,
  metric,
  previousItem,
  previousTimestamp,
}: CalculateInterfaceRateOptions): number => {
  const durationSeconds = resolveDurationSeconds(currentTimestamp, previousTimestamp);

  if (currentItem === null || previousItem === null || durationSeconds === null) {
    return 0;
  }

  const deltaValue = currentItem.statistics[metric] - previousItem.statistics[metric];

  if (deltaValue <= 0) {
    return 0;
  }

  return deltaValue / durationSeconds;
};

type CalculateCombinedInterfaceRateOptions = {
  currentItem: NetworkMetricInterfaceSnapshotDTO | null;
  currentTimestamp: string | null;
  metrics: NetworkStatisticMetric[];
  previousItem: NetworkMetricInterfaceSnapshotDTO | null;
  previousTimestamp: string | null;
};

/**
 * Calculates one combined per-second interface metric from multiple cumulative counters.
 */
const calculateCombinedInterfaceRate = ({
  currentItem,
  currentTimestamp,
  metrics,
  previousItem,
  previousTimestamp,
}: CalculateCombinedInterfaceRateOptions): number => {
  const durationSeconds = resolveDurationSeconds(currentTimestamp, previousTimestamp);

  if (currentItem === null || previousItem === null || durationSeconds === null) {
    return 0;
  }

  const deltaValue = metrics.reduce((total, metric) => {
    return total + (currentItem.statistics[metric] - previousItem.statistics[metric]);
  }, 0);

  if (deltaValue <= 0) {
    return 0;
  }

  return deltaValue / durationSeconds;
};

/**
 * Resolves one valid sample duration in seconds for rate calculations.
 */
const resolveDurationSeconds = (
  currentTimestamp: string | null,
  previousTimestamp: string | null,
): number | null => {
  if (currentTimestamp === null || previousTimestamp === null) {
    return null;
  }

  const currentTimeMs = Date.parse(currentTimestamp);
  const previousTimeMs = Date.parse(previousTimestamp);

  if (
    Number.isNaN(currentTimeMs) ||
    Number.isNaN(previousTimeMs) ||
    currentTimeMs <= previousTimeMs
  ) {
    return null;
  }

  return (currentTimeMs - previousTimeMs) / 1000;
};

/**
 * Resolves one human-readable adapter line for the interface card body.
 */
const resolveInterfaceAdapterLabel = (
  adapter: NetworkMetricInterfaceAdapterDTO | null | undefined,
): string => {
  const vendorAndDevice = [adapter?.vendor_name, adapter?.device_name]
    .filter((value): value is string => typeof value === "string" && value.trim().length > 0)
    .join(" ")
    .trim();

  if (vendorAndDevice.length > 0) {
    return vendorAndDevice;
  }

  if (adapter?.description && adapter.description.trim().length > 0) {
    return adapter.description;
  }

  return "Adapter details unavailable";
};

/**
 * Formats one raw integer without compact notation for metadata rows.
 */
const formatRawMetricInteger = (value: number): string => {
  return new Intl.NumberFormat(undefined, { maximumFractionDigits: 0 }).format(value);
};

/**
 * Resolves one compact link-details line rendered above the total RX/TX row.
 */
const resolveInterfaceLinkDetailsLabel = (
  networkInterface: NetworkMetricInterfaceSnapshotDTO,
): string => {
  const details = [
    networkInterface.kind,
    networkInterface.speed_mbps
      ? `${formatRawMetricInteger(networkInterface.speed_mbps)} Mbps`
      : null,
    networkInterface.duplex,
    networkInterface.mtu ? `MTU ${formatRawMetricInteger(networkInterface.mtu)}` : null,
  ].filter((value): value is string => typeof value === "string" && value.trim().length > 0);

  if (details.length === 0) {
    return "Link details unavailable";
  }

  return details.join(" | ");
};
