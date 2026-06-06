import { AlertsTableTextCell } from "@components/alerts/AlertsTable/AlertsTableTextCell";
import { formatAlertsTimestamp } from "@components/alerts/AlertsTable/helpers";
import type { AlertListItemDTO } from "@dto/alerts/alerts";
import type { ReactElement } from "react";

type AlertsTableAcknowledgedAtCellProps = {
  /**
   * Alert item supplying the acknowledged-at timestamp.
   */
  item: AlertListItemDTO;
};

/**
 * Renders the starter acknowledged-at timestamp cell.
 */
export const AlertsTableAcknowledgedAtCell = ({
  item,
}: AlertsTableAcknowledgedAtCellProps): ReactElement => {
  return (
    <AlertsTableTextCell
      cellName="acknowledged-at"
      value={formatAlertsTimestamp(item.AcknowledgedAt)}
    />
  );
};
