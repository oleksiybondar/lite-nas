import Box from "@mui/material/Box";
import type { ReactElement, ReactNode } from "react";

/** Frames Moments screens with the admin panel's header, content, and footer structure. */
export const MomentsLayout = ({
  header,
  children,
}: {
  header: ReactNode;
  children: ReactNode;
}): ReactElement => {
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
    </Box>
  );
};
