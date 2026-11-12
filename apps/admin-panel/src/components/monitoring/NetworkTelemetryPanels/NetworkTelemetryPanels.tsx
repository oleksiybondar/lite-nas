import Stack from "@mui/material/Stack";
import type { ReactElement } from "react";
import { NetworkInterfacesPanel } from "./NetworkInterfacesPanel";
import { NetworkProtocolsPanel } from "./NetworkProtocolsPanel";
import type { NetworkTelemetryPanelsProps } from "./types";

/**
 * Renders the network telemetry sections as full-width workspace rows.
 */
export const NetworkTelemetryPanels = ({
  capacity,
  interfaces,
  protocols,
}: NetworkTelemetryPanelsProps): ReactElement => {
  return (
    <Stack data-testid="network-telemetry-sections" spacing={1.5} width="100%">
      <NetworkInterfacesPanel capacity={capacity} interfaces={interfaces} />
      <NetworkProtocolsPanel capacity={capacity} protocols={protocols} />
    </Stack>
  );
};
