import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

/**
 * Shared footer rendered by public and protected admin-panel layouts.
 */
export const AppFooter = (): ReactElement => {
  return (
    <Box
      borderColor="divider"
      borderTop={1}
      component="footer"
      data-testid="app-footer"
      px={3}
      py={1.5}
      textAlign="center"
    >
      <Typography color="text.secondary" data-testid="app-footer-title" variant="caption">
        LiteNAS
      </Typography>
    </Box>
  );
};
