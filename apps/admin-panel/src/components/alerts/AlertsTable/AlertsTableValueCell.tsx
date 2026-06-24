import { AlertOccurrencesDialog } from "@components/alerts/AlertOccurrencesDialog";
import { formatAlertLastValue } from "@components/alerts/AlertsTable/helpers";
import type { AlertDomain, AlertListItemDTO } from "@dto/alerts/alerts";
import ListAltIcon from "@mui/icons-material/ListAlt";
import ShowChartIcon from "@mui/icons-material/ShowChart";
import TableCell from "@mui/material/TableCell";
import type { ReactElement } from "react";

type AlertsTableValueCellProps = {
  /**
   * Alert domain owning the occurrences endpoint.
   */
  domain: AlertDomain;
  /**
   * Alert item supplying the last recorded value fields.
   */
  item: AlertListItemDTO;
};

/**
 * Renders the last-value cell with primary emphasis for current measurements.
 */
export const AlertsTableValueCell = ({ domain, item }: AlertsTableValueCellProps): ReactElement => {
  const isNumericValue = item.LastValueNum !== null;

  return (
    <TableCell data-test-class="alerts-table-cell" data-test-name="value" data-test-tone="primary">
      <AlertOccurrencesDialog
        domain={domain}
        eventId={item.EventID}
        triggerIcon={
          isNumericValue ? <ShowChartIcon fontSize="inherit" /> : <ListAltIcon fontSize="inherit" />
        }
        triggerLabel={formatAlertLastValue(item)}
      />
    </TableCell>
  );
};
