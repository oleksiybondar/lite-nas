import { useLanguage } from "@lite-nas/language";
import LogoutRoundedIcon from "@mui/icons-material/LogoutRounded";
import ManageAccountsRoundedIcon from "@mui/icons-material/ManageAccountsRounded";
import SettingsRoundedIcon from "@mui/icons-material/SettingsRounded";
import Avatar from "@mui/material/Avatar";
import Box from "@mui/material/Box";
import Divider from "@mui/material/Divider";
import IconButton from "@mui/material/IconButton";
import ListItemIcon from "@mui/material/ListItemIcon";
import ListItemText from "@mui/material/ListItemText";
import Menu from "@mui/material/Menu";
import MenuItem from "@mui/material/MenuItem";
import Stack from "@mui/material/Stack";
import Tooltip from "@mui/material/Tooltip";
import Typography from "@mui/material/Typography";
import type { MouseEvent, ReactElement } from "react";
import { useState } from "react";

/** Identity and actions provided by an authenticated application. */
export interface UserMenuProps {
  login: string;
  fullName?: string | undefined;
  avatarUrl?: string | null | undefined;
  onProfile: () => void;
  onApplicationSettings: () => void;
  onLogout: () => void;
}

/** Renders the shared avatar button and authenticated user dropdown. */
export const UserMenu = ({
  login,
  fullName,
  avatarUrl,
  onProfile,
  onApplicationSettings,
  onLogout,
}: UserMenuProps): ReactElement => {
  const { getLabel } = useLanguage();
  const [anchorElement, setAnchorElement] = useState<HTMLElement | null>(null);
  const closeAndRun = (action: () => void): void => {
    setAnchorElement(null);
    action();
  };

  return (
    <>
      <Tooltip title={getLabel("profile.userMenu")}>
        <IconButton
          aria-label={getLabel("profile.userMenu")}
          color="inherit"
          data-testid="user-menu-button"
          onClick={(event: MouseEvent<HTMLButtonElement>) => setAnchorElement(event.currentTarget)}
          sx={{ borderRadius: "8px", p: 0.5 }}
        >
          <Stack alignItems="center" direction="row" spacing={1}>
            <Avatar
              alt={fullName ?? login}
              data-testid="user-avatar"
              src={avatarUrl ?? undefined}
              sx={{ height: 34, width: 34 }}
            >
              {initials(fullName ?? login)}
            </Avatar>
            <Box
              data-testid="user-menu-summary"
              display={{ sm: "block", xs: "none" }}
              textAlign="left"
            >
              <Typography data-testid="user-menu-login" lineHeight={1.2} variant="body2">
                {login}
              </Typography>
              {fullName ? (
                <Typography
                  color="text.secondary"
                  data-testid="user-menu-full-name"
                  lineHeight={1.2}
                  variant="caption"
                >
                  {fullName}
                </Typography>
              ) : null}
            </Box>
          </Stack>
        </IconButton>
      </Tooltip>
      <Menu
        anchorEl={anchorElement}
        anchorOrigin={{ horizontal: "right", vertical: "bottom" }}
        data-testid="user-menu"
        onClose={() => setAnchorElement(null)}
        open={anchorElement !== null}
        transformOrigin={{ horizontal: "right", vertical: "top" }}
      >
        <MenuItem data-testid="user-menu-profile-link" onClick={() => closeAndRun(onProfile)}>
          <ListItemIcon>
            <ManageAccountsRoundedIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText primary={getLabel("profile.userProfile")} />
        </MenuItem>
        <MenuItem
          data-testid="user-menu-application-settings-link"
          onClick={() => closeAndRun(onApplicationSettings)}
        >
          <ListItemIcon>
            <SettingsRoundedIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText primary={getLabel("profile.applicationSettings")} />
        </MenuItem>
        <Divider />
        <MenuItem data-testid="user-menu-logout-button" onClick={() => closeAndRun(onLogout)}>
          <ListItemIcon>
            <LogoutRoundedIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText primary={getLabel("profile.logout")} />
        </MenuItem>
      </Menu>
    </>
  );
};

/** Builds an accessible avatar fallback from the user's display name. */
const initials = (name: string): string => {
  const source = name.trim();
  if (!source) return "?";
  return (
    source
      .split(/\s+/)
      .slice(0, 2)
      .map((part) => part[0]?.toUpperCase() ?? "")
      .join("") || "?"
  );
};
