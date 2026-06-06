import { AlertsTableTextCell } from "@components/alerts/AlertsTable/AlertsTableTextCell";
import type { AlertListItemDTO } from "@dto/alerts/alerts";
import type { ReactElement } from "react";

type AlertsTableEventIdCellProps = {
  /**
   * Alert item supplying the event identifier.
   */
  item: AlertListItemDTO;
};

/**
 * Renders the plain-text event identifier cell.
 */
export const AlertsTableEventIdCell = ({ item }: AlertsTableEventIdCellProps): ReactElement => {
  return <AlertsTableTextCell cellName="event-id" value={item.EventID} />;
};
