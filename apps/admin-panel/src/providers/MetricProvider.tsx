import { MetricContext } from "@contexts/metric-context";
import { useMetricQuery } from "@domain/monitoring/hooks/useMetricQuery";
import type { MetricContextValue, MetricResponseParser } from "@dto/monitoring/metric";
import {
  appendSnapshotToMetricSeries,
  mergeHistoryIntoMetricSeries,
  normalizeMetricSeriesItems,
  trimMetricSeries,
} from "@helpers/metric-series";
import { useMonitoringPollingSettings } from "@hooks/useMonitoringPollingSettings";
import type { UseQueryResult } from "@tanstack/react-query";
import type { PropsWithChildren, ReactElement } from "react";
import { useEffect, useMemo, useRef, useState } from "react";

type MetricProviderProps<TItem> = PropsWithChildren<{
  /**
   * Stable source identifier reused by monitoring polling settings.
   */
  storageKey: string;
  /**
   * History endpoint used for initial bootstrap and history polling mode.
   */
  historyPath: string;
  /**
   * Snapshot endpoint used for snapshot polling mode.
   */
  snapshotPath: string;
  /**
   * Parses and validates one history transport response.
   */
  parseHistoryResponse: MetricResponseParser<TItem[]>;
  /**
   * Parses and validates one snapshot transport response.
   */
  parseSnapshotResponse: MetricResponseParser<TItem>;
  /**
   * Resolves the timestamp identity used by merge helpers.
   */
  getTimestamp: (item: TItem) => string;
}>;

/**
 * Provides one generic metrics polling slice backed by history bootstrap and active polling.
 */
export const MetricProvider = <TItem,>({
  children,
  getTimestamp,
  historyPath,
  parseHistoryResponse,
  parseSnapshotResponse,
  snapshotPath,
  storageKey,
}: MetricProviderProps<TItem>): ReactElement => {
  const { historyIntervalMs, historyResetGapMs, maxRecords, mode, snapshotIntervalMs } =
    useMonitoringPollingSettings();
  const sessionId = useMetricSessionId(mode);
  const bootstrapQuery = useMetricQuery<TItem[]>({
    errorMessage: `Failed to load ${storageKey} history metrics.`,
    parser: parseHistoryResponse,
    path: historyPath,
    queryKey: ["metric", storageKey, "bootstrap", sessionId],
  });
  const activeQuery = useMetricQuery<TItem[] | TItem>({
    enabled: bootstrapQuery.isSuccess,
    errorMessage: `Failed to load ${storageKey} ${mode} metrics.`,
    parser: mode === "history" ? parseHistoryResponse : parseSnapshotResponse,
    path: mode === "history" ? historyPath : snapshotPath,
    queryKey: ["metric", storageKey, mode, sessionId],
    refetchIntervalMs: mode === "history" ? historyIntervalMs : snapshotIntervalMs,
  });
  const items = useMetricItems({
    activeQuery,
    bootstrapQuery,
    getTimestamp,
    historyResetGapMs,
    maxRecords,
    mode,
    sessionId,
  });

  const value = useMemo<MetricContextValue<TItem>>(() => {
    return buildMetricContextValue({ activeQuery, bootstrapQuery, items, mode });
  }, [activeQuery, bootstrapQuery, items, mode]);

  return <MetricContext.Provider value={value}>{children}</MetricContext.Provider>;
};

type UseMetricItemsOptions<TItem> = {
  activeQuery: UseQueryResult<TItem[] | TItem, Error>;
  bootstrapQuery: UseQueryResult<TItem[], Error>;
  getTimestamp: (item: TItem) => string;
  historyResetGapMs: number;
  maxRecords: number;
  mode: MetricContextValue<TItem>["mode"];
  sessionId: number;
};

const useMetricItems = <TItem,>({
  activeQuery,
  bootstrapQuery,
  getTimestamp,
  historyResetGapMs,
  maxRecords,
  mode,
  sessionId,
}: UseMetricItemsOptions<TItem>): TItem[] => {
  const [items, setItems] = useState<TItem[]>([]);

  useEffect(() => {
    if (sessionId === 0) {
      return;
    }

    setItems([]);
  }, [sessionId]);

  useEffect(() => {
    if (!bootstrapQuery.data) {
      return;
    }

    setItems(
      trimMetricSeries(normalizeMetricSeriesItems(bootstrapQuery.data, getTimestamp), maxRecords),
    );
  }, [bootstrapQuery.data, getTimestamp, maxRecords]);

  useEffect(() => {
    if (activeQuery.data === undefined) {
      return;
    }

    setItems((currentItems) => {
      if (mode === "history") {
        return mergeHistoryIntoMetricSeries({
          currentItems,
          getTimestamp,
          historyResetGapMs,
          incomingItems: activeQuery.data as TItem[],
          maxRecords,
        });
      }

      return appendSnapshotToMetricSeries({
        currentItems,
        getTimestamp,
        maxRecords,
        snapshotItem: activeQuery.data as TItem,
      });
    });
  }, [activeQuery.data, getTimestamp, historyResetGapMs, maxRecords, mode]);

  return items;
};

const useMetricSessionId = (mode: MetricContextValue<unknown>["mode"]): number => {
  const [sessionId, setSessionId] = useState(0);
  const previousModeRef = useRef(mode);

  useEffect(() => {
    if (previousModeRef.current === mode) {
      return;
    }

    previousModeRef.current = mode;
    setSessionId((currentSessionId) => currentSessionId + 1);
  }, [mode]);

  return sessionId;
};

type BuildMetricContextValueOptions<TItem> = {
  activeQuery: UseQueryResult<TItem[] | TItem, Error>;
  bootstrapQuery: UseQueryResult<TItem[], Error>;
  items: TItem[];
  mode: MetricContextValue<TItem>["mode"];
};

const buildMetricContextValue = <TItem,>({
  activeQuery,
  bootstrapQuery,
  items,
  mode,
}: BuildMetricContextValueOptions<TItem>): MetricContextValue<TItem> => {
  return {
    error: activeQuery.error ?? bootstrapQuery.error ?? null,
    isError: activeQuery.isError || bootstrapQuery.isError,
    isFetching: activeQuery.isFetching || bootstrapQuery.isFetching,
    isLoading: bootstrapQuery.isLoading,
    items,
    latestItem: items.at(-1) ?? null,
    mode,
    refetch: async (): Promise<unknown> => {
      if (!bootstrapQuery.isSuccess) {
        return bootstrapQuery.refetch();
      }

      return activeQuery.refetch();
    },
  };
};
