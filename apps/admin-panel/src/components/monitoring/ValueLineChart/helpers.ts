import {
  buildChartLegendItems,
  buildEvenlyDistributedIndexes,
  chartSeriesPalette,
  createPercentChartFrame,
  formatChartStamp,
  hasMultiSeriesChartData,
  mapPercentChartSeriesX,
} from "@components/monitoring/percent-chart-shared";
import type {
  ValueLineChartAxisLabel,
  ValueLineChartLegendItem,
  ValueLineChartLine,
  ValueLineChartYAxisLabel,
} from "./types";

const minValue = 0;
const yGridSegments = 5;
const valueLineChartEntriesPerLabel = 60;

/**
 * Static frame dimensions shared by the dynamic value line chart.
 */
export const valueLineChartFrame = createPercentChartFrame({
  bottomPadding: 34,
  chartHeight: 280,
  chartWidth: 720,
  leftPadding: 52,
  rightPadding: 20,
  topPadding: 12,
});

/**
 * Returns whether all series can be rendered against the same timestamp axis.
 */
export const hasValueLineChartData = (
  valuesByKey: Record<string, number[]>,
  stamps: string[],
): boolean => {
  return hasMultiSeriesChartData(valuesByKey, stamps);
};

/**
 * Builds the rendered line paths and colors for each value series.
 */
export const buildValueLineChartLines = (
  valuesByKey: Record<string, number[]>,
  capacity: number,
): ValueLineChartLine[] => {
  const maxValue = resolveValueLineChartMax(valuesByKey);

  return Object.entries(valuesByKey).map(([key, values], index) => {
    const color = chartSeriesPalette[index % chartSeriesPalette.length];

    return {
      color,
      key,
      path: buildValueLineChartLinePath(values, capacity, maxValue),
    };
  });
};

/**
 * Builds legend items that mirror the rendered series colors and latest values.
 */
export const buildValueLineChartLegendItems = (
  valuesByKey: Record<string, number[]>,
): ValueLineChartLegendItem[] => {
  return buildChartLegendItems(valuesByKey);
};

/**
 * Builds X-axis labels using the available timestamps while preserving the full chart capacity scale.
 */
export const buildValueLineChartAxisLabels = (
  stamps: string[],
  capacity: number,
): ValueLineChartAxisLabel[] => {
  if (stamps.length === 0) {
    return [];
  }

  const preferredLabelCount = resolveValueLineChartLabelCount(stamps.length);
  const labelIndexes = buildEvenlyDistributedIndexes(stamps.length, preferredLabelCount);

  return labelIndexes.flatMap((index) => {
    const stamp = stamps[index];

    if (stamp === undefined) {
      return [];
    }

    return [
      {
        label: formatChartStamp(stamp),
        x: mapPercentChartSeriesX(valueLineChartFrame, index, stamps.length, capacity),
      },
    ];
  });
};

/**
 * Builds Y-axis labels from zero to the current maximum value range.
 */
export const buildValueLineChartYAxisLabels = (
  valuesByKey: Record<string, number[]>,
  formatValue: (value: number) => string,
): ValueLineChartYAxisLabel[] => {
  const maxValue = resolveValueLineChartMax(valuesByKey);
  const roundedMaxValue = roundValueLineChartMax(maxValue);

  return Array.from({ length: yGridSegments + 1 }, (_, index) => {
    const value = roundedMaxValue * ((yGridSegments - index) / yGridSegments);

    return {
      label: formatValue(value),
      value,
      y: mapValueLineChartY(value, roundedMaxValue),
    };
  });
};

/**
 * Uses a sparser timestamp-label density than percent charts so compact value cards remain readable.
 */
const resolveValueLineChartLabelCount = (length: number): number => {
  return Math.min(length, Math.max(1, Math.ceil(length / valueLineChartEntriesPerLabel)));
};

/**
 * Default compact formatter used by dynamic-value chart labels.
 */
export const formatValueLineChartValue = (value: number): string => {
  return new Intl.NumberFormat(undefined, {
    maximumFractionDigits: value >= 10 ? 0 : 1,
    notation: value >= 1000 ? "compact" : "standard",
  }).format(value);
};

/**
 * Builds the SVG line path for one value series within the fixed-capacity frame.
 */
const buildValueLineChartLinePath = (
  values: number[],
  capacity: number,
  maxValue: number,
): string => {
  const roundedMaxValue = roundValueLineChartMax(maxValue);

  return values
    .map((value, index) => {
      const x = mapPercentChartSeriesX(valueLineChartFrame, index, values.length, capacity);
      const y = mapValueLineChartY(value, roundedMaxValue);
      const command = index === 0 ? "M" : "L";

      return `${command} ${x} ${y}`;
    })
    .join(" ");
};

/**
 * Resolves the highest numeric value currently visible across every rendered series.
 */
const resolveValueLineChartMax = (valuesByKey: Record<string, number[]>): number => {
  return Math.max(1, ...Object.values(valuesByKey).flat());
};

/**
 * Rounds the dynamic chart maximum up to a stable human-readable grid ceiling.
 */
const roundValueLineChartMax = (value: number): number => {
  if (value <= 1) {
    return 1;
  }

  const magnitude = 10 ** Math.floor(Math.log10(value));
  const normalizedValue = value / magnitude;

  if (normalizedValue <= 2) {
    return 2 * magnitude;
  }

  if (normalizedValue <= 5) {
    return 5 * magnitude;
  }

  return 10 * magnitude;
};

/**
 * Resolves the SVG Y coordinate for one value on the current rounded vertical domain.
 */
const mapValueLineChartY = (value: number, maxValue: number): number => {
  const boundedValue = Math.min(maxValue, Math.max(minValue, value));
  const percent = boundedValue / maxValue;

  return (
    valueLineChartFrame.topPadding +
    valueLineChartFrame.innerHeight -
    percent * valueLineChartFrame.innerHeight
  );
};
