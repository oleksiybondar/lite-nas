import { useAuth } from "@hooks/useAuth";
import { useProfileAvatar } from "@hooks/useProfileAvatar";
import { UserMenu } from "@lite-nas/profile";
import Alert from "@mui/material/Alert";
import Snackbar from "@mui/material/Snackbar";
import type { ReactElement } from "react";
import { useState } from "react";
import { useNavigate } from "react-router-dom";

/** Connects the shared user menu to admin routes and session state. */
export const AppUserMenu = (): ReactElement | null => {
  const { isAuthenticated, logout, me } = useAuth();
  const avatar = useProfileAvatar();
  const navigate = useNavigate();
  const [logoutError, setLogoutError] = useState(false);

  if (!isAuthenticated || me === null) return null;

  return (
    <>
      <UserMenu
        login={me.user.login}
        fullName={me.user.full_name}
        avatarUrl={avatar ?? me.user.avatar_url}
        onProfile={() => navigate("/preferences/profile")}
        onApplicationSettings={() => navigate("/preferences/application")}
        onLogout={() => {
          void logout()
            .then((response) => {
              if (!response.ok) setLogoutError(true);
            })
            .catch(() => setLogoutError(true));
        }}
      />
      <Snackbar open={logoutError} onClose={() => setLogoutError(false)} autoHideDuration={6000}>
        <Alert severity="error" onClose={() => setLogoutError(false)}>
          Sign out failed.
        </Alert>
      </Snackbar>
    </>
  );
};
