import StorageRoundedIcon from "@mui/icons-material/StorageRounded";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import { Link as RouterLink } from "react-router-dom";

/**
 * Brand mark shown in the application header.
 *
 * The logo links to the protected dashboard root. Anonymous users still remain
 * on the login screen because the auth guard owns access to protected content.
 */
export const AppLogo = (): ReactElement => {
  return (
    <Stack
      alignItems="center"
      component={RouterLink}
      data-testid="app-logo-link"
      direction="row"
      spacing={1.25}
      sx={{ color: "inherit", textDecoration: "none" }}
      to="/"
    >
      <StorageRoundedIcon color="primary" data-testid="app-logo-icon" fontSize="small" />
      <Typography
        component="span"
        data-testid="app-logo-title"
        fontWeight={700}
        variant="subtitle1"
      >
        LiteNAS
      </Typography>
    </Stack>
  );
};
