import type { NetworkInterfacePanelItem } from "@helpers/network-metric-panel";
import type { NetworkProtocolsPanelData } from "@helpers/network-protocols-panel";

export type NetworkTelemetryPanelsProps = {
  /**
   * Maximum number of values represented by the fixed X scale.
   */
  capacity: number;
  /**
   * Latest browser-facing interface rows rendered in the interfaces panel.
   */
  interfaces: NetworkInterfacePanelItem[];
  /**
   * Browser-facing summary and chart data rendered by the protocols-and-sockets panel.
   */
  protocols: NetworkProtocolsPanelData;
};
