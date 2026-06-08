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

  if (normalizedCurrentItems.length === 0) {
    return trimMetricSeries(normalizedIncomingItems, maxRecords);
  }

  if (normalizedIncomingItems.length === 0) {
    return trimMetricSeries(normalizedCurrentItems, maxRecords);
  }

  const latestCurrentTimestampMs = toTimestampMs(
    getTimestamp(normalizedCurrentItems[normalizedCurrentItems.length - 1]),
  );
  const newIncomingItems = normalizedIncomingItems.filter((item) => {
    return toTimestampMs(getTimestamp(item)) > latestCurrentTimestampMs;
  });

  if (newIncomingItems.length === 0) {
    return trimMetricSeries(normalizedCurrentItems, maxRecords);
  }

  const firstNewTimestampMs = toTimestampMs(getTimestamp(newIncomingItems[0]));

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

  const latestCurrentTimestampMs = toTimestampMs(
    getTimestamp(normalizedCurrentItems[normalizedCurrentItems.length - 1]),
  );
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

    return getTimestamp(item) !== getTimestamp(sortedItems[index - 1]);
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
