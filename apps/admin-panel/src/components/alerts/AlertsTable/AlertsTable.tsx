import { AlertsTableRow } from "@components/alerts/AlertsTable/AlertsTableRow";
import {
  buildAlertsStickyColumnSx,
  buildAlertsTableColumns,
} from "@components/alerts/AlertsTable/helpers";
import { useAlerts } from "@hooks/useAlerts";
import Paper from "@mui/material/Paper";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableFooter from "@mui/material/TableFooter";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

/**
 * Renders the starter alerts table for the current route slice.
 */
export const AlertsTable = (): ReactElement => {
  const { domain, items } = useAlerts();
  const columns = buildAlertsTableColumns(domain);

  return (
    <TableContainer
      component={Paper}
      data-testid="alerts-table"
      sx={{ maxWidth: "100%", minWidth: 0, overflowX: "auto", width: "100%" }}
    >
      <Table size="small" stickyHeader sx={{ minWidth: 1560 }}>
        <TableHead>
          <TableRow>
            {columns.map((column) => {
              return (
                <TableCell
                  data-test-class="alerts-table-header-cell"
                  data-test-name={column.key}
                  key={column.key}
                  sx={buildAlertsStickyColumnSx(column.key, true)}
                >
                  {column.label}
                </TableCell>
              );
            })}
          </TableRow>
        </TableHead>
        <TableBody>
          {items.length === 0 ? (
            <TableRow data-testid="alerts-table-empty-row">
              <TableCell colSpan={columns.length}>
                <Typography color="text.secondary" variant="body2">
                  No alerts found on this page.
                </Typography>
              </TableCell>
            </TableRow>
          ) : (
            items.map((item) => {
              return (
                <AlertsTableRow
                  domain={domain}
                  item={item}
                  key={`${item.EventID}:${item.EventRecID}`}
                />
              );
            })
          )}
        </TableBody>
        <TableFooter>
          <TableRow data-testid="alerts-table-footer-row">
            {columns.map((column) => {
              return (
                <TableCell
                  data-test-class="alerts-table-footer-cell"
                  data-test-name={column.key}
                  key={`footer-${column.key}`}
                  sx={buildAlertsStickyColumnSx(column.key, true)}
                >
                  {column.label}
                </TableCell>
              );
            })}
          </TableRow>
        </TableFooter>
      </Table>
    </TableContainer>
  );
};
