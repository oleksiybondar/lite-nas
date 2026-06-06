import { AlertsTableTextCell } from "@components/alerts/AlertsTable/AlertsTableTextCell";
import type { AlertListItemDTO } from "@dto/alerts/alerts";
import type { ReactElement } from "react";

type AlertsTableCategoryCellProps = {
  /**
   * Alert item supplying the category label.
   */
  item: AlertListItemDTO;
};

/**
 * Renders the plain-text category cell.
 */
export const AlertsTableCategoryCell = ({ item }: AlertsTableCategoryCellProps): ReactElement => {
  return <AlertsTableTextCell cellName="category" value={item.Category} />;
};
