/**
 * One legend row rendered for a colored value series.
 */
export type ValueLineChartLegendItem = {
  color: string;
  key: string;
  latestValue: number;
};

/**
 * One positioned X-axis label derived from the available chart timestamps.
 */
export type ValueLineChartAxisLabel = {
  label: string;
  x: number;
};

/**
 * One positioned Y-axis label derived from the current value-domain grid.
 */
export type ValueLineChartYAxisLabel = {
  label: string;
  value: number;
  y: number;
};

/**
 * One rendered chart line with its assigned stroke color.
 */
export type ValueLineChartLine = {
  color: string;
  key: string;
  path: string;
};

/**
 * Input contract accepted by the dynamic value line chart component.
 */
export type ValueLineChartProps = {
  /**
   * Maximum number of values represented by the fixed X scale.
   */
  capacity: number;
  /**
   * Optional formatter used by Y-axis and legend labels.
   */
  formatValue?: (value: number) => string;
  /**
   * Ordered timestamps associated with each plotted point.
   */
  stamps: string[];
  /**
   * Ordered value series plotted on a dynamic vertical scale.
   */
  valuesByKey: Record<string, number[]>;
};
