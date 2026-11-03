import type { NetworkProtocolsPanelData } from "@helpers/network-protocols-panel";

export type NetworkProtocolsPanelProps = {
  /**
   * Maximum number of values represented by the fixed X scale.
   */
  capacity: number;
  /**
   * Browser-facing summary and chart data rendered by the protocols-and-sockets panel.
   */
  protocols: NetworkProtocolsPanelData;
};
