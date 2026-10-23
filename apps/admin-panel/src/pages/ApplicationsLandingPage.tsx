import { CategoryLandingPage } from "@components/navigation/CategoryLandingPage";
import { useMomentsAvailability } from "@hooks/useMomentsAvailability";
import FolderOpenIcon from "@mui/icons-material/FolderOpen";
import PhotoLibraryIcon from "@mui/icons-material/PhotoLibrary";
import Box from "@mui/material/Box";
import type { ReactElement } from "react";

/** Lists installed applications as full-card navigation targets. */
export const ApplicationsLandingPage = (): ReactElement => {
  const momentsAvailable = useMomentsAvailability();

  return (
    <Box data-testid="applications-landing-page">
      <CategoryLandingPage
        cards={[
          {
            description: "Browse files and upload to directories where you have permission.",
            icon: <FolderOpenIcon />,
            path: "/applications/file-explorer",
            title: "File Explorer",
          },
          ...(momentsAvailable
            ? [
                {
                  description: "Open your Moments library.",
                  icon: <PhotoLibraryIcon />,
                  path: "/applications/moments",
                  title: "Moments",
                },
              ]
            : []),
        ]}
        overline="Applications"
        summary="Open installed LiteNAS applications."
        title="Applications"
      />
    </Box>
  );
};
