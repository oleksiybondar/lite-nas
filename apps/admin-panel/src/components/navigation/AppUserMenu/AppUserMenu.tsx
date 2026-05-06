import type { AuthMeUserDTO } from "@dto/auth/auth";
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
      {renderUserMenuButton(me.user, setAnchorElement)}
      {renderUserMenu({
        anchorElement,
        onClose: closeMenu,
        onLogout: handleLogout,
      })}
    </>
  );
};

/**
 * Builds the button that opens the authenticated user menu.
 */
const renderUserMenuButton = (
  user: AuthMeUserDTO,
  setAnchorElement: (anchorElement: HTMLElement | null) => void,
): ReactElement => {
  return (
    <Tooltip title="User menu">
      <IconButton
        aria-label="User menu"
        color="inherit"
        data-testid="user-menu-button"
        onClick={(event: MouseEvent<HTMLButtonElement>) => {
          setAnchorElement(event.currentTarget);
        }}
        sx={{ borderRadius: "8px", p: 0.5 }}
      >
        <Stack alignItems="center" direction="row" spacing={1}>
          <UserAvatar user={user} />
          {renderUserMenuSummary(user)}
        </Stack>
      </IconButton>
    </Tooltip>
  );
};

/**
 * Builds the visible user identity text inside the menu button.
 */
const renderUserMenuSummary = (user: AuthMeUserDTO): ReactElement => {
  return (
    <Box data-testid="user-menu-summary" display={{ sm: "block", xs: "none" }} textAlign="left">
      <Typography data-testid="user-menu-login" lineHeight={1.2} variant="body2">
        {user.login}
      </Typography>
      {user.full_name !== undefined ? (
        <Typography
          color="text.secondary"
          data-testid="user-menu-full-name"
          lineHeight={1.2}
          variant="caption"
        >
          {user.full_name}
        </Typography>
      ) : null}
    </Box>
  );
};

/**
 * State and commands required to render the authenticated user menu.
 */
type UserMenuRenderOptions = {
  anchorElement: HTMLElement | null;
  onClose: () => void;
  onLogout: () => void;
};

/**
 * Builds the authenticated user action menu.
 */
const renderUserMenu = ({
  anchorElement,
  onClose,
  onLogout,
}: UserMenuRenderOptions): ReactElement => {
  return (
    <Menu
      anchorEl={anchorElement}
      anchorOrigin={{ horizontal: "right", vertical: "bottom" }}
      data-testid="user-menu"
      onClose={onClose}
      open={anchorElement !== null}
      transformOrigin={{ horizontal: "right", vertical: "top" }}
    >
      {renderUserMenuLink({
        icon: <ManageAccountsRoundedIcon fontSize="small" />,
        label: "User profile",
        onClick: onClose,
        testId: "user-menu-profile-link",
        to: "/preferences/profile",
      })}
      {renderUserMenuLink({
        icon: <SettingsRoundedIcon fontSize="small" />,
        label: "Application settings",
        onClick: onClose,
        testId: "user-menu-application-settings-link",
        to: "/preferences/application",
      })}
      <Divider />
      <MenuItem data-testid="user-menu-logout-button" onClick={onLogout}>
        <ListItemIcon>
          <LogoutRoundedIcon fontSize="small" />
        </ListItemIcon>
        <ListItemText primary="Logout" />
      </MenuItem>
    </Menu>
  );
};

/**
 * Values required to render one route link in the authenticated user menu.
 */
type UserMenuLinkRenderOptions = {
  icon: ReactElement;
  label: string;
  onClick: () => void;
  testId: string;
  to: string;
};

/**
 * Builds one route link in the authenticated user menu.
 */
const renderUserMenuLink = ({
  icon,
  label,
  onClick,
  testId,
  to,
}: UserMenuLinkRenderOptions): ReactElement => {
  return (
    <MenuItem component={RouterLink} data-testid={testId} onClick={onClick} to={to}>
      <ListItemIcon>{icon}</ListItemIcon>
      <ListItemText primary={label} />
    </MenuItem>
  );
};
