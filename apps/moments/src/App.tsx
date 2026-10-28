import { LanguageProvider, type LanguageTransport, useLanguage } from "@lite-nas/language";
import {
  loadLocalSharedUi,
  loadSharedUiPreference,
  loginDestination,
  loginRedirectPath,
} from "@lite-nas/profile";
import { createAppTheme, type ThemeMode } from "@lite-nas/theme";
import Box from "@mui/material/Box";
import Container from "@mui/material/Container";
import CssBaseline from "@mui/material/CssBaseline";
import { ThemeProvider } from "@mui/material/styles";
import type { ReactElement } from "react";
import { useEffect, useMemo, useState } from "react";
import { BrowserRouter, Navigate, useLocation } from "react-router-dom";
import { loadMe, request, sessionExpiredEvent } from "./api";
import { Login } from "./Login";
import { MomentsLayout } from "./MomentsLayout";
import { MomentsPublicHeader } from "./MomentsPublicHeader";
import { MomentsShell } from "./MomentsShell";
import { StreamProvider } from "./StreamContext";
import type { Me } from "./schemas";

/** Uses the existing cookie-authenticated Moments request pipeline for profile settings. */
const languageTransport: LanguageTransport = {
  read: request,
  write: (path, body) =>
    request(path, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    }),
};

/** Resolves the browser session before showing the Moments shell. */
export const App = (): ReactElement => {
  const [me, setMe] = useState<Me | null | undefined>(undefined);
  const [mode, setMode] = useState<ThemeMode>(() => loadLocalSharedUi()?.preferredMode ?? "dark");
  const theme = useMemo(() => createAppTheme("default", mode), [mode]);

  useEffect(() => {
    void loadMe()
      .then(setMe)
      .catch(() => setMe(null));
  }, []);

  useEffect(() => {
    const expire = (): void => setMe(null);
    window.addEventListener(sessionExpiredEvent, expire);
    return () => window.removeEventListener(sessionExpiredEvent, expire);
  }, []);

  useEffect(() => {
    if (!me) return;
    void loadSharedUiPreference(request).then((value) => {
      if (value) setMode(value.preferredMode);
    });
  }, [me]);

  return (
    <LanguageProvider userId={me?.user.id} transport={languageTransport}>
      <StreamProvider>
        <ThemeProvider theme={theme}>
          <CssBaseline />
          <Box sx={{ bgcolor: "background.default", color: "text.primary", minHeight: "100vh" }}>
            <BrowserRouter
              basename={
                window.location.pathname.startsWith("/apps/moments/") ? "/apps/moments" : undefined
              }
            >
              <SessionContent
                me={me}
                mode={mode}
                onModeChange={setMode}
                onSignedOut={() => setMe(null)}
                onSignedIn={setMe}
              />
            </BrowserRouter>
          </Box>
        </ThemeProvider>
      </StreamProvider>
    </LanguageProvider>
  );
};

/** Keeps the requested route in the login URL until the browser session is restored. */
const SessionContent = ({
  me,
  mode,
  onModeChange,
  onSignedOut,
  onSignedIn,
}: {
  me: Me | null | undefined;
  mode: ThemeMode;
  onModeChange: (mode: ThemeMode) => void;
  onSignedOut: () => void;
  onSignedIn: (me: Me) => void;
}): ReactElement => {
  const { getLabel } = useLanguage();
  const location = useLocation();
  if (me === undefined)
    return (
      <MomentsLayout header={<MomentsPublicHeader />}>
        <Container sx={{ py: 8, my: "auto" }}>{getLabel("moments.loading")}</Container>
      </MomentsLayout>
    );
  if (me === null) {
    if (location.pathname !== "/login")
      return (
        <Navigate
          replace
          to={loginRedirectPath(location.pathname + location.search + location.hash)}
        />
      );
    return <Login onSuccess={onSignedIn} />;
  }
  if (location.pathname === "/login")
    return <Navigate replace to={loginDestination(location.search)} />;
  return <MomentsShell me={me} mode={mode} onModeChange={onModeChange} onSignedOut={onSignedOut} />;
};
