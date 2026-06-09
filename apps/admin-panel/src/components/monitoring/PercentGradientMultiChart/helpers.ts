import {
  buildChartLegendItems,
  buildPercentChartAxisLabels,
  chartSeriesPalette,
  clampPercentChartValue,
  createPercentChartFrame,
  hasMultiSeriesChartData,
  mapPercentChartSeriesX,
  mapPercentChartY,
} from "@components/monitoring/percent-chart-shared";
import type {
  PercentGradientMultiChartAxisLabel,
  PercentGradientMultiChartLegendItem,
  PercentGradientMultiChartLine,
} from "./types";

const maxPercent = 100;
const minPercent = 0;
const gridStepPercent = 10;

/**
 * Static chart frame dimensions shared by the multi-series percent gradient chart.
 */
export const percentGradientMultiChartFrame = createPercentChartFrame({
  bottomPadding: 34,
  chartHeight: 280,
  chartWidth: 720,
  leftPadding: 40,
  rightPadding: 20,
  topPadding: 12,
});

/**
 * Fixed Y-axis markers rendered for the multi-series percent chart.
 */
export const percentGradientMultiChartGridValues = Array.from(
  { length: maxPercent / gridStepPercent + 1 },
  (_, index) => index * gridStepPercent,
);

/**
 * Returns whether all series can be rendered against the same timestamp axis.
 */
export const hasPercentGradientMultiChartData = (
  valuesByKey: Record<string, number[]>,
  stamps: string[],
): boolean => {
  return hasMultiSeriesChartData(valuesByKey, stamps);
};

/**
 * Clamps one input value to the fixed chart percent domain.
 */
export const clampPercentGradientMultiChartValue = (value: number): number => {
  return clampPercentChartValue(value, minPercent, maxPercent);
};

/**
 * Builds the rendered line paths and colors for each percent series.
 */
export const buildPercentGradientMultiChartLines = (
  valuesByKey: Record<string, number[]>,
  capacity: number,
): PercentGradientMultiChartLine[] => {
  return Object.entries(valuesByKey).map(([key, values], index) => {
    return {
      color: chartSeriesPalette[index % chartSeriesPalette.length] ?? "#000",
      key,
      path: buildPercentGradientMultiChartLinePath(values, capacity),
    };
  });
};

/**
 * Builds legend items that mirror the rendered series colors and latest values.
 */
export const buildPercentGradientMultiChartLegendItems = (
  valuesByKey: Record<string, number[]>,
): PercentGradientMultiChartLegendItem[] => {
  return buildChartLegendItems(valuesByKey);
};

/**
 * Builds the horizontal guide line coordinates for one percent marker.
 */
export const mapPercentGradientMultiChartGuideY = (value: number): number => {
  return mapPercentGradientMultiChartY(value);
};

/**
 * Builds X-axis labels using the available timestamps while preserving the full chart capacity scale.
 */
export const buildPercentGradientMultiChartAxisLabels = (
  stamps: string[],
  capacity: number,
): PercentGradientMultiChartAxisLabel[] => {
  return buildPercentChartAxisLabels(percentGradientMultiChartFrame, stamps, capacity);
};

/**
 * Formats one fixed-scale percent value for multi-series legends and hover tooltips.
 */
export const formatPercentGradientMultiChartValue = (value: number): string => {
  return `${new Intl.NumberFormat(undefined, {
    maximumFractionDigits: value >= 10 ? 0 : 1,
  }).format(clampPercentGradientMultiChartValue(value))}%`;
};

/**
 * Builds the SVG line path for one percent series within the fixed-capacity frame.
 */
const buildPercentGradientMultiChartLinePath = (values: number[], capacity: number): string => {
  return values
    .map((value, index) => {
      const x = mapPercentChartSeriesX(
        percentGradientMultiChartFrame,
        index,
        values.length,
        capacity,
      );
      const y = mapPercentGradientMultiChartY(value);
      const command = index === 0 ? "M" : "L";

      return `${command} ${x} ${y}`;
    })
    .join(" ");
};

/**
 * Resolves the SVG Y coordinate for one percent value on the fixed vertical domain.
 */
const mapPercentGradientMultiChartY = (value: number): number => {
  return mapPercentChartY(percentGradientMultiChartFrame, value, minPercent, maxPercent);
};
