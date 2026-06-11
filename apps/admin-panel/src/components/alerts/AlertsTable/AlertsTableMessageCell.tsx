import { AlertsTableTextCell } from "@components/alerts/AlertsTable/AlertsTableTextCell";
import type { AlertListItemDTO } from "@dto/alerts/alerts";
import type { ReactElement } from "react";

type AlertsTableMessageCellProps = {
  /**
   * Alert item supplying the message text.
   */
  item: AlertListItemDTO;
};

/**
 * Renders the plain-text message cell.
 */
export const AlertsTableMessageCell = ({ item }: AlertsTableMessageCellProps): ReactElement => {
  return <AlertsTableTextCell cellName="message" value={item.Message} />;
};
