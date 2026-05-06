import { useAuth } from "@hooks/useAuth";
import LogoutRoundedIcon from "@mui/icons-material/LogoutRounded";
import ManageAccountsRoundedIcon from "@mui/icons-material/ManageAccountsRounded";
import SettingsRoundedIcon from "@mui/icons-material/SettingsRounded";
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
import { Link as RouterLink } from "react-router-dom";
import { UserAvatar } from "./UserAvatar";

/**
 * Header user control for authenticated dashboard sessions.
 */
export const AppUserMenu = (): ReactElement | null => {
  const { isAuthenticated, logout, me } = useAuth();
  const [anchorElement, setAnchorElement] = useState<HTMLElement | null>(null);

  if (!isAuthenticated || me === null) {
    return null;
  }

  const closeMenu = (): void => {
    setAnchorElement(null);
  };

  const handleLogout = (): void => {
    closeMenu();
    void logout();
  };

  return (
    <>
      <Tooltip title="User menu">
        <IconButton
          aria-label="User menu"
          color="inherit"
          onClick={(event: MouseEvent<HTMLButtonElement>) => {
            setAnchorElement(event.currentTarget);
          }}
          sx={{ borderRadius: "8px", p: 0.5 }}
        >
          <Stack alignItems="center" direction="row" spacing={1}>
            <UserAvatar user={me.user} />
            <Box display={{ sm: "block", xs: "none" }} textAlign="left">
              <Typography lineHeight={1.2} variant="body2">
                {me.user.login}
              </Typography>
              {me.user.full_name !== undefined ? (
                <Typography color="text.secondary" lineHeight={1.2} variant="caption">
                  {me.user.full_name}
                </Typography>
              ) : null}
            </Box>
          </Stack>
        </IconButton>
      </Tooltip>
      <Menu
        anchorEl={anchorElement}
        anchorOrigin={{ horizontal: "right", vertical: "bottom" }}
        onClose={closeMenu}
        open={anchorElement !== null}
        transformOrigin={{ horizontal: "right", vertical: "top" }}
      >
        <MenuItem component={RouterLink} onClick={closeMenu} to="/preferences/profile">
          <ListItemIcon>
            <ManageAccountsRoundedIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText primary="User profile" />
        </MenuItem>
        <MenuItem component={RouterLink} onClick={closeMenu} to="/preferences/application">
          <ListItemIcon>
            <SettingsRoundedIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText primary="Application settings" />
        </MenuItem>
        <Divider />
        <MenuItem onClick={handleLogout}>
          <ListItemIcon>
            <LogoutRoundedIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText primary="Logout" />
        </MenuItem>
      </Menu>
    </>
  );
};
