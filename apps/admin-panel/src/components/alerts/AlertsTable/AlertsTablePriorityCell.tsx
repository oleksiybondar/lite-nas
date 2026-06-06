import { AlertsTableTextCell } from "@components/alerts/AlertsTable/AlertsTableTextCell";
import type { AlertListItemDTO } from "@dto/alerts/alerts";
import type { ReactElement } from "react";

type AlertsTablePriorityCellProps = {
  /**
   * Alert item supplying the numeric priority.
   */
  item: AlertListItemDTO;
};

/**
 * Renders the plain-text priority cell.
 */
export const AlertsTablePriorityCell = ({ item }: AlertsTablePriorityCellProps): ReactElement => {
  return <AlertsTableTextCell cellName="priority" value={item.Priority} />;
};
