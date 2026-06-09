/**
 * One formatted series row rendered inside a monitoring chart hover tooltip.
 */
export type MonitoringChartTooltipItem = {
  color?: string;
  label: string;
  value: string;
};

/**
 * Input contract accepted by the shared monitoring chart hover tooltip component.
 */
export type MonitoringChartTooltipProps = {
  /**
   * Fixed SVG chart width used to resolve the tooltip anchor side.
   */
  chartWidth: number;
  /**
   * Formatted series rows rendered under the timestamp label.
   */
  items: MonitoringChartTooltipItem[];
  /**
   * Timestamp or X-axis label currently hovered by the user.
   */
  label: string;
  /**
   * Hovered SVG X coordinate used to place the tooltip horizontally.
   */
  x: number;
};
