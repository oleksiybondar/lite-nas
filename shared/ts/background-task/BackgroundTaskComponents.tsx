import { Button, LinearProgress, Stack } from "@mui/material";
import type { ReactElement, ReactNode } from "react";
import { useBackgroundTask } from "./BackgroundTaskContext";

/** Props for reusable start and optional cancel actions. */
export interface BackgroundTaskActionsProps {
  /** Accessible text on the start button. */
  startLabel: string;
  /** Accessible text on the optional cancel button. */
  cancelLabel?: string;
  /** True while the tracked task is active. */
  active: boolean;
  /** Disables actions while surrounding task history is loading. */
  disabled?: boolean;
  /** Starts or retries the service-owned task. */
  onStart: () => void;
  /** Requests service-owned cancellation when provided. */
  onCancel?: () => void;
}

/** Props for the reusable active-task progress indicator and status region. */
export interface BackgroundTaskProgressProps {
  /** True while the tracked task is active. */
  active: boolean;
  /** Accessible progress label supplied by the application. */
  label: string;
  /** Measured percentage, or null when only indeterminate progress is known. */
  value?: number | null;
  /** Application-owned status text, counters, or result details. */
  children?: ReactNode;
}

/** Renders localized-by-caller start and optional cancel controls. */
export function BackgroundTaskActions({
  startLabel,
  cancelLabel,
  active,
  disabled = false,
  onStart,
  onCancel,
}: BackgroundTaskActionsProps): ReactElement {
  const { busy } = useBackgroundTask<unknown>();
  return (
    <Stack direction="row" spacing={1}>
      <Button disabled={disabled || busy || active} onClick={onStart}>
        {startLabel}
      </Button>
      {active && cancelLabel && onCancel && (
        <Button disabled={busy} onClick={onCancel}>
          {cancelLabel}
        </Button>
      )}
    </Stack>
  );
}

/** Shows task details and indeterminate or measured active-task progress. */
export function BackgroundTaskProgress({
  active,
  label,
  value = null,
  children,
}: BackgroundTaskProgressProps): ReactElement {
  const percent = value === null ? null : Math.max(0, Math.min(100, value));
  return (
    <Stack spacing={1}>
      {children}
      {active && (
        <LinearProgress
          aria-label={label}
          variant={percent === null ? "indeterminate" : "determinate"}
          value={percent ?? undefined}
        />
      )}
    </Stack>
  );
}
