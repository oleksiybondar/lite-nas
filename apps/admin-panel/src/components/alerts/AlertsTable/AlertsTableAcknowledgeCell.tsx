import { formatAcknowledgedByValue } from "@components/alerts/AlertsTable/helpers";
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
 * Renders the acknowledgement state cell for one alert event.
 *
 * Non-acknowledged rows expose the acknowledge action. Acknowledged rows render
 * the actor name instead of an inactive action button.
 */
export const AlertsTableAcknowledgeCell = ({
  eventId,
}: AlertsTableAcknowledgeCellProps): ReactElement => {
  const { acknowledge, isAcknowledging, items } = useAlerts();
  const item = items.find((currentItem) => currentItem.EventID === eventId);

  if (item?.Acknowledged) {
    return (
      <TableCell
        data-test-class="alerts-table-cell"
        data-test-name="acknowledgement"
        data-test-tone="primary"
        sx={{ color: "primary.main", fontWeight: 600 }}
      >
        {formatAcknowledgedByValue(item.AcknowledgedBy)}
      </TableCell>
    );
  }

  return (
    <TableCell data-test-class="alerts-table-cell" data-test-name="acknowledgement">
      <Button
        data-testid={`alerts-acknowledge-button-${eventId}`}
        disabled={isAcknowledging || item === undefined}
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
