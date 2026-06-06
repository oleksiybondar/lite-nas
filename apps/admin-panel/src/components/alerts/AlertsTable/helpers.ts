import type { AlertDomain, AlertListItemDTO } from "@dto/alerts/alerts";
import type { SxProps, Theme } from "@mui/material/styles";

type AlertsTableColumn = {
  key: string;
  label: string;
};

const severityColumnWidth = 56;
const priorityColumnWidth = 56;

const baseAlertsTableColumns: AlertsTableColumn[] = [
  { key: "severity", label: "S" },
  { key: "priority", label: "P" },
  { key: "event-id", label: "Event ID" },
  { key: "source", label: "Source" },
  { key: "category", label: "Category" },
  { key: "status", label: "Status" },
  { key: "created-at", label: "Created at" },
  { key: "acknowledged-at", label: "Acknowledged at" },
  { key: "acknowledged", label: "Acknowledged" },
  { key: "value", label: "Value" },
  { key: "message", label: "Message" },
  { key: "acknowledge", label: "Acknowledge" },
];

/**
 * Returns the starter alerts table columns for one concrete alert domain.
 */
export const buildAlertsTableColumns = (domain: AlertDomain): AlertsTableColumn[] => {
  if (domain === "security") {
    return [...baseAlertsTableColumns, { key: "mitigate", label: "Mitigate" }];
  }

  return baseAlertsTableColumns;
};

/**
 * Formats one optional timestamp field for starter table text rendering.
 */
export const formatAlertsTimestamp = (value: string | null): string => {
  if (value === null || value === "") {
    return "-";
  }

  return value;
};

/**
 * Returns sticky column styling for supported alerts table columns.
 */
export const buildAlertsStickyColumnSx = (columnKey: string, isHeader = false): SxProps<Theme> => {
  const stickyConfig = resolveStickyColumnConfig(columnKey);

  if (stickyConfig === null) {
    return {};
  }

  return {
    backgroundColor: "background.paper",
    left: stickyConfig.left,
    minWidth: stickyConfig.width,
    position: "sticky",
    width: stickyConfig.width,
    zIndex: isHeader ? 4 : 2,
  };
};

/**
 * Formats one severity value for display in typed severity cells.
 */
export const formatAlertSeverityLabel = (severity: string): string => {
  return severity.slice(0, 1).toUpperCase() + severity.slice(1);
};

/**
 * Formats the starter acknowledged flag for table display.
 */
export const formatAcknowledgedValue = (acknowledged: boolean): string => {
  return acknowledged ? "Yes" : "No";
};

/**
 * Formats the last observed alert value and optional unit for table display.
 */
export const formatAlertLastValue = (item: AlertListItemDTO): string => {
  const rawValue = resolveAlertLastValue(item);

  if (rawValue === null) {
    return "-";
  }

  if (item.LastValueUnit === null || item.LastValueUnit === "") {
    return String(rawValue);
  }

  return `${rawValue} ${item.LastValueUnit}`;
};

/**
 * Formats the starter mitigate column value for security alerts.
 */
export const formatMitigateValue = (item: AlertListItemDTO): string => {
  return item.Meta?.mitigate ?? "-";
};

type StickyColumnConfig = {
  left: number;
  width: number;
};

const stickyColumnConfigByKey: Partial<Record<string, StickyColumnConfig>> = {
  priority: {
    left: severityColumnWidth,
    width: priorityColumnWidth,
  },
  severity: {
    left: 0,
    width: severityColumnWidth,
  },
};

const resolveAlertLastValue = (item: AlertListItemDTO): boolean | number | string | null => {
  const lastTextValue = resolveAlertTextValue(item);

  if (lastTextValue !== null) {
    return lastTextValue;
  }

  const lastNumericValue = resolveAlertNumericValue(item);

  if (lastNumericValue !== null) {
    return lastNumericValue;
  }

  return resolveAlertBooleanValue(item);
};

/**
 * Resolves sticky column layout data for supported alerts table columns.
 */
const resolveStickyColumnConfig = (columnKey: string): StickyColumnConfig | null => {
  return stickyColumnConfigByKey[columnKey] ?? null;
};

/**
 * Resolves the starter text representation for the last text value field.
 */
const resolveAlertTextValue = (item: AlertListItemDTO): string | null => {
  if (item.LastValueText === null || item.LastValueText === "") {
    return null;
  }

  return item.LastValueText;
};

/**
 * Resolves the starter numeric representation for the last numeric value field.
 */
const resolveAlertNumericValue = (item: AlertListItemDTO): number | null => {
  return item.LastValueNum;
};

/**
 * Resolves the starter boolean representation for the last boolean value field.
 */
const resolveAlertBooleanValue = (item: AlertListItemDTO): string | null => {
  if (item.LastValueBool === null) {
    return null;
  }

  return item.LastValueBool ? "true" : "false";
};
