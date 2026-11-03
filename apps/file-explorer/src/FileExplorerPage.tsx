import type { ReactElement } from "react";

import { FileExplorerProvider } from "./FileExplorerContext";
import { FileExplorerWorkspace } from "./FileExplorerWorkspace";

/** Mounts the two-panel state for the File Explorer application. */
export function FileExplorerPage(): ReactElement {
  return (
    <FileExplorerProvider>
      <FileExplorerWorkspace />
    </FileExplorerProvider>
  );
}
