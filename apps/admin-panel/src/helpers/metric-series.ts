/**
 * Options used to merge a history window into one local rolling metrics buffer.
 */
type MergeHistoryIntoMetricSeriesOptions<TItem> = {
  currentItems: TItem[];
  getTimestamp: (item: TItem) => string;
  historyResetGapMs: number;
  incomingItems: TItem[];
  maxRecords: number;
};

/**
 * Options used to append one snapshot item into one local rolling metrics buffer.
 */
type AppendSnapshotToMetricSeriesOptions<TItem> = {
  currentItems: TItem[];
  getTimestamp: (item: TItem) => string;
  maxRecords: number;
  snapshotItem: TItem;
};

/**
 * Replaces or extends one rolling metrics buffer with a freshly fetched history window.
 */
export const mergeHistoryIntoMetricSeries = <TItem>({
  currentItems,
  getTimestamp,
  historyResetGapMs,
  incomingItems,
  maxRecords,
}: MergeHistoryIntoMetricSeriesOptions<TItem>): TItem[] => {
  const normalizedCurrentItems = normalizeMetricSeriesItems(currentItems, getTimestamp);
  const normalizedIncomingItems = normalizeMetricSeriesItems(incomingItems, getTimestamp);

  if (shouldUseHistorySeed(normalizedCurrentItems, normalizedIncomingItems)) {
    return trimMetricSeries(
      resolvePreferredHistorySeed(normalizedCurrentItems, normalizedIncomingItems),
      maxRecords,
    );
  }

  const latestCurrentTimestampMs = resolveLatestTimestampMs(normalizedCurrentItems, getTimestamp);

  if (latestCurrentTimestampMs === null) {
    return trimMetricSeries(normalizedCurrentItems, maxRecords);
  }

  const newIncomingItems = normalizedIncomingItems.filter((item) => {
    return toTimestampMs(getTimestamp(item)) > latestCurrentTimestampMs;
  });

  if (newIncomingItems.length === 0) {
    return trimMetricSeries(normalizedCurrentItems, maxRecords);
  }

  const firstNewTimestampMs = resolveLatestTimestampMs(newIncomingItems, getTimestamp);

  if (firstNewTimestampMs === null) {
    return trimMetricSeries(normalizedCurrentItems, maxRecords);
  }

  if (firstNewTimestampMs - latestCurrentTimestampMs > historyResetGapMs) {
    return trimMetricSeries(normalizedIncomingItems, maxRecords);
  }

  return trimMetricSeries([...normalizedCurrentItems, ...newIncomingItems], maxRecords);
};

/**
 * Appends one newer snapshot item to the local rolling metrics buffer.
 */
export const appendSnapshotToMetricSeries = <TItem>({
  currentItems,
  getTimestamp,
  maxRecords,
  snapshotItem,
}: AppendSnapshotToMetricSeriesOptions<TItem>): TItem[] => {
  const normalizedCurrentItems = normalizeMetricSeriesItems(currentItems, getTimestamp);

  if (normalizedCurrentItems.length === 0) {
    return trimMetricSeries([snapshotItem], maxRecords);
  }

  const latestCurrentItem = normalizedCurrentItems.at(-1);

  if (latestCurrentItem === undefined) {
    return trimMetricSeries([snapshotItem], maxRecords);
  }

  const latestCurrentTimestampMs = toTimestampMs(getTimestamp(latestCurrentItem));
  const snapshotTimestampMs = toTimestampMs(getTimestamp(snapshotItem));

  if (snapshotTimestampMs <= latestCurrentTimestampMs) {
    return trimMetricSeries(normalizedCurrentItems, maxRecords);
  }

  return trimMetricSeries([...normalizedCurrentItems, snapshotItem], maxRecords);
};

/**
 * Normalizes one metrics series into ascending, timestamp-unique order.
 */
export const normalizeMetricSeriesItems = <TItem>(
  items: TItem[],
  getTimestamp: (item: TItem) => string,
): TItem[] => {
  const sortedItems = [...items].sort((first, second) => {
    return toTimestampMs(getTimestamp(first)) - toTimestampMs(getTimestamp(second));
  });

  return sortedItems.filter((item, index) => {
    if (index === 0) {
      return true;
    }

    const previousItem = sortedItems[index - 1];

    if (previousItem === undefined) {
      return true;
    }

    return getTimestamp(item) !== getTimestamp(previousItem);
  });
};

/**
 * Trims a normalized metrics series to the latest configured rolling capacity.
 */
export const trimMetricSeries = <TItem>(items: TItem[], maxRecords: number): TItem[] => {
  if (items.length <= maxRecords) {
    return items;
  }

  return items.slice(items.length - maxRecords);
};

const toTimestampMs = (value: string): number => {
  return Date.parse(value);
};

const resolvePreferredHistorySeed = <TItem>(
  normalizedCurrentItems: TItem[],
  normalizedIncomingItems: TItem[],
): TItem[] => {
  return normalizedCurrentItems.length === 0 ? normalizedIncomingItems : normalizedCurrentItems;
};

const shouldUseHistorySeed = <TItem>(
  normalizedCurrentItems: TItem[],
  normalizedIncomingItems: TItem[],
): boolean => {
  return normalizedCurrentItems.length === 0 || normalizedIncomingItems.length === 0;
};

const resolveLatestTimestampMs = <TItem>(
  items: TItem[],
  getTimestamp: (item: TItem) => string,
): number | null => {
  const latestItem = items.at(-1);

  if (latestItem === undefined) {
    return null;
  }

  return toTimestampMs(getTimestamp(latestItem));
};
