import { AlertsTableTextCell } from "@components/alerts/AlertsTable/AlertsTableTextCell";
import type { AlertListItemDTO } from "@dto/alerts/alerts";
import type { ReactElement } from "react";

type AlertsTableSeverityCellProps = {
  /**
   * Alert item supplying the severity label.
   */
  item: AlertListItemDTO;
};

/**
 * Renders the plain-text severity cell.
 */
export const AlertsTableSeverityCell = ({ item }: AlertsTableSeverityCellProps): ReactElement => {
  return <AlertsTableTextCell cellName="severity" value={item.Severity} />;
};
