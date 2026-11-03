import { AlertsTableTextCell } from "@components/alerts/AlertsTable/AlertsTableTextCell";
import type { AlertListItemDTO } from "@dto/alerts/alerts";
import { useLanguage } from "@lite-nas/language";
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
  const { getLabel } = useLanguage();
  return (
    <AlertsTableTextCell
      cellName="status"
      tone={resolveStatusTone(item.Status)}
      value={resolveStatusLabel(item.Status, getLabel)}
    />
  );
};

/** Resolves known alert lifecycle states to localized labels while preserving unknown values. */
const resolveStatusLabel = (status: string, getLabel: (key: string) => string): string => {
  const knownStatus = ["normal", "active", "failure"].find(
    (value) => value === status.toLowerCase(),
  );
  return knownStatus ? getLabel(`admin.alertTable.state.${knownStatus}`) : status;
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
