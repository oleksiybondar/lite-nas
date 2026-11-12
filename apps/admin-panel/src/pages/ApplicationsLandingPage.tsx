import { CategoryLandingPage } from "@components/navigation/CategoryLandingPage";
import { useMomentsAvailability } from "@hooks/useMomentsAvailability";
import { useLanguage } from "@lite-nas/language";
import FolderOpenIcon from "@mui/icons-material/FolderOpen";
import PhotoLibraryIcon from "@mui/icons-material/PhotoLibrary";
import Box from "@mui/material/Box";
import type { ReactElement } from "react";

/** Lists installed applications as full-card navigation targets. */
export const ApplicationsLandingPage = (): ReactElement => {
  const momentsAvailable = useMomentsAvailability();
  const { getLabel } = useLanguage();

  return (
    <Box data-testid="applications-landing-page">
      <CategoryLandingPage
        cards={[
          {
            description: getLabel("admin.landing.applicationsExplorer"),
            icon: <FolderOpenIcon />,
            path: "/applications/file-explorer",
            title: getLabel("navigation.File Explorer"),
          },
          ...(momentsAvailable
            ? [
                {
                  description: getLabel("admin.landing.applicationsMoments"),
                  icon: <PhotoLibraryIcon />,
                  path: "/applications/moments",
                  title: getLabel("navigation.Moments"),
                },
              ]
            : []),
        ]}
        overline={getLabel("navigation.Applications")}
        summary={getLabel("admin.landing.applicationsSummary")}
        title={getLabel("navigation.Applications")}
      />
    </Box>
  );
};
