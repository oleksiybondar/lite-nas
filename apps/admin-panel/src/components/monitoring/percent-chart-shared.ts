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
 * One hovered point resolved from a fixed-capacity chart's shared X scale.
 */
export type ChartHoverPoint = {
  index: number;
  x: number;
};

/**
 * Input contract used to resolve one hovered point on a fixed-capacity chart.
 */
export type ResolveHoveredChartPointInput = {
  capacity: number;
  clientX: number;
  frame: PercentChartFrame;
  length: number;
  rectLeft: number;
  rectWidth: number;
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
  const visibleRange = resolveVisibleChartSlotRange(length, capacity);
  const slotIndex = visibleRange.firstVisibleSlot + index;

  return mapPercentChartSlotX(frame, slotIndex, visibleRange.capacity);
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
 * Resolves one hovered series point from the pointer's client X position on a fixed-capacity chart.
 */
export const resolveHoveredChartPointFromClientX = ({
  capacity,
  clientX,
  frame,
  length,
  rectLeft,
  rectWidth,
}: ResolveHoveredChartPointInput): ChartHoverPoint | null => {
  if (!hasHoverableChartGeometry(length, rectWidth)) {
    return null;
  }

  const localFrameX = ((clientX - rectLeft) / rectWidth) * frame.chartWidth;

  if (!isHoverWithinChartBody(frame, localFrameX)) {
    return null;
  }

  const visibleRange = resolveVisibleChartSlotRange(length, capacity);
  const approximateSlotIndex = Math.round(
    ((localFrameX - frame.leftPadding) / frame.innerWidth) * (visibleRange.capacity - 1),
  );

  if (!isHoverWithinVisibleSlots(approximateSlotIndex, visibleRange)) {
    return null;
  }

  const index = approximateSlotIndex - visibleRange.firstVisibleSlot;

  if (!isHoverIndexVisible(index, visibleRange.length)) {
    return null;
  }

  return {
    index,
    x: mapPercentChartSeriesX(frame, index, length, capacity),
  };
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

/**
 * Returns whether the current chart has usable geometry for hover interactions.
 */
const hasHoverableChartGeometry = (length: number, rectWidth: number): boolean => {
  return length > 0 && rectWidth > 0;
};

/**
 * Returns whether the pointer is still inside the drawable chart body on the shared frame.
 */
const isHoverWithinChartBody = (frame: PercentChartFrame, localFrameX: number): boolean => {
  return localFrameX >= frame.leftPadding && localFrameX <= frame.chartWidth - frame.rightPadding;
};

/**
 * Resolves the visible slot range occupied by the current chart data within the fixed capacity.
 */
const resolveVisibleChartSlotRange = (
  length: number,
  capacity: number,
): {
  capacity: number;
  firstVisibleSlot: number;
  lastVisibleSlot: number;
  length: number;
} => {
  const safeCapacity = Math.max(1, capacity);
  const boundedLength = Math.min(length, safeCapacity);
  const firstVisibleSlot = safeCapacity - boundedLength;

  return {
    capacity: safeCapacity,
    firstVisibleSlot,
    lastVisibleSlot: safeCapacity - 1,
    length: boundedLength,
  };
};

/**
 * Returns whether the hovered slot still intersects the visible data window on the fixed-capacity scale.
 */
const isHoverWithinVisibleSlots = (
  slotIndex: number,
  visibleRange: {
    firstVisibleSlot: number;
    lastVisibleSlot: number;
  },
): boolean => {
  return slotIndex >= visibleRange.firstVisibleSlot && slotIndex <= visibleRange.lastVisibleSlot;
};

/**
 * Returns whether one resolved series index maps to an actually visible chart point.
 */
const isHoverIndexVisible = (index: number, length: number): boolean => {
  return index >= 0 && index < length;
};

/**
 * Shared color palette applied in series order by all multi-series monitoring charts.
 */
export const chartSeriesPalette = [
  "#2563eb",
  "#16a34a",
  "#ea580c",
  "#dc2626",
  "#7c3aed",
  "#0891b2",
  "#ca8a04",
  "#db2777",
] as const;

/**
 * Formats one ISO timestamp into a compact HH:MM:SS label used across all monitoring charts.
 */
export const formatChartStamp = (value: string): string => {
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
 * Returns whether all multi-series values can be rendered against the same timestamp axis.
 */
export const hasMultiSeriesChartData = (
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
 * Builds legend items pairing each series key with its palette color and latest value.
 */
export const buildChartLegendItems = (
  valuesByKey: Record<string, number[]>,
): { color: string; key: string; latestValue: number }[] => {
  return Object.entries(valuesByKey).flatMap(([key, values], index) => {
    const color = chartSeriesPalette[index % chartSeriesPalette.length];
    const latestValue = values.at(-1);

    if (latestValue === undefined) {
      return [];
    }

    return [
      {
        color,
        key,
        latestValue,
      },
    ];
  });
};

/**
 * Builds X-axis labels for a percent-scale chart using the shared label-count strategy.
 */
export const buildPercentChartAxisLabels = (
  frame: PercentChartFrame,
  stamps: string[],
  capacity: number,
): { label: string; x: number }[] => {
  if (stamps.length === 0) {
    return [];
  }

  const preferredLabelCount = resolvePercentChartLabelCount(stamps.length);
  const labelIndexes = buildEvenlyDistributedIndexes(stamps.length, preferredLabelCount);

  return labelIndexes.flatMap((index) => {
    const stamp = stamps[index];

    if (stamp === undefined) {
      return [];
    }

    return [
      {
        label: formatChartStamp(stamp),
        x: mapPercentChartSeriesX(frame, index, stamps.length, capacity),
      },
    ];
  });
};
