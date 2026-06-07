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
 * Renders the status cell with severity-like emphasis for user scanning.
 */
export const AlertsTableStatusCell = ({ item }: AlertsTableStatusCellProps): ReactElement => {
  return (
    <AlertsTableTextCell
      cellName="status"
      tone={resolveStatusTone(item.Status)}
      value={item.Status}
    />
  );
};

/**
 * Maps alert status values onto the agreed semantic cell tones.
 */
const resolveStatusTone = (status: string): "error" | "success" | "warning" => {
  if (status === "normal") {
    return "success";
  }

  if (status === "active" || status === "failure") {
    return "error";
  }

  return "warning";
};
