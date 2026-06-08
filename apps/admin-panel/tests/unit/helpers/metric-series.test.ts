import {
  appendSnapshotToMetricSeries,
  mergeHistoryIntoMetricSeries,
  normalizeMetricSeriesItems,
} from "@helpers/metric-series";

type SampleItem = {
  timestamp: string;
  value: number;
};

const getTimestamp = (item: SampleItem): string => item.timestamp;

describe("normalizeMetricSeriesItems", () => {
  test("sorts items ascending and ignores duplicate timestamps", () => {
    expect(
      normalizeMetricSeriesItems(
        [
          buildItem("2026-06-07T10:00:02Z", 2),
          buildItem("2026-06-07T10:00:01Z", 1),
          buildItem("2026-06-07T10:00:02Z", 999),
        ],
        getTimestamp,
      ),
    ).toEqual([buildItem("2026-06-07T10:00:01Z", 1), buildItem("2026-06-07T10:00:02Z", 2)]);
  });
});

describe("mergeHistoryIntoMetricSeries", () => {
  test("appends only newer history items and trims to the latest capacity", () => {
    expect(
      mergeHistoryIntoMetricSeries({
        currentItems: [buildItem("2026-06-07T10:00:01Z", 1), buildItem("2026-06-07T10:00:02Z", 2)],
        getTimestamp,
        historyResetGapMs: 10000,
        incomingItems: [
          buildItem("2026-06-07T10:00:01Z", 1),
          buildItem("2026-06-07T10:00:02Z", 2),
          buildItem("2026-06-07T10:00:03Z", 3),
          buildItem("2026-06-07T10:00:04Z", 4),
        ],
        maxRecords: 3,
      }),
    ).toEqual([
      buildItem("2026-06-07T10:00:02Z", 2),
      buildItem("2026-06-07T10:00:03Z", 3),
      buildItem("2026-06-07T10:00:04Z", 4),
    ]);
  });

  test("replaces the local cache when the history gap exceeds the reset threshold", () => {
    expect(
      mergeHistoryIntoMetricSeries({
        currentItems: [buildItem("2026-06-07T10:00:01Z", 1)],
        getTimestamp,
        historyResetGapMs: 10000,
        incomingItems: [
          buildItem("2026-06-07T10:00:20Z", 20),
          buildItem("2026-06-07T10:00:21Z", 21),
        ],
        maxRecords: 5,
      }),
    ).toEqual([buildItem("2026-06-07T10:00:20Z", 20), buildItem("2026-06-07T10:00:21Z", 21)]);
  });
});

describe("appendSnapshotToMetricSeries", () => {
  test("ignores same-timestamp snapshots and appends only newer data", () => {
    expect(
      appendSnapshotToMetricSeries({
        currentItems: [buildItem("2026-06-07T10:00:01Z", 1), buildItem("2026-06-07T10:00:02Z", 2)],
        getTimestamp,
        maxRecords: 2,
        snapshotItem: buildItem("2026-06-07T10:00:02Z", 999),
      }),
    ).toEqual([buildItem("2026-06-07T10:00:01Z", 1), buildItem("2026-06-07T10:00:02Z", 2)]);

    expect(
      appendSnapshotToMetricSeries({
        currentItems: [buildItem("2026-06-07T10:00:01Z", 1), buildItem("2026-06-07T10:00:02Z", 2)],
        getTimestamp,
        maxRecords: 2,
        snapshotItem: buildItem("2026-06-07T10:00:03Z", 3),
      }),
    ).toEqual([buildItem("2026-06-07T10:00:02Z", 2), buildItem("2026-06-07T10:00:03Z", 3)]);
  });
});

const buildItem = (timestamp: string, value: number): SampleItem => ({
  timestamp,
  value,
});
