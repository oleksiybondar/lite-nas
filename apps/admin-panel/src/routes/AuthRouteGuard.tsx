import { useAuth } from "@hooks/useAuth";
import { loginDestination, loginRedirectPath } from "@lite-nas/profile";
import Box from "@mui/material/Box";
import CircularProgress from "@mui/material/CircularProgress";
import { LoginPage } from "@pages/LoginPage";
import type { ReactElement } from "react";
import { Navigate, Outlet, useLocation } from "react-router-dom";

/**
 * Route guard for admin-panel application content.
 *
 * The guard waits for the SPA-wide auth bootstrap, renders the login page for
 * anonymous sessions, and renders the protected child route when the user is
 * authenticated.
 */
export const AuthRouteGuard = (): ReactElement => {
  const { isAuthInited, isAuthenticated } = useAuth();
  const location = useLocation();

  if (!isAuthInited) {
    return <AuthLoadingView />;
  }

  if (!isAuthenticated) {
    if (location.pathname === "/login") return <LoginPage />;
    const from = location.pathname + location.search + location.hash;
    return <Navigate replace to={loginRedirectPath(from)} state={{ from }} />;
  }

  if (location.pathname === "/login") {
    const from = (location.state as { from?: unknown } | null)?.from;
    return <Navigate replace to={loginDestination(location.search, from)} />;
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
