import { SessionsPageContent } from "@pages/SessionsPage/SessionsPageContent";
import { SessionsProvider } from "@providers/SessionsProvider";
import type { ReactElement } from "react";

/** Security-only host session inventory and management page. */
export const SessionsPage = (): ReactElement => (
  <SessionsProvider>
    <SessionsPageContent />
  </SessionsProvider>
);
