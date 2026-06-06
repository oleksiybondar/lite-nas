import { useAlertsControlPanel } from "@hooks/useAlertsControlPanel";
import Pagination from "@mui/material/Pagination";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

/**
 * Renders page navigation for the current alerts route slice.
 */
export const AlertsControlPanelPagination = (): ReactElement => {
  const { page, setPage, totalCount, totalPages } = useAlertsControlPanel();

  return (
    <Stack
      alignItems={{ md: "center", xs: "flex-start" }}
      data-testid="alerts-pagination-control"
      direction={{ md: "row", xs: "column" }}
      justifyContent="space-between"
      spacing={2}
    >
      <Typography color="text.secondary" data-testid="alerts-pagination-summary" variant="body2">
        {`${totalCount} alerts across ${Math.max(totalPages, 1)} pages`}
      </Typography>
      <Pagination
        count={Math.max(totalPages, 1)}
        data-testid="alerts-pagination"
        onChange={(_, nextPage) => {
          setPage(nextPage);
        }}
        page={Math.min(page, Math.max(totalPages, 1))}
        shape="rounded"
      />
    </Stack>
  );
};
