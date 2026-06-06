import { AlertsTableTextCell } from "@components/alerts/AlertsTable/AlertsTableTextCell";
import { formatAlertLastValue } from "@components/alerts/AlertsTable/helpers";
import type { AlertListItemDTO } from "@dto/alerts/alerts";
import type { ReactElement } from "react";

type AlertsTableValueCellProps = {
  /**
   * Alert item supplying the last recorded value fields.
   */
  item: AlertListItemDTO;
};

/**
 * Renders the starter last-value cell.
 */
export const AlertsTableValueCell = ({ item }: AlertsTableValueCellProps): ReactElement => {
  return <AlertsTableTextCell cellName="value" value={formatAlertLastValue(item)} />;
};
