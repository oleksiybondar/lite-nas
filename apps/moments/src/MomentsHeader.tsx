import { LanguageSelector, useLanguage } from "@lite-nas/language";
import { UserMenu } from "@lite-nas/profile";
import AppBar from "@mui/material/AppBar";
import Button from "@mui/material/Button";
import Link from "@mui/material/Link";
import Toolbar from "@mui/material/Toolbar";
import type { ReactElement } from "react";
import { Link as RouterLink } from "react-router-dom";

/** Shows Moments navigation and the same user dropdown used by the admin panel. */
export const MomentsHeader = ({
  username,
  fullName,
  avatarUrl,
  onMine,
  onShared,
  onProfile,
  onApplicationSettings,
  onLogout,
}: {
  username: string;
  fullName?: string;
  avatarUrl?: string | null;
  onMine: () => void;
  onShared: () => void;
  onProfile: () => void;
  onApplicationSettings: () => void;
  onLogout: () => void;
}): ReactElement => {
  const { getLabel } = useLanguage();
  return (
    <AppBar color="transparent" data-testid="app-top-bar" elevation={0} position="sticky">
      <Toolbar sx={{ borderBottom: 1, borderColor: "divider", flexWrap: "wrap", gap: 2 }}>
        <Link
          color="inherit"
          component={RouterLink}
          data-testid="moments-title"
          to="/"
          variant="h6"
          underline="none"
          sx={{ flexGrow: 1 }}
        >
          Moments
        </Link>
        <Button component={RouterLink} data-testid="moments-home-button" to="/">
          {getLabel("common.home")}
        </Button>
        <Button onClick={onMine}>{getLabel("moments.myCollections")}</Button>
        <Button onClick={onShared}>{getLabel("moments.sharedCollections")}</Button>
        <LanguageSelector />
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
};
