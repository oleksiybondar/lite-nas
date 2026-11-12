import { TaskPage } from "@pages/SecurityTasks/TaskPage";
import type { ReactElement } from "react";

/** Security page for selecting a directory and monitoring antivirus scans. */
export const AntivirusPage = (): ReactElement | null => <TaskPage kind="antivirus" />;
