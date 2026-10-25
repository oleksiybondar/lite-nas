import { UserMenu } from "@lite-nas/profile";
import AppBar from "@mui/material/AppBar";
import Button from "@mui/material/Button";
import Toolbar from "@mui/material/Toolbar";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

/** Shows Moments navigation and the same user dropdown used by the admin panel. */
export const MomentsHeader = ({
  username,
  fullName,
  avatarUrl,
  onHome,
  onThumbnailMaker,
  onProfile,
  onApplicationSettings,
  onLogout,
}: {
  username: string;
  fullName?: string;
  avatarUrl?: string | null;
  onHome: () => void;
  onThumbnailMaker: () => void;
  onProfile: () => void;
  onApplicationSettings: () => void;
  onLogout: () => void;
}): ReactElement => (
  <AppBar color="transparent" data-testid="app-top-bar" elevation={0} position="sticky">
    <Toolbar sx={{ borderBottom: 1, borderColor: "divider", gap: 2 }}>
      <Typography data-testid="moments-title" sx={{ flexGrow: 1 }} variant="h6">
        Moments
      </Typography>
      <Button data-testid="moments-home-button" onClick={onHome}>
        Home
      </Button>
      <Button onClick={onThumbnailMaker}>Thumbnail maker</Button>
      <UserMenu
        login={username}
        fullName={fullName}
        avatarUrl={avatarUrl}
        onProfile={onProfile}
        onApplicationSettings={onApplicationSettings}
        onLogout={onLogout}
      />
    </Toolbar>
  </AppBar>
);
