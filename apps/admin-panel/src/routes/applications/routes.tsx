import { ApplicationsLandingPage } from "@pages/ApplicationsLandingPage";
import { FileExplorerEmbeddedPage } from "@pages/FileExplorerEmbeddedPage";
import { MomentsEmbeddedPage } from "@pages/MomentsEmbeddedPage";
import type { RouteObject } from "react-router-dom";

/** Admin panel routes that embed separately served LiteNAS applications. */
export const applicationsRoutes: RouteObject[] = [
  { path: "/applications", element: <ApplicationsLandingPage /> },
  { path: "/applications/file-explorer", element: <FileExplorerEmbeddedPage /> },
  { path: "/applications/moments", element: <MomentsEmbeddedPage /> },
];
