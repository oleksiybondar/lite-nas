import { AlertsTableTextCell } from "@components/alerts/AlertsTable/AlertsTableTextCell";
import type { AlertListItemDTO } from "@dto/alerts/alerts";
import type { ReactElement } from "react";

type AlertsTableSourceCellProps = {
  /**
   * Alert item supplying the source label.
   */
  item: AlertListItemDTO;
};

/**
 * Renders the plain-text source cell.
 */
export const AlertsTableSourceCell = ({ item }: AlertsTableSourceCellProps): ReactElement => {
  return <AlertsTableTextCell cellName="source" value={item.Source} />;
};
