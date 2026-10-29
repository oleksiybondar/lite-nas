import { useLanguage } from "@lite-nas/language";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import type { ReactElement, ReactNode } from "react";

/** Frames Moments screens with the admin panel's header, content, and footer structure. */
export const MomentsLayout = ({
  header,
  children,
}: {
  header: ReactNode;
  children: ReactNode;
}): ReactElement => {
  const { getLabel } = useLanguage();
  return (
    <Box
      data-testid="moments-app-layout"
      display="flex"
      flexDirection="column"
      height="100dvh"
      overflow="hidden"
    >
      {header}
      <Box
        component="main"
        data-testid="moments-app-main"
        display="flex"
        flex={1}
        minHeight={0}
        overflow="auto"
      >
        {children}
      </Box>
      <Box
        borderColor="divider"
        borderTop={1}
        component="footer"
        data-testid="app-footer"
        px={3}
        py={1.5}
        textAlign="center"
      >
        <Typography color="text.secondary" variant="caption">
          LiteNAS · {getLabel("moments.footer.locationData")}{" "}
          <a href="https://www.geonames.org/" rel="noreferrer" target="_blank">
            GeoNames
          </a>
          {" · "}
          <a href="https://www.openstreetmap.org/copyright" rel="noreferrer" target="_blank">
            OpenStreetMap contributors
          </a>
        </Typography>
      </Box>
    </Box>
  );
};
