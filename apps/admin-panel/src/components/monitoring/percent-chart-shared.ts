/**
 * Shared frame dimensions used by fixed-scale percent charts.
 */
export type PercentChartFrame = {
  bottomPadding: number;
  chartHeight: number;
  chartWidth: number;
  innerHeight: number;
  innerWidth: number;
  leftPadding: number;
  rightPadding: number;
  topPadding: number;
};

/**
 * Builds one fixed-scale percent chart frame from the supplied dimensions.
 */
export const createPercentChartFrame = ({
  bottomPadding,
  chartHeight,
  chartWidth,
  leftPadding,
  rightPadding,
  topPadding,
}: {
  bottomPadding: number;
  chartHeight: number;
  chartWidth: number;
  leftPadding: number;
  rightPadding: number;
  topPadding: number;
}): PercentChartFrame => {
  return {
    bottomPadding,
    chartHeight,
    chartWidth,
    innerHeight: chartHeight - topPadding - bottomPadding,
    innerWidth: chartWidth - leftPadding - rightPadding,
    leftPadding,
    rightPadding,
    topPadding,
  };
};

/**
 * Clamps one input value to the provided fixed chart percent domain.
 */
export const clampPercentChartValue = (
  value: number,
  minPercent: number,
  maxPercent: number,
): number => {
  return Math.min(maxPercent, Math.max(minPercent, value));
};

/**
 * Resolves the SVG X coordinate for one visible series index on the fixed-capacity scale.
 */
export const mapPercentChartSeriesX = (
  frame: PercentChartFrame,
  index: number,
  length: number,
  capacity: number,
): number => {
  const safeCapacity = Math.max(1, capacity);
  const boundedLength = Math.min(length, safeCapacity);
  const slotIndex = safeCapacity - boundedLength + index;

  return mapPercentChartSlotX(frame, slotIndex, safeCapacity);
};

/**
 * Resolves the SVG X coordinate for one capacity slot on the supplied chart frame.
 */
export const mapPercentChartSlotX = (
  frame: PercentChartFrame,
  slotIndex: number,
  capacity: number,
): number => {
  if (capacity <= 1) {
    return frame.leftPadding + frame.innerWidth / 2;
  }

  return frame.leftPadding + (slotIndex / (capacity - 1)) * frame.innerWidth;
};

/**
 * Resolves the SVG Y coordinate for one percent value on the provided frame and domain.
 */
export const mapPercentChartY = (
  frame: PercentChartFrame,
  value: number,
  minPercent: number,
  maxPercent: number,
): number => {
  const safeValue = clampPercentChartValue(value, minPercent, maxPercent);
  const percent = safeValue / maxPercent;

  return frame.topPadding + frame.innerHeight - percent * frame.innerHeight;
};

/**
 * Resolves how many timestamp labels should be shown for the current series length.
 */
export const resolvePercentChartLabelCount = (length: number): number => {
  return Math.min(length, Math.max(1, Math.ceil(length / 30)));
};

/**
 * Picks evenly spaced source indexes used to label the fixed-capacity X axis.
 */
export const buildEvenlyDistributedIndexes = (length: number, count: number): number[] => {
  if (count <= 1) {
    return [length - 1];
  }

  return Array.from({ length: count }, (_, index) => {
    return Math.round((index / (count - 1)) * (length - 1));
  }).filter((value, index, values) => {
    return index === 0 || value !== values[index - 1];
  });
};
