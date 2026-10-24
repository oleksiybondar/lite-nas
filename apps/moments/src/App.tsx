import { loadLocalSharedUi, loadSharedUiPreference } from "@lite-nas/profile";
import { createAppTheme, type ThemeMode } from "@lite-nas/theme";
import Box from "@mui/material/Box";
import Container from "@mui/material/Container";
import CssBaseline from "@mui/material/CssBaseline";
import { ThemeProvider } from "@mui/material/styles";
import type { ReactElement } from "react";
import { useEffect, useMemo, useState } from "react";
import { loadMe, request } from "./api";
import { Login } from "./Login";
import { MomentsLayout } from "./MomentsLayout";
import { MomentsPublicHeader } from "./MomentsPublicHeader";
import { MomentsShell } from "./MomentsShell";
import type { Me } from "./schemas";

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
    if (!me) return;
    void loadSharedUiPreference(request).then((value) => {
      if (value) setMode(value.preferredMode);
    });
  }, [me]);

  let content: ReactElement;
  if (me === undefined) {
    content = (
      <MomentsLayout header={<MomentsPublicHeader />}>
        <Container sx={{ py: 8, my: "auto" }}>Loading Moments…</Container>
      </MomentsLayout>
    );
  } else if (me === null) {
    content = <Login onSuccess={setMe} />;
  } else {
    content = (
      <MomentsShell me={me} mode={mode} onModeChange={setMode} onSignedOut={() => setMe(null)} />
    );
  }

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Box sx={{ bgcolor: "background.default", color: "text.primary", minHeight: "100vh" }}>
        {content}
      </Box>
    </ThemeProvider>
  );
};
