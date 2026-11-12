import type { PaginationControlState } from "@dto/pagination";
import Pagination from "@mui/material/Pagination";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

type PaginationControlProps = {
  /**
   * Shared pagination state used to render and mutate the current page.
   */
  pagination: PaginationControlState;
  /**
   * Human-readable summary rendered next to the page selector.
   */
  summary: string;
  /**
   * Stable prefix used to derive the component test selectors.
   */
  testIdPrefix: string;
};

/**
 * Reusable pagination summary and page-selector surface shared across list views.
 */
export const PaginationControl = ({
  pagination,
  summary,
  testIdPrefix,
}: PaginationControlProps): ReactElement => {
  return (
    <Stack
      alignItems={{ md: "center", xs: "flex-start" }}
      data-testid={`${testIdPrefix}-pagination-control`}
      direction={{ md: "row", xs: "column" }}
      justifyContent="space-between"
      spacing={2}
    >
      <Typography
        color="text.secondary"
        data-testid={`${testIdPrefix}-pagination-summary`}
        variant="body2"
      >
        {summary}
      </Typography>
      <Pagination
        count={Math.max(pagination.totalPages, 1)}
        data-testid={`${testIdPrefix}-pagination`}
        onChange={(_, nextPage) => {
          pagination.setPage(nextPage);
        }}
        page={Math.min(pagination.page, Math.max(pagination.totalPages, 1))}
        shape="rounded"
      />
    </Stack>
  );
};
