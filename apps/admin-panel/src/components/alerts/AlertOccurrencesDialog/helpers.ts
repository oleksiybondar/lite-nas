import { formatChartStamp } from "@components/monitoring/percent-chart-shared";
import type { AlertOccurrenceItemDTO } from "@dto/alerts/alerts";

const fallbackOccurrenceValue = "-";

/**
 * Returns whether the supplied occurrences should be rendered as a numeric chart.
 */
export const hasNumericAlertOccurrences = (occurrences: AlertOccurrenceItemDTO[]): boolean => {
  if (occurrences.length === 0) {
    return false;
  }

  return occurrences.every((occurrence) => occurrence.ValueNum !== null);
};

/**
 * Formats one occurrence value for table cells, tooltips, and summaries.
 */
export const formatAlertOccurrenceValue = (occurrence: AlertOccurrenceItemDTO): string => {
  return (
    resolveAlertOccurrenceTextValue(occurrence) ??
    resolveAlertOccurrenceNumericValue(occurrence) ??
    resolveAlertOccurrenceBooleanValue(occurrence) ??
    fallbackOccurrenceValue
  );
};

/**
 * Formats one dialog summary line from the current event identifier and count.
 */
export const buildAlertOccurrencesSummary = (
  eventId: string,
  count: number,
  isNumeric: boolean,
): string => {
  const viewLabel = isNumeric ? "Chart view" : "Table view";
  const occurrenceLabel = count === 1 ? "occurrence" : "occurrences";

  return `${viewLabel}. Event ${eventId}. ${count} ${occurrenceLabel} loaded.`;
};

/**
 * Builds the free-text search blob used by the table variant.
 */
export const buildAlertOccurrenceSearchText = (occurrence: AlertOccurrenceItemDTO): string => {
  return [
    occurrence.Timestamp,
    formatChartStamp(occurrence.Timestamp),
    occurrence.EventID,
    String(occurrence.EventRecID),
    formatAlertOccurrenceValue(occurrence),
    occurrence.ValueType,
    occurrence.ValueUnit ?? "",
  ].join(" ");
};

/**
 * Returns the numeric points sorted in ascending timestamp order for chart rendering.
 */
export const buildSortedNumericAlertOccurrences = (
  occurrences: AlertOccurrenceItemDTO[],
): AlertOccurrenceItemDTO[] => {
  return [...occurrences].sort((first, second) => {
    return Date.parse(first.Timestamp) - Date.parse(second.Timestamp);
  });
};

/**
 * Resolves the textual value variant for one occurrence.
 */
const resolveAlertOccurrenceTextValue = (occurrence: AlertOccurrenceItemDTO): string | null => {
  if (occurrence.ValueText === null || occurrence.ValueText === "") {
    return null;
  }

  return occurrence.ValueText;
};

/**
 * Resolves the numeric value variant for one occurrence, including its optional unit.
 */
const resolveAlertOccurrenceNumericValue = (occurrence: AlertOccurrenceItemDTO): string | null => {
  if (occurrence.ValueNum === null) {
    return null;
  }

  if (occurrence.ValueUnit === null || occurrence.ValueUnit === "") {
    return String(occurrence.ValueNum);
  }

  return `${occurrence.ValueNum} ${occurrence.ValueUnit}`;
};

/**
 * Resolves the boolean value variant for one occurrence.
 */
const resolveAlertOccurrenceBooleanValue = (occurrence: AlertOccurrenceItemDTO): string | null => {
  if (occurrence.ValueBool === null) {
    return null;
  }

  return occurrence.ValueBool ? "true" : "false";
};
