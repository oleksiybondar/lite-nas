import { NetworkTelemetryPanels } from "@components/monitoring/NetworkTelemetryPanels";
import { buildNetworkInterfacePanelItems } from "@helpers/network-metric-panel";
import { buildNetworkProtocolsPanelData } from "@helpers/network-protocols-panel";
import { useMonitoringPollingSettings } from "@hooks/useMonitoringPollingSettings";
import { useNetworkMetric } from "@hooks/useNetworkMetric";
import Stack from "@mui/material/Stack";
import type { ReactElement } from "react";

/**
 * Route state rendered when gateway-backed network telemetry is available.
 */
export const SystemTelemetryNetworkMetricState = (): ReactElement => {
  const { items } = useNetworkMetric();
  const { maxRecords } = useMonitoringPollingSettings();
  const interfaces = buildNetworkInterfacePanelItems(items);
  const protocols = buildNetworkProtocolsPanelData(items);

  return (
    <Stack spacing={3}>
      <NetworkTelemetryPanels capacity={maxRecords} interfaces={interfaces} protocols={protocols} />
    </Stack>
  );
};
