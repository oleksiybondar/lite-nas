/**
 * Shared outlined detail-column layout used by monitoring table workspaces.
 */
export const monitoringDetailColumnSx = {
  flex: "1 1 480px",
  minWidth: 0,
  p: 1,
} as const;

/**
 * Builds the shared scrollable table container layout for monitoring detail tables.
 */
export const createMonitoringScrollableTableSx = (height: number) => {
  return {
    maxHeight: height,
    minHeight: height,
    overflowY: "auto",
  } as const;
};
