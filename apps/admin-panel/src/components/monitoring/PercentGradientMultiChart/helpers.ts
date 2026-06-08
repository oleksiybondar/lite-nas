import {
  buildEvenlyDistributedIndexes,
  clampPercentChartValue,
  createPercentChartFrame,
  mapPercentChartSeriesX,
  mapPercentChartY,
  resolvePercentChartLabelCount,
} from "@components/monitoring/percent-chart-shared";
import type {
  PercentGradientMultiChartAxisLabel,
  PercentGradientMultiChartLegendItem,
  PercentGradientMultiChartLine,
} from "./types";

const maxPercent = 100;
const minPercent = 0;
const seriesPalette = [
  "#2563eb",
  "#16a34a",
  "#ea580c",
  "#dc2626",
  "#7c3aed",
  "#0891b2",
  "#ca8a04",
  "#db2777",
] as const;
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
  const seriesEntries = Object.entries(valuesByKey);

  return (
    stamps.length > 0 &&
    seriesEntries.length > 0 &&
    seriesEntries.every(([, values]) => values.length === stamps.length && values.length > 0)
  );
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
      color: seriesPalette[index % seriesPalette.length],
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
  return Object.entries(valuesByKey).map(([key, values], index) => {
    return {
      color: seriesPalette[index % seriesPalette.length],
      key,
      latestValue: values[values.length - 1],
    };
  });
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
  if (stamps.length === 0) {
    return [];
  }

  const preferredLabelCount = resolvePercentChartLabelCount(stamps.length);
  const labelIndexes = buildEvenlyDistributedIndexes(stamps.length, preferredLabelCount);

  return labelIndexes.map((index) => {
    return {
      label: formatPercentGradientMultiChartStamp(stamps[index]),
      x: mapPercentChartSeriesX(percentGradientMultiChartFrame, index, stamps.length, capacity),
    };
  });
};

/**
 * Formats one timestamp for compact chart footer labels.
 */
export const formatPercentGradientMultiChartStamp = (value: string): string => {
  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return date.toLocaleTimeString([], {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
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
