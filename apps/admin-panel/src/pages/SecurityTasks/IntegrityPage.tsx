import { TaskPage } from "@pages/SecurityTasks/TaskPage";
import type { ReactElement } from "react";

/** Security page for starting and monitoring AIDE integrity checks. */
export const IntegrityPage = (): ReactElement | null => (
  <>
    <TaskPage kind="integrity" />
    <TaskPage kind="baseline" />
  </>
);
