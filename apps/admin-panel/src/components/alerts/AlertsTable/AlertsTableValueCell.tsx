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
 * Renders the last-value cell with primary emphasis for current measurements.
 */
export const AlertsTableValueCell = ({ item }: AlertsTableValueCellProps): ReactElement => {
  return <AlertsTableTextCell cellName="value" tone="primary" value={formatAlertLastValue(item)} />;
};
