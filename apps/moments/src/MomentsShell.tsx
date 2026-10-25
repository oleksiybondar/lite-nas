// biome-ignore-all lint/complexity/noExcessiveLinesPerFunction: shell navigation stays together while the demo views are reviewed.
// biome-ignore-all lint/complexity/noExcessiveCognitiveComplexity: shell view routing stays explicit for the current screens.

import { FileSelectionDialog, FolderSelectionDialog } from "@lite-nas/file-browser";
import type { ThemeMode } from "@lite-nas/theme";
import Box from "@mui/material/Box";
import Container from "@mui/material/Container";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import { useState } from "react";
import { logout, saveSharedMode } from "./api";
import { DemoGallery } from "./DemoGallery";
import { MomentsHeader } from "./MomentsHeader";
import { MomentsLayout } from "./MomentsLayout";
import { Settings } from "./Settings";
import type { Me } from "./schemas";
import { ThumbnailMaker } from "./ThumbnailMaker";
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
  const [view, setView] = useState<"home" | "thumbnail" | "profile" | "appearance">("home");
  const [folderDialogOpen, setFolderDialogOpen] = useState(false);
  const [fileDialogOpen, setFileDialogOpen] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [selectedFolder, setSelectedFolder] = useState<string | null>(null);
  const [selectedFile, setSelectedFile] = useState<string | null>(null);
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
          onThumbnailMaker={() => setView("thumbnail")}
          onSelectFolder={() => setFolderDialogOpen(true)}
          onSelectFile={() => setFileDialogOpen(true)}
          onProfile={() => setView("profile")}
          onApplicationSettings={() => setView("appearance")}
          onLogout={() => void onLogout()}
        />
      }
    >
      <Container
        maxWidth={false}
        sx={{
          py: 2,
          width: "100%",
          height: "100%",
          minHeight: 0,
          boxSizing: "border-box",
          display: "flex",
          flexDirection: "column",
        }}
      >
        {view === "home" && (
          <Box data-testid="moments-home">
            {(selectedFolder || selectedFile) && (
              <Stack spacing={0.5} sx={{ mb: 2 }}>
                {selectedFolder && (
                  <Typography variant="body2">Selected folder: {selectedFolder}</Typography>
                )}
                {selectedFile && (
                  <Typography variant="body2">Selected file: {selectedFile}</Typography>
                )}
              </Stack>
            )}
            <DemoGallery />
          </Box>
        )}
        {view === "thumbnail" && <ThumbnailMaker />}
        {(view === "profile" || view === "appearance") && (
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
      <FolderSelectionDialog
        open={folderDialogOpen}
        onClose={() => setFolderDialogOpen(false)}
        initialDirectory="/home/oleksiybondar"
        onSelect={setSelectedFolder}
      />
      <FileSelectionDialog
        open={fileDialogOpen}
        onClose={() => setFileDialogOpen(false)}
        initialDirectory="/home/oleksiybondar"
        onSelect={setSelectedFile}
      />
    </MomentsLayout>
  );
};
