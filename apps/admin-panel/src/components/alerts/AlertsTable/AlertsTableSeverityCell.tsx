import {
  buildAlertsStickyColumnSx,
  formatAlertSeverityLabel,
} from "@components/alerts/AlertsTable/helpers";
import type { AlertListItemDTO } from "@dto/alerts/alerts";
import ErrorOutlineIcon from "@mui/icons-material/ErrorOutline";
import InfoOutlinedIcon from "@mui/icons-material/InfoOutlined";
import PriorityHighIcon from "@mui/icons-material/PriorityHigh";
import WarningAmberIcon from "@mui/icons-material/WarningAmber";
import Stack from "@mui/material/Stack";
import TableCell from "@mui/material/TableCell";
import Tooltip from "@mui/material/Tooltip";
import type { ReactElement } from "react";

type AlertsTableSeverityCellProps = {
  /**
   * Alert item supplying the severity label.
   */
  item: AlertListItemDTO;
};

/**
 * Renders the typed severity cell with icon and color treatment.
 */
export const AlertsTableSeverityCell = ({ item }: AlertsTableSeverityCellProps): ReactElement => {
  return (
    <TableCell
      data-test-class="alerts-table-cell"
      data-test-name="severity"
      sx={buildAlertsStickyColumnSx("severity")}
    >
      <Tooltip title={formatAlertSeverityLabel(item.Severity)}>
        <Stack
          alignItems="center"
          color={resolveSeverityColor(item.Severity)}
          data-testid={`alerts-severity-cell-${item.EventID}`}
          direction="row"
          justifyContent="center"
        >
          {renderSeverityIcon(item.Severity)}
        </Stack>
      </Tooltip>
    </TableCell>
  );
};

const renderSeverityIcon = (severity: AlertListItemDTO["Severity"]): ReactElement => {
  if (severity === "critical") {
    return <PriorityHighIcon fontSize="small" />;
  }

  if (severity === "error") {
    return <ErrorOutlineIcon fontSize="small" />;
  }

  if (severity === "warning") {
    return <WarningAmberIcon fontSize="small" />;
  }

  return <InfoOutlinedIcon fontSize="small" />;
};

const resolveSeverityColor = (severity: AlertListItemDTO["Severity"]): string => {
  if (severity === "critical" || severity === "error") {
    return "error.main";
  }

  if (severity === "warning") {
    return "warning.main";
  }

  return "info.main";
};
