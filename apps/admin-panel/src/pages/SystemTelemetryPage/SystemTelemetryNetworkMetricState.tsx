import { NetworkTelemetryPanels } from "@components/monitoring/NetworkTelemetryPanels";
import type { NetworkInterfacePanelItem } from "@helpers/network-metric-panel";
import { buildNetworkInterfacePanelItems } from "@helpers/network-metric-panel";
import { buildNetworkProtocolsPanelData } from "@helpers/network-protocols-panel";
import { useFilteredRecords } from "@hooks/useFilteredRecords";
import { useMonitoringPollingSettings } from "@hooks/useMonitoringPollingSettings";
import { useNetworkMetric } from "@hooks/useNetworkMetric";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

/**
 * Route state rendered when gateway-backed network telemetry is available.
 */
export const SystemTelemetryNetworkMetricState = (): ReactElement => {
  const { items } = useNetworkMetric();
  const { maxRecords } = useMonitoringPollingSettings();
  const interfaces = buildNetworkInterfacePanelItems(items);
  const protocols = buildNetworkProtocolsPanelData(items);
  const {
    filteredCount,
    records: visibleInterfaces,
    search,
    setSearch,
    totalCount,
  } = useFilteredRecords<NetworkInterfacePanelItem, string>({
    getSearchText: buildNetworkInterfaceSearchText,
    initialFilters: {},
    matchesFilter: () => true,
    records: interfaces,
  });

  return (
    <Stack spacing={3}>
      <Stack direction={{ md: "row", xs: "column" }} spacing={1.5}>
        <TextField
          data-testid="network-metric-search-control"
          label="Search interfaces"
          name="networkMetricSearch"
          onChange={(event) => {
            setSearch(event.target.value);
          }}
          size="small"
          sx={{ minWidth: 240 }}
          value={search}
        />
        <Typography data-testid="network-metric-search-summary" variant="body2">
          {filteredCount} of {totalCount} interfaces match the current search.
        </Typography>
      </Stack>
      <NetworkTelemetryPanels
        capacity={maxRecords}
        interfaces={visibleInterfaces}
        protocols={protocols}
      />
    </Stack>
  );
};

/**
 * Builds the interface search blob used by the generic filtered-records hook.
 */
const buildNetworkInterfaceSearchText = (item: NetworkInterfacePanelItem): string => {
  return [
    item.name,
    item.adapterLabel,
    item.linkDetailsLabel,
    item.status,
    ...item.totalsLabels.map((label) => `${label.key} ${label.value}`),
  ].join(" ");
};
