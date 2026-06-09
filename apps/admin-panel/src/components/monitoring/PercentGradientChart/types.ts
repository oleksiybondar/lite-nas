/**
 * Input contract accepted by the percent gradient chart component.
 * The component is visual-only and does not own card titles or summaries.
 */
export type PercentGradientChartProps = {
  /**
   * Maximum number of values represented by the fixed X scale.
   */
  capacity: number;
  /**
   * Optional rendered chart height in pixels.
   */
  heightPx?: number | undefined;
  /**
   * Ordered timestamps associated with each plotted value.
   */
  stamps: string[];
  /**
   * Ordered percent values plotted on a fixed 0-100 vertical scale.
   */
  values: number[];
};
