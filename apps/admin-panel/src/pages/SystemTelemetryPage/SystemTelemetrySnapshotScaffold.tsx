import { PaginationControl } from "@components/pagination";
import type { PaginationControlState } from "@dto/pagination";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement, ReactNode } from "react";

type SystemTelemetrySnapshotScaffoldProps = {
  /**
   * Main list or table content rendered between the pagination controls.
   */
  content: ReactNode;
  /**
   * Search and filter controls rendered below the title block.
   */
  controls: ReactElement;
  /**
   * Text rendered below the title as the page summary.
   */
  summary: string;
  /**
   * Human-readable total-count summary rendered above the first pagination control.
   */
  totalSummary: string;
  /**
   * Shared pagination state for the reusable pagination control.
   */
  pagination: PaginationControlState;
  /**
   * Whether the previous page exists.
   */
  hasPreviousPage: boolean;
  /**
   * Whether the next page exists.
   */
  hasNextPage: boolean;
  /**
   * Current page size rendered in the pagination-state caption.
   */
  pageSize: number;
  /**
   * Prefix used to derive stable test selectors.
   */
  testIdPrefix: string;
  /**
   * Singular or plural label used in the pagination caption.
   */
  itemLabel: string;
  /**
   * Card title rendered at the top of the scaffold.
   */
  title: string;
};

/**
 * Shared scaffold for snapshot-style telemetry pages with controls and pagination.
 */
export const SystemTelemetrySnapshotScaffold = ({
  content,
  controls,
  hasNextPage,
  hasPreviousPage,
  itemLabel,
  pageSize,
  pagination,
  summary,
  testIdPrefix,
  title,
  totalSummary,
}: SystemTelemetrySnapshotScaffoldProps): ReactElement => {
  return (
    <Paper data-testid={`${testIdPrefix}-state-card`} sx={{ p: 3 }}>
      <Stack spacing={2}>
        <Stack spacing={1.5}>
          <Typography data-testid={`${testIdPrefix}-state-title`} variant="h2">
            {title}
          </Typography>
          <Typography
            color="text.secondary"
            data-testid={`${testIdPrefix}-state-summary`}
            variant="body2"
          >
            {summary}
          </Typography>
        </Stack>
        {controls}
        <Typography data-testid={`${testIdPrefix}-total-${itemLabel}`} variant="body2">
          {totalSummary}
        </Typography>
        <PaginationControl
          pagination={pagination}
          summary={`${pagination.totalCount} matching ${itemLabel} across ${Math.max(pagination.totalPages, 1)} pages`}
          testIdPrefix={`${testIdPrefix}-top`}
        />
        {content}
        <PaginationControl
          pagination={pagination}
          summary={`${pagination.totalCount} matching ${itemLabel} across ${Math.max(pagination.totalPages, 1)} pages`}
          testIdPrefix={`${testIdPrefix}-bottom`}
        />
        <Typography
          color="text.secondary"
          data-testid={`${testIdPrefix}-pagination-state`}
          variant="caption"
        >
          Page {pagination.page} of {Math.max(pagination.totalPages, 1)}. Showing up to {pageSize}{" "}
          {itemLabel} per page. Previous page available: {String(hasPreviousPage)}. Next page
          available: {String(hasNextPage)}.
        </Typography>
      </Stack>
    </Paper>
  );
};
