import type { AlertDomain, AlertListItemDTO } from "@dto/alerts/alerts";

type AlertsTableColumn = {
  key: string;
  label: string;
};

const baseAlertsTableColumns: AlertsTableColumn[] = [
  { key: "event-id", label: "Event ID" },
  { key: "source", label: "Source" },
  { key: "category", label: "Category" },
  { key: "severity", label: "Severity" },
  { key: "priority", label: "Priority" },
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
