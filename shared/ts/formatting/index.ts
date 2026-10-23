/** Formats a nullable byte count with compact binary units for shared UI display. */
export function formatByteSize(size: number | null): string {
  if (size === null) return "—";
  if (size < 1024) return `${size} B`;

  const units = ["KB", "MB", "GB", "TB"];
  let value = size;
  let unit = -1;
  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024;
    unit += 1;
  }
  return `${Number(value.toFixed(1))} ${units[unit]}`;
}
