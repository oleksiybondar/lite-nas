/**
 * Theme palette color token used by monitoring cards for semantic emphasis.
 */
export type MonitoringSemanticColor = "error.main" | "info.main" | "success.main" | "warning.main";

/**
 * Formats one byte count into compact binary units for telemetry cards.
 */
export const formatMetricBytes = (value: number): string => {
  const units = ["B", "KiB", "MiB", "GiB", "TiB"] as const;
  let currentValue = value;
  let unitIndex = 0;

  while (currentValue >= 1024 && unitIndex < units.length - 1) {
    currentValue /= 1024;
    unitIndex += 1;
  }

  const roundedValue =
    currentValue >= 10 || unitIndex === 0 ? currentValue.toFixed(0) : currentValue.toFixed(1);

  return `${roundedValue} ${units[unitIndex]}`;
};

/**
 * Formats one byte-rate value into compact binary throughput units for telemetry charts.
 */
export const formatMetricBytesPerSecond = (value: number): string => {
  const units = ["B/s", "KiB/s", "MiB/s", "GiB/s", "TiB/s"] as const;
  let currentValue = value;
  let unitIndex = 0;

  while (currentValue >= 1024 && unitIndex < units.length - 1) {
    currentValue /= 1024;
    unitIndex += 1;
  }

  return `${formatMetricScaledValue(currentValue, unitIndex)} ${units[unitIndex]}`;
};

/**
 * Formats one numeric value into a compact non-percent legend or axis label.
 */
export const formatMetricValue = (value: number): string => {
  return new Intl.NumberFormat(undefined, {
    maximumFractionDigits: value >= 10 ? 0 : 1,
    notation: value >= 1000 ? "compact" : "standard",
  }).format(value);
};

/**
 * Resolves one semantic theme color token from the shared percent threshold bands.
 */
export const resolveMetricPercentColor = (value: number): MonitoringSemanticColor => {
  if (value < 25) {
    return "info.main";
  }

  if (value < 50) {
    return "success.main";
  }

  if (value < 75) {
    return "warning.main";
  }

  return "error.main";
};

/**
 * Resolves one semantic theme color token for ZFS health values rendered in pool metadata.
 */
export const resolveZFSHealthColor = (value: string): MonitoringSemanticColor => {
  const normalizedValue = value.trim().toLowerCase();

  if (normalizedValue === "online") {
    return "success.main";
  }

  if (normalizedValue === "degraded") {
    return "warning.main";
  }

  return "error.main";
};

/**
 * Formats one raw ZFS health value into a human-readable label for pool headers.
 */
export const formatZFSHealthLabel = (value: string): string => {
  const normalizedValue = value.trim().toLowerCase();

  if (normalizedValue.length === 0) {
    return value;
  }

  return `${normalizedValue.charAt(0).toUpperCase()}${normalizedValue.slice(1)}`;
};

/**
 * Formats the latest pool-level error summary string for human-readable metadata output.
 */
export const formatZFSPoolErrorSummary = (value: string): string => {
  return value.toLowerCase() === "none" ? "No known data errors" : value;
};

/**
 * Formats one scaled monitoring value using compact precision that adapts to the current unit range.
 */
const formatMetricScaledValue = (value: number, unitIndex: number): string => {
  if (unitIndex === 0) {
    return value.toFixed(0);
  }

  if (value >= 100) {
    return value.toFixed(0);
  }

  if (value >= 10) {
    return trimMetricTrailingZeroes(value.toFixed(1));
  }

  return trimMetricTrailingZeroes(value.toFixed(2));
};

/**
 * Removes redundant trailing zeroes from one pre-rounded monitoring display value.
 */
const trimMetricTrailingZeroes = (value: string): string => {
  return value.replace(/\.0+$|(\.\d*?[1-9])0+$/u, "$1");
};
