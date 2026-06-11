import { AlertsTableTextCell } from "@components/alerts/AlertsTable/AlertsTableTextCell";
import { formatAlertsTimestamp } from "@components/alerts/AlertsTable/helpers";
import type { AlertListItemDTO } from "@dto/alerts/alerts";
import type { ReactElement } from "react";

type AlertsTableCreatedAtCellProps = {
  /**
   * Alert item supplying the creation timestamp.
   */
  item: AlertListItemDTO;
};

/**
 * Renders the starter created-at timestamp cell.
 */
export const AlertsTableCreatedAtCell = ({ item }: AlertsTableCreatedAtCellProps): ReactElement => {
  return (
    <AlertsTableTextCell cellName="created-at" value={formatAlertsTimestamp(item.CreatedAt)} />
  );
};
