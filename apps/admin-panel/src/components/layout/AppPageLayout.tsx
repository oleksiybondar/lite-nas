import { AppTopBar } from "@components/navigation/AppTopBar";
import Box from "@mui/material/Box";
import Container from "@mui/material/Container";
import type { PropsWithChildren, ReactElement } from "react";

/**
 * Simple top-bar page layout with a dedicated scrolling work area.
 */
export const AppPageLayout = ({ children }: PropsWithChildren): ReactElement => {
  return (
    <Box
      data-testid="app-page-layout"
      display="flex"
      flexDirection="column"
      height="100dvh"
      overflow="hidden"
    >
      <AppTopBar />
      <Container
        component="main"
        data-testid="app-page-content"
        maxWidth={false}
        sx={{
          "& > *": {
            minWidth: 0,
            width: "100%",
          },
          flex: 1,
          minHeight: 0,
          overflowY: "auto",
          py: 4,
        }}
      >
        {children}
      </Container>
    </Box>
  );
};
