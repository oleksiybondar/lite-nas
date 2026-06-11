import { AlertsTableTextCell } from "@components/alerts/AlertsTable/AlertsTableTextCell";
import { formatMitigateValue } from "@components/alerts/AlertsTable/helpers";
import type { AlertListItemDTO } from "@dto/alerts/alerts";
import type { ReactElement } from "react";

type AlertsTableMitigateCellProps = {
  /**
   * Alert item supplying the future mitigation hint metadata.
   */
  item: AlertListItemDTO;
};

/**
 * Renders the security-only mitigation hint cell.
 */
export const AlertsTableMitigateCell = ({ item }: AlertsTableMitigateCellProps): ReactElement => {
  const value = formatMitigateValue(item);

  return (
    <AlertsTableTextCell
      cellName="mitigate"
      tone={value === "-" ? "default" : "primary"}
      value={value}
    />
  );
};
