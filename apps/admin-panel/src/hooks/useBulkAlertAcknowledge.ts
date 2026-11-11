import { useLanguage } from "@lite-nas/language";
import { useRef, useState } from "react";

type Notification = { id: number; message: string; severity: "success" | "warning" | "error" };
type Result = { acknowledged: number; failed: number };

/** Converts a rejected bulk request into an all-failed notification result. */
const acknowledgeOutcome = async (
  ids: string[],
  acknowledgeMany: (ids: string[]) => Promise<Result>,
): Promise<Result> => {
  try {
    return await acknowledgeMany(ids);
  } catch {
    return { acknowledged: 0, failed: ids.length };
  }
};

/** Formats the outcome of one bulk acknowledgement for the page notification. */
const describeResult = (
  result: Result,
  getLabel: (key: string) => string,
): Omit<Notification, "id"> => {
  let severity: Notification["severity"] = "error";
  let label = "admin.alerts.bulkFailed";
  if (result.failed === 0) {
    severity = "success";
    label = "admin.alerts.bulkSuccess";
  } else if (result.acknowledged > 0) {
    severity = "warning";
    label = "admin.alerts.bulkPartial";
  }
  return {
    severity,
    message: getLabel(label)
      .replace("{acknowledged}", String(result.acknowledged))
      .replace("{failed}", String(result.failed)),
  };
};

/** Owns pending state and notifications for the page's bulk acknowledgement action. */
export const useBulkAlertAcknowledge = (
  ids: string[],
  acknowledgeMany: (ids: string[]) => Promise<Result>,
) => {
  const { getLabel } = useLanguage();
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [isBulkAcknowledging, setIsBulkAcknowledging] = useState(false);
  const nextNotificationId = useRef(0);
  const bulkInProgress = useRef(false);

  const notify = (result: Result): void => {
    setNotifications((current) => [
      ...current,
      { id: nextNotificationId.current++, ...describeResult(result, getLabel) },
    ]);
  };
  const acknowledgeAll = async (): Promise<void> => {
    if (bulkInProgress.current || ids.length === 0) return;
    bulkInProgress.current = true;
    setIsBulkAcknowledging(true);
    const result = await acknowledgeOutcome(ids, acknowledgeMany);
    notify(result);
    bulkInProgress.current = false;
    setIsBulkAcknowledging(false);
  };
  const dismiss = (id: number): void =>
    setNotifications((current) => current.filter((item) => item.id !== id));
  return { notifications, isBulkAcknowledging, acknowledgeAll, dismiss };
};
