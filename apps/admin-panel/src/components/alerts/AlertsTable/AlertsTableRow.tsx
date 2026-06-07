import { AlertsTableAcknowledgeCell } from "@components/alerts/AlertsTable/AlertsTableAcknowledgeCell";
import { AlertsTableAcknowledgedAtCell } from "@components/alerts/AlertsTable/AlertsTableAcknowledgedAtCell";
import { AlertsTableCategoryCell } from "@components/alerts/AlertsTable/AlertsTableCategoryCell";
import { AlertsTableCreatedAtCell } from "@components/alerts/AlertsTable/AlertsTableCreatedAtCell";
import { AlertsTableEventIdCell } from "@components/alerts/AlertsTable/AlertsTableEventIdCell";
import { AlertsTableMessageCell } from "@components/alerts/AlertsTable/AlertsTableMessageCell";
import { AlertsTableMitigateCell } from "@components/alerts/AlertsTable/AlertsTableMitigateCell";
import { AlertsTablePriorityCell } from "@components/alerts/AlertsTable/AlertsTablePriorityCell";
import { AlertsTableSeverityCell } from "@components/alerts/AlertsTable/AlertsTableSeverityCell";
import { AlertsTableSourceCell } from "@components/alerts/AlertsTable/AlertsTableSourceCell";
import { AlertsTableStatusCell } from "@components/alerts/AlertsTable/AlertsTableStatusCell";
import { AlertsTableValueCell } from "@components/alerts/AlertsTable/AlertsTableValueCell";
import type { AlertDomain, AlertListItemDTO } from "@dto/alerts/alerts";
import TableRow from "@mui/material/TableRow";
import type { ReactElement } from "react";

type AlertsTableRowProps = {
  /**
   * Alert domain controlling security-only columns.
   */
  domain: AlertDomain;
  /**
   * One alert item rendered in the table row.
   */
  item: AlertListItemDTO;
};

/**
 * Renders one alerts table row composed from dedicated cell components.
 */
export const AlertsTableRow = ({ domain, item }: AlertsTableRowProps): ReactElement => {
  return (
    <TableRow
      data-test-class="alerts-table-row"
      data-test-name={`${item.EventID}:${item.EventRecID}`}
      hover
    >
      <AlertsTableSeverityCell item={item} />
      <AlertsTablePriorityCell item={item} />
      <AlertsTableEventIdCell item={item} />
      <AlertsTableMessageCell item={item} />
      <AlertsTableValueCell item={item} />
      <AlertsTableSourceCell item={item} />
      <AlertsTableStatusCell item={item} />
      {domain === "security" ? <AlertsTableMitigateCell item={item} /> : null}
      <AlertsTableAcknowledgeCell eventId={item.EventID} />
      <AlertsTableCategoryCell item={item} />
      <AlertsTableCreatedAtCell item={item} />
      <AlertsTableAcknowledgedAtCell item={item} />
    </TableRow>
  );
};
