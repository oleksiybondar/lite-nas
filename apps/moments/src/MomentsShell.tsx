import type { ThemeMode } from "@lite-nas/theme";
import Box from "@mui/material/Box";
import Container from "@mui/material/Container";
import type { ReactElement } from "react";
import { useState } from "react";
import { logout, saveSharedMode } from "./api";
import { DemoGallery } from "./DemoGallery";
import { MomentsHeader } from "./MomentsHeader";
import { MomentsLayout } from "./MomentsLayout";
import { Settings } from "./Settings";
import type { Me } from "./schemas";
import { useProfileAvatar } from "./useProfileAvatar";

/** Provides the authenticated header, demo gallery, and shared settings views. */
export const MomentsShell = ({
  me,
  mode,
  onModeChange,
  onSignedOut,
}: {
  me: Me;
  mode: ThemeMode;
  onModeChange: (mode: ThemeMode) => void;
  onSignedOut: () => void;
}): ReactElement => {
  const [view, setView] = useState<"home" | "profile" | "appearance">("home");
  const [message, setMessage] = useState<string | null>(null);
  const avatar = useProfileAvatar();

  const saveMode = async (value: ThemeMode): Promise<void> => {
    try {
      await saveSharedMode(value);
      setMessage(null);
    } catch {
      setMessage("Saved in this browser, but the user file could not be updated.");
    }
  };
  const onLogout = async (): Promise<void> => {
    try {
      await logout();
      onSignedOut();
    } catch {
      setMessage("Sign out failed.");
    }
  };

  return (
    <MomentsLayout
      header={
        <MomentsHeader
          username={me.user.login}
          fullName={me.user.full_name}
          avatarUrl={avatar ?? me.user.avatar_url}
          onHome={() => setView("home")}
          onProfile={() => setView("profile")}
          onApplicationSettings={() => setView("appearance")}
          onLogout={() => void onLogout()}
        />
      }
    >
      <Container maxWidth={false} sx={{ py: 4, width: "100%" }}>
        {view === "home" ? (
          <Box data-testid="moments-home">
            <DemoGallery />
          </Box>
        ) : (
          <Settings
            view={view}
            username={me.user.login}
            avatarUrl={avatar}
            fallbackAvatarUrl={me.user.avatar_url}
            mode={mode}
            message={message}
            onModeChange={onModeChange}
            onSaveMode={(value) => void saveMode(value)}
          />
        )}
      </Container>
    </MomentsLayout>
  );
};
