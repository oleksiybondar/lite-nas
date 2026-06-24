import type { AlertDomain, AlertOccurrenceItemDTO } from "@dto/alerts/alerts";
import type { ReactNode } from "react";

/**
 * Public props accepted by the occurrences dialog entry component.
 */
export type AlertOccurrencesDialogProps = {
  /**
   * Alert domain owning the occurrences endpoint.
   */
  domain: AlertDomain;
  /**
   * Stable alert event identifier used to fetch the occurrence history.
   */
  eventId: string;
  /**
   * Icon shown before the trigger label to communicate chart or table mode.
   */
  triggerIcon: ReactNode;
  /**
   * Current last-value text rendered inside the trigger button.
   */
  triggerLabel: string;
};

/**
 * Layout props shared by occurrences dialog content variants.
 */
export type AlertOccurrencesDialogLayoutProps = {
  /**
   * Dialog body content rendered under the title and summary.
   */
  children: ReactNode;
  /**
   * Invoked when the user closes the dialog.
   */
  onClose: () => void;
  /**
   * Whether the dialog is currently open.
   */
  open: boolean;
  /**
   * Secondary summary shown under the dialog title.
   */
  summary: string;
  /**
   * Primary title shown in the dialog header.
   */
  title: string;
};

/**
 * Props accepted by the numeric occurrences chart.
 */
export type AlertOccurrencesChartProps = {
  /**
   * Numeric occurrence rows plotted in chronological order.
   */
  occurrences: AlertOccurrenceItemDTO[];
};

/**
 * Props accepted by the non-numeric occurrences table.
 */
export type AlertOccurrencesTableProps = {
  /**
   * Non-numeric occurrence rows rendered with local search and pagination.
   */
  occurrences: AlertOccurrenceItemDTO[];
};
