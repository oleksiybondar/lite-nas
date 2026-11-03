import type { AlertDomain, AlertListItemDTO } from "@dto/alerts/alerts";
import type { SxProps, Theme } from "@mui/material/styles";

type AlertsTableColumn = {
  key: string;
  label: string;
  title?: string;
};

const severityColumnWidth = 44;
const priorityColumnWidth = 44;

/**
 * Returns the alerts table columns for one concrete alert domain.
 */
export const buildAlertsTableColumns = (
  domain: AlertDomain,
  getLabel: (key: string) => string,
): AlertsTableColumn[] => {
  const columns: AlertsTableColumn[] = [
    {
      key: "severity",
      label: "S",
      title: getLabel("admin.alertTable.severityShortTitle"),
    },
    {
      key: "priority",
      label: "P",
      title: getLabel("admin.alertTable.priorityShortTitle"),
    },
    { key: "event-id", label: getLabel("admin.alertTable.eventId") },
    { key: "message", label: getLabel("admin.alertTable.message") },
    { key: "value", label: getLabel("admin.alertTable.value") },
    { key: "source", label: getLabel("admin.alertTable.source") },
    { key: "status", label: getLabel("admin.alertTable.columnStatus") },
    { key: "acknowledgement", label: getLabel("admin.alertTable.acknowledged") },
    { key: "category", label: getLabel("admin.alertTable.category") },
    { key: "created-at", label: getLabel("admin.alertTable.createdAt") },
    { key: "acknowledged-at", label: getLabel("admin.alertTable.acknowledgedAt") },
  ];
  if (domain === "security") {
    return [
      ...columns.slice(0, 7),
      { key: "mitigate", label: getLabel("admin.alertTable.mitigate") },
      ...columns.slice(7),
    ];
  }

  return columns;
};

/**
 * Formats one optional timestamp field for table text rendering.
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
export const formatAlertSeverityLabel = (
  severity: string,
  getLabel: (key: string) => string,
): string => {
  const knownSeverity = ["critical", "error", "warning", "info"].find(
    (value) => value === severity.toLowerCase(),
  );
  return knownSeverity ? getLabel(`admin.alertTable.severity.${knownSeverity}`) : severity;
};

/**
 * Formats the acknowledged-by label shown once an alert has already been acknowledged.
 */
export const formatAcknowledgedByValue = (value: string | null): string => {
  if (value === null || value === "") {
    return "-";
  }

  return value;
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
 * Resolves the text representation for the last text value field.
 */
const resolveAlertTextValue = (item: AlertListItemDTO): string | null => {
  if (item.LastValueText === null || item.LastValueText === "") {
    return null;
  }

  return item.LastValueText;
};

/**
 * Resolves the numeric representation for the last numeric value field.
 */
const resolveAlertNumericValue = (item: AlertListItemDTO): number | null => {
  return item.LastValueNum;
};

/**
 * Resolves the boolean representation for the last boolean value field.
 */
const resolveAlertBooleanValue = (item: AlertListItemDTO): string | null => {
  if (item.LastValueBool === null) {
    return null;
  }

  return item.LastValueBool ? "true" : "false";
};
