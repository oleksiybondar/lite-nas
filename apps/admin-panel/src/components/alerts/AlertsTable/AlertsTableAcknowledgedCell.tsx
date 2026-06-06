import { AlertsTableTextCell } from "@components/alerts/AlertsTable/AlertsTableTextCell";
import { formatAcknowledgedValue } from "@components/alerts/AlertsTable/helpers";
import type { AlertListItemDTO } from "@dto/alerts/alerts";
import type { ReactElement } from "react";

type AlertsTableAcknowledgedCellProps = {
  /**
   * Alert item supplying the acknowledged state.
   */
  item: AlertListItemDTO;
};

/**
 * Renders the starter acknowledged-state cell.
 */
export const AlertsTableAcknowledgedCell = ({
  item,
}: AlertsTableAcknowledgedCellProps): ReactElement => {
  return (
    <AlertsTableTextCell
      cellName="acknowledged"
      value={formatAcknowledgedValue(item.Acknowledged)}
    />
  );
};
