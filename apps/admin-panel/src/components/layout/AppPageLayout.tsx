import { AppTopBar } from "@components/navigation/AppTopBar";

import Box from "@mui/material/Box";
import Container from "@mui/material/Container";
import type { PropsWithChildren, ReactElement } from "react";

export const AppPageLayout = ({ children }: PropsWithChildren): ReactElement => {
  return (
    <Box data-testid="app-page-layout" sx={{ minHeight: "100vh" }}>
      <AppTopBar />
      <Container component="main" data-testid="app-page-content" sx={{ py: 4 }}>
        {children}
      </Container>
    </Box>
  );
};
