import { CategoryLandingPage } from "@components/navigation/CategoryLandingPage";
import FolderOpenIcon from "@mui/icons-material/FolderOpen";
import Box from "@mui/material/Box";
import type { ReactElement } from "react";

/** Lists installed applications as full-card navigation targets. */
export const ApplicationsLandingPage = (): ReactElement => (
  <Box data-testid="applications-landing-page">
    <CategoryLandingPage
      cards={[
        {
          description: "Browse files and upload to directories where you have permission.",
          icon: <FolderOpenIcon />,
          path: "/applications/file-explorer",
          title: "File Explorer",
        },
      ]}
      overline="Applications"
      summary="Open installed LiteNAS applications."
      title="Applications"
    />
  </Box>
);
