import type { SxProps, Theme } from "@mui/material/styles";
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
  /**
   * Optional table-cell styling used by sticky or aligned cells.
   */
  sx?: SxProps<Theme>;
};

/**
 * Renders one plain-text alerts table cell used by starter field cells.
 */
export const AlertsTableTextCell = ({
  cellName,
  sx,
  value,
}: AlertsTableTextCellProps): ReactElement => {
  return (
    <TableCell data-test-class="alerts-table-cell" data-test-name={cellName} sx={sx}>
      {value}
    </TableCell>
  );
};
