import type { SystemMetricsCardData } from "@helpers/system-metric-chart";

export type SystemMetricsCardProps = {
  /**
   * Maximum number of metric records that can be displayed by the charts.
   */
  capacity: number;
  /**
   * Unified system metrics data rendered by the card.
   */
  metrics: SystemMetricsCardData;
};
