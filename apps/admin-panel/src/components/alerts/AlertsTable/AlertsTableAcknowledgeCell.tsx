import { useAlerts } from "@hooks/useAlerts";
import Button from "@mui/material/Button";
import TableCell from "@mui/material/TableCell";
import type { ReactElement } from "react";

type AlertsTableAcknowledgeCellProps = {
  /**
   * Event identifier used to resolve the alert item from shared alerts context.
   */
  eventId: string;
};

/**
 * Renders the acknowledge action cell for one alert event.
 */
export const AlertsTableAcknowledgeCell = ({
  eventId,
}: AlertsTableAcknowledgeCellProps): ReactElement => {
  const { acknowledge, isAcknowledging, items } = useAlerts();
  const item = items.find((currentItem) => currentItem.EventID === eventId);
  const isDisabled = isAcknowledging || item === undefined || item.Acknowledged;

  return (
    <TableCell data-test-class="alerts-table-cell" data-test-name="acknowledge">
      <Button
        data-testid={`alerts-acknowledge-button-${eventId}`}
        disabled={isDisabled}
        onClick={() => {
          void acknowledge(eventId);
        }}
        size="small"
        variant="outlined"
      >
        Acknowledge
      </Button>
    </TableCell>
  );
};
