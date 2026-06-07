import { AlertsTableTextCell } from "@components/alerts/AlertsTable/AlertsTableTextCell";
import type { AlertListItemDTO } from "@dto/alerts/alerts";
import type { ReactElement } from "react";

type AlertsTableSourceCellProps = {
  /**
   * Alert item supplying the source label.
   */
  item: AlertListItemDTO;
};

/**
 * Renders the source cell with warning emphasis to surface alert provenance.
 */
export const AlertsTableSourceCell = ({ item }: AlertsTableSourceCellProps): ReactElement => {
  return <AlertsTableTextCell cellName="source" tone="warning" value={item.Source} />;
};
