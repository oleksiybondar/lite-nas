import { PaginationControl } from "@components/pagination";
import { useAlertsControlPanel } from "@hooks/useAlertsControlPanel";
import type { ReactElement } from "react";

/**
 * Renders page navigation for the current alerts route slice.
 */
export const AlertsControlPanelPagination = (): ReactElement => {
  const { page, setPage, totalCount, totalPages } = useAlertsControlPanel();

  return (
    <PaginationControl
      pagination={{
        page,
        setPage,
        totalCount,
        totalPages,
      }}
      summary={`${totalCount} alerts across ${Math.max(totalPages, 1)} pages`}
      testIdPrefix="alerts"
    />
  );
};
