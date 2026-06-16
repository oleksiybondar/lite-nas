import type { NetworkInterfacePanelItem } from "@helpers/network-metric-panel";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import { NetworkInterfaceCard } from "./NetworkInterfaceCard";

type NetworkInterfacesPanelProps = {
  /**
   * Maximum number of values represented by the fixed X scale.
   */
  capacity: number;
  /**
   * Latest browser-facing interface rows rendered in the interfaces panel.
   */
  interfaces: NetworkInterfacePanelItem[];
};

/**
 * Renders the first draft of the network interfaces panel with full interface cards.
 */
export const NetworkInterfacesPanel = ({
  capacity,
  interfaces,
}: NetworkInterfacesPanelProps): ReactElement => {
  return (
    <Paper data-testid="network-interfaces-panel" sx={{ p: 2, width: "100%" }}>
      <Stack spacing={1.5}>
        <Typography data-testid="network-interfaces-panel-title" variant="h2">
          Interfaces
        </Typography>
        {interfaces.length === 0 ? (
          <Typography
            color="text.secondary"
            data-testid="network-interfaces-panel-empty"
            variant="body2"
          >
            Interface data will appear after telemetry points are loaded.
          </Typography>
        ) : (
          <Stack data-testid="network-interfaces-subrow" spacing={1.5}>
            {interfaces.map((item) => {
              return (
                <NetworkInterfaceCard capacity={capacity} key={item.name} networkInterface={item} />
              );
            })}
          </Stack>
        )}
      </Stack>
    </Paper>
  );
};
