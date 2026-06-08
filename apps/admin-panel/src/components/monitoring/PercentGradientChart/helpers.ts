import {
  buildEvenlyDistributedIndexes,
  clampPercentChartValue,
  createPercentChartFrame,
  mapPercentChartSeriesX,
  mapPercentChartY,
  resolvePercentChartLabelCount,
} from "@components/monitoring/percent-chart-shared";

const maxPercent = 100;
const minPercent = 0;

/**
 * One positioned X-axis label derived from the available chart timestamps.
 */
export type PercentGradientChartAxisLabel = {
  label: string;
  x: number;
};

/**
 * Static chart frame dimensions shared by the percent gradient chart.
 */
export const percentGradientChartFrame = createPercentChartFrame({
  bottomPadding: 34,
  chartHeight: 220,
  chartWidth: 720,
  leftPadding: 40,
  rightPadding: 20,
  topPadding: 12,
});

/**
 * Returns whether one percent chart has enough data points to draw a series.
 */
export const hasPercentGradientChartData = (values: number[], stamps: string[]): boolean => {
  return values.length > 0 && values.length === stamps.length;
};

/**
 * Clamps one input value to the fixed chart percent domain.
 */
export const clampPercentGradientChartValue = (value: number): number => {
  return clampPercentChartValue(value, minPercent, maxPercent);
};

/**
 * Builds the polyline path used by the percent gradient chart line overlay.
 */
export const buildPercentGradientChartLinePath = (values: number[], capacity: number): string => {
  return values
    .map((value, index) => {
      const x = mapPercentChartSeriesX(percentGradientChartFrame, index, values.length, capacity);
      const y = mapPercentGradientChartY(value);
      const command = index === 0 ? "M" : "L";

      return `${command} ${x} ${y}`;
    })
    .join(" ");
};

/**
 * Builds the closed area path rendered under the percent chart line.
 */
export const buildPercentGradientChartAreaPath = (values: number[], capacity: number): string => {
  if (values.length === 0) {
    return "";
  }

  const linePath = buildPercentGradientChartLinePath(values, capacity);
  const firstX = mapPercentChartSeriesX(percentGradientChartFrame, 0, values.length, capacity);
  const lastX = mapPercentChartSeriesX(
    percentGradientChartFrame,
    values.length - 1,
    values.length,
    capacity,
  );
  const baselineY = percentGradientChartFrame.topPadding + percentGradientChartFrame.innerHeight;

  return `${linePath} L ${lastX} ${baselineY} L ${firstX} ${baselineY} Z`;
};

/**
 * Builds the horizontal guide line coordinates for one percent marker.
 */
export const mapPercentGradientChartGuideY = (value: number): number => {
  return mapPercentGradientChartY(value);
};

/**
 * Builds X-axis labels using the available timestamps while preserving the full chart capacity scale.
 */
export const buildPercentGradientChartAxisLabels = (
  stamps: string[],
  capacity: number,
): PercentGradientChartAxisLabel[] => {
  if (stamps.length === 0) {
    return [];
  }

  const preferredLabelCount = resolvePercentChartLabelCount(stamps.length);
  const labelIndexes = buildEvenlyDistributedIndexes(stamps.length, preferredLabelCount);

  return labelIndexes.map((index) => {
    return {
      label: formatPercentGradientChartStamp(stamps[index]),
      x: mapPercentChartSeriesX(percentGradientChartFrame, index, stamps.length, capacity),
    };
  });
};

/**
 * Resolves the dashed guide Y coordinate for the latest plotted value.
 */
export const buildPercentGradientChartLatestGuideY = (values: number[]): number | null => {
  if (values.length === 0) {
    return null;
  }

  return mapPercentGradientChartY(values[values.length - 1]);
};

/**
 * Formats one timestamp for compact chart footer labels.
 */
export const formatPercentGradientChartStamp = (value: string): string => {
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
 * Resolves the SVG Y coordinate for one percent value on the fixed vertical domain.
 */
const mapPercentGradientChartY = (value: number): number => {
  return mapPercentChartY(percentGradientChartFrame, value, minPercent, maxPercent);
};
