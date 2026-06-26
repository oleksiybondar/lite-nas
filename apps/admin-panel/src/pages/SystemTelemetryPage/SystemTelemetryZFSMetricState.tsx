import { ZFSPoolCard } from "@components/monitoring/ZFSPoolCard";
import type { ZFSPoolCardData } from "@helpers/zfs-metric-chart";
import { buildZFSPoolCardData } from "@helpers/zfs-metric-chart";
import { useFilteredRecords } from "@hooks/useFilteredRecords";
import { useMonitoringPollingSettings } from "@hooks/useMonitoringPollingSettings";
import { useZFSMetric } from "@hooks/useZFSMetric";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import type { SupportedTelemetryRoute } from "@pages/SystemTelemetryPage/helpers";
import type { ReactElement } from "react";

type SystemTelemetryZFSMetricStateProps = {
  /**
   * Current route metadata resolved from the browser URL.
   */
  route: SupportedTelemetryRoute;
};

/**
 * Route state rendered when gateway-backed ZFS pool metrics are available.
 */
export const SystemTelemetryZFSMetricState = ({
  route,
}: SystemTelemetryZFSMetricStateProps): ReactElement => {
  const { items } = useZFSMetric();
  const { maxRecords } = useMonitoringPollingSettings();
  const poolCards = buildZFSPoolCardData(items);
  const {
    filteredCount,
    records: visiblePoolCards,
    search,
    setSearch,
    totalCount,
  } = useFilteredRecords<ZFSPoolCardData, string>({
    getSearchText: buildZFSPoolSearchText,
    initialFilters: {},
    matchesFilter: () => true,
    records: poolCards,
  });

  if (poolCards.length === 0) {
    return (
      <Typography data-testid="system-telemetry-zfs-empty" variant="body2">
        {route.summary}
      </Typography>
    );
  }

  return (
    <Stack spacing={3}>
      <Stack direction={{ md: "row", xs: "column" }} spacing={1.5}>
        <TextField
          data-testid="zfs-metric-search-control"
          label="Search pools"
          name="zfsMetricSearch"
          onChange={(event) => {
            setSearch(event.target.value);
          }}
          size="small"
          sx={{ minWidth: 240 }}
          value={search}
        />
        <Typography data-testid="zfs-metric-search-summary" variant="body2">
          {filteredCount} of {totalCount} pools match the current search.
        </Typography>
      </Stack>
      {visiblePoolCards.length === 0 ? (
        <Typography data-testid="system-telemetry-zfs-search-empty" variant="body2">
          No pools match the current search.
        </Typography>
      ) : (
        visiblePoolCards.map((pool) => {
          return <ZFSPoolCard capacity={maxRecords} key={pool.name} pool={pool} />;
        })
      )}
    </Stack>
  );
};

/**
 * Builds the ZFS pool search blob used by the generic filtered-records hook.
 */
const buildZFSPoolSearchText = (pool: ZFSPoolCardData): string => {
  return [
    pool.name,
    pool.health,
    pool.scan,
    pool.poolErrorSummary,
    pool.usedPercentLabel,
    ...pool.metadataLabels.map((label) => `${label.key} ${label.value}`),
  ].join(" ");
};
