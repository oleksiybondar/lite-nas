/**
 * Input contract accepted by the multi-series percent gradient chart component.
 * The component is visual-only and does not own card titles or summaries.
 */
export type PercentGradientMultiChartProps = {
  /**
   * Maximum number of values represented by the fixed X scale.
   */
  capacity: number;
  /**
   * Ordered timestamps associated with each plotted point.
   */
  stamps: string[];
  /**
   * Ordered percent series plotted on a fixed 0-100 vertical scale.
   */
  valuesByKey: Record<string, number[]>;
};

/**
 * One legend row rendered for a colored series.
 */
export type PercentGradientMultiChartLegendItem = {
  color: string;
  key: string;
  latestValue: number;
};

/**
 * One positioned X-axis label derived from the available chart timestamps.
 */
export type PercentGradientMultiChartAxisLabel = {
  label: string;
  x: number;
};

/**
 * One rendered chart line with its assigned stroke color.
 */
export type PercentGradientMultiChartLine = {
  color: string;
  key: string;
  path: string;
};
