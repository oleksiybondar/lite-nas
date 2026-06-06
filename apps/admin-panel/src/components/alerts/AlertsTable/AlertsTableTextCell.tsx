import TableCell from "@mui/material/TableCell";
import type { ReactElement } from "react";

type AlertsTableTextCellProps = {
  /**
   * Stable cell discriminator for test selectors.
   */
  cellName: string;
  /**
   * Starter plain-text content rendered by the cell.
   */
  value: number | string;
};

/**
 * Renders one plain-text alerts table cell used by starter field cells.
 */
export const AlertsTableTextCell = ({
  cellName,
  value,
}: AlertsTableTextCellProps): ReactElement => {
  return (
    <TableCell data-test-class="alerts-table-cell" data-test-name={cellName}>
      {value}
    </TableCell>
  );
};
