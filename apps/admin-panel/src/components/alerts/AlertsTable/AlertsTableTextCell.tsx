import type { SxProps, Theme } from "@mui/material/styles";
import TableCell from "@mui/material/TableCell";
import type { ReactElement } from "react";

type AlertsTableTextCellTone = "default" | "error" | "primary" | "success" | "warning";

type AlertsTableTextCellProps = {
  /**
   * Stable cell discriminator for test selectors.
   */
  cellName: string;
  /**
   * Optional semantic tone used to highlight important values.
   */
  tone?: AlertsTableTextCellTone;
  /**
   * Plain-text content rendered by the cell.
   */
  value: number | string;
  /**
   * Optional table-cell styling used by sticky or aligned cells.
   */
  sx?: SxProps<Theme>;
};

/**
 * Renders one plain-text alerts table cell used by field-specific cells.
 */
export const AlertsTableTextCell = ({
  cellName,
  sx,
  tone = "default",
  value,
}: AlertsTableTextCellProps): ReactElement => {
  return (
    <TableCell
      data-test-class="alerts-table-cell"
      data-test-name={cellName}
      data-test-tone={tone}
      sx={buildTextCellSx(tone, sx)}
    >
      {value}
    </TableCell>
  );
};

/**
 * Applies semantic text emphasis while preserving caller-provided table cell styling.
 */
const buildTextCellSx = (tone: AlertsTableTextCellTone, sx?: SxProps<Theme>): SxProps<Theme> => {
  return [resolveTextCellToneSx(tone), sx];
};

/**
 * Resolves the palette color and weight used for one semantic cell tone.
 */
const resolveTextCellToneSx = (tone: AlertsTableTextCellTone): SxProps<Theme> => {
  switch (tone) {
    case "primary":
      return { color: "primary.main", fontWeight: 600 };
    case "success":
      return { color: "success.main", fontWeight: 600 };
    case "warning":
      return { color: "warning.main", fontWeight: 600 };
    case "error":
      return { color: "error.main", fontWeight: 600 };
    default:
      return {};
  }
};
