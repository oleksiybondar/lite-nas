import type { NetworkInterfacePanelItem } from "@helpers/network-metric-panel";

export type NetworkInterfaceCardProps = {
  /**
   * Maximum number of values represented by the fixed X scale.
   */
  capacity: number;
  /**
   * Browser-facing interface card data built from network metrics history.
   */
  networkInterface: NetworkInterfacePanelItem;
};
