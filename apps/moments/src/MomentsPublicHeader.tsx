import StorageRoundedIcon from "@mui/icons-material/StorageRounded";
import AppBar from "@mui/material/AppBar";
import Toolbar from "@mui/material/Toolbar";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

/** Shows the LiteNAS brand in the same top-bar position used by the admin panel. */
export const MomentsPublicHeader = (): ReactElement => (
  <AppBar color="transparent" data-testid="app-top-bar" elevation={0} position="sticky">
    <Toolbar sx={{ borderBottom: 1, borderColor: "divider", gap: 1.25 }}>
      <StorageRoundedIcon color="primary" fontSize="small" />
      <Typography fontWeight={700} variant="subtitle1">
        LiteNAS
      </Typography>
    </Toolbar>
  </AppBar>
);
