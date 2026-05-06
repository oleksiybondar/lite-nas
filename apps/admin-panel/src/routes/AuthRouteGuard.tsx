import { useAuth } from "@hooks/useAuth";
import Box from "@mui/material/Box";
import CircularProgress from "@mui/material/CircularProgress";
import { LoginPage } from "@pages/LoginPage";
import type { ReactElement } from "react";
import { Outlet } from "react-router-dom";

/**
 * Route guard for admin-panel application content.
 *
 * The guard waits for the SPA-wide auth bootstrap, renders the login page for
 * anonymous sessions, and renders the protected child route when the user is
 * authenticated.
 */
export const AuthRouteGuard = (): ReactElement => {
  const { isAuthInited, isAuthenticated } = useAuth();

  if (!isAuthInited) {
    return <AuthLoadingView />;
  }

  if (!isAuthenticated) {
    return <LoginPage />;
  }

  return <Outlet />;
};

/**
 * Loading state shown while the initial `/api/auth/me` request is pending.
 */
const AuthLoadingView = (): ReactElement => {
  return (
    <Box
      alignItems="center"
      aria-label="Loading authentication state"
      data-testid="auth-loading-view"
      display="flex"
      justifyContent="center"
      minHeight="100vh"
    >
      <CircularProgress data-testid="auth-loading-progress" />
    </Box>
  );
};
