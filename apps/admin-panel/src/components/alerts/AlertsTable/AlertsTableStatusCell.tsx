import { AlertsTableTextCell } from "@components/alerts/AlertsTable/AlertsTableTextCell";
import type { AlertListItemDTO } from "@dto/alerts/alerts";
import type { ReactElement } from "react";

type AlertsTableStatusCellProps = {
  /**
   * Alert item supplying the current status label.
   */
  item: AlertListItemDTO;
};

/**
 * Renders the plain-text status cell.
 */
export const AlertsTableStatusCell = ({ item }: AlertsTableStatusCellProps): ReactElement => {
  return <AlertsTableTextCell cellName="status" value={item.Status} />;
};
