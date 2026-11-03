import { PaginationControl } from "@components/pagination";
import { useAlertsControlPanel } from "@hooks/useAlertsControlPanel";
import { useLanguage } from "@lite-nas/language";
import type { ReactElement } from "react";

/**
 * Renders page navigation for the current alerts route slice.
 */
export const AlertsControlPanelPagination = (): ReactElement => {
  const { page, setPage, totalCount, totalPages } = useAlertsControlPanel();
  const { getLabel } = useLanguage();

  return (
    <PaginationControl
      pagination={{
        page,
        setPage,
        totalCount,
        totalPages,
      }}
      summary={getLabel("admin.alerts.paginationSummary")
        .replace("{count}", String(totalCount))
        .replace("{pages}", String(Math.max(totalPages, 1)))}
      testIdPrefix="alerts"
    />
  );
};
