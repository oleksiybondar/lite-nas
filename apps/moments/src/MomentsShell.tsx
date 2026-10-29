// biome-ignore-all lint/complexity/noExcessiveLinesPerFunction: shell navigation stays together for gallery and settings views.
// biome-ignore-all lint/complexity/noExcessiveCognitiveComplexity: shell view routing stays explicit for the current screens.

import { useLanguage } from "@lite-nas/language";
import type { ThemeMode } from "@lite-nas/theme";
import Box from "@mui/material/Box";
import Container from "@mui/material/Container";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import { useEffect, useRef, useState } from "react";
import { Navigate, Route, Routes, useLocation, useNavigate } from "react-router-dom";
import {
  loadGalleryConcurrency,
  loadMomentsConfig,
  logout,
  saveGalleryConcurrency,
  saveSharedMode,
} from "./api";
import { FriendsSettings } from "./FriendsSettings";
import { GalleryFeed, type GalleryFeedScope } from "./GalleryFeed";
import { GalleryRoute } from "./GalleryRoute";
import { MomentsHeader } from "./MomentsHeader";
import { MomentsLayout } from "./MomentsLayout";
import { Settings, type SettingsView } from "./Settings";
import type { Me } from "./schemas";
import { imageConcurrencySchema } from "./schemas";
import { useProfileAvatar } from "./useProfileAvatar";

/** Feed routes retain the same scope names used by the gallery query. */
const feedRoutes: Record<string, GalleryFeedScope> = {
  "/": "home",
  "/mine": "mine",
  "/shared": "shared",
};

/** Settings URLs select the matching sidebar tab. */
const settingsRoutes: Record<string, SettingsView> = {
  "/settings/profile": "profile",
  "/settings/appearance": "appearance",
  "/settings/gallery": "gallery",
};

/** Provides the authenticated header, live gallery views, and shared settings. */
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
  const { getLabel } = useLanguage();
  const navigate = useNavigate();
  const location = useLocation();
  const scope = feedRoutes[location.pathname] ?? "home";
  const settingsView = settingsRoutes[location.pathname] ?? "appearance";
  const [workingDir, setWorkingDir] = useState<string | null>(null);
  const [configError, setConfigError] = useState(false);
  const [maxConcurrent, setMaxConcurrent] = useState(() => {
    const stored = Number(window.localStorage.getItem("moments-image-concurrency"));
    const parsed = imageConcurrencySchema.safeParse(stored);
    return parsed.success ? parsed.data : 10;
  });
  const [message, setMessage] = useState<string | null>(null);
  const [galleryMessage, setGalleryMessage] = useState<string | null>(null);
  const galleryChanged = useRef(false);
  const avatar = useProfileAvatar();

  useEffect(() => {
    let active = true;
    void loadMomentsConfig()
      .then((path) => {
        if (active) setWorkingDir(path);
      })
      .catch(() => {
        if (active) setConfigError(true);
      });
    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    let active = true;
    void loadGalleryConcurrency().then((value) => {
      if (active && !galleryChanged.current) setMaxConcurrent(value);
    });
    return () => {
      active = false;
    };
  }, []);

  const showFeed = (scope: GalleryFeedScope): void => {
    navigate(scope === "home" ? "/" : `/${scope}`);
  };

  const saveMode = async (value: ThemeMode): Promise<void> => {
    try {
      await saveSharedMode(value);
      setMessage(null);
    } catch {
      setMessage(getLabel("language.saveFailed"));
    }
  };
  const onLogout = async (): Promise<void> => {
    try {
      await logout();
      onSignedOut();
    } catch {
      setMessage(getLabel("moments.settings.logoutFailed"));
    }
  };

  return (
    <MomentsLayout
      header={
        <MomentsHeader
          username={me.user.login}
          fullName={me.user.full_name}
          avatarUrl={avatar ?? me.user.avatar_url}
          onMine={() => showFeed("mine")}
          onShared={() => showFeed("shared")}
          onFriends={() => navigate("/friends")}
          onProfile={() => navigate("/settings/profile")}
          onApplicationSettings={() => navigate("/settings/appearance")}
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
        {configError && (
          <Typography color="error">{getLabel("moments.settings.configFailed")}</Typography>
        )}
        <Routes>
          {["/", "/mine", "/shared"].map((path) => (
            <Route
              key={path}
              path={path}
              element={
                <Box data-testid="moments-home">
                  {workingDir ? (
                    <GalleryFeed
                      key={scope}
                      scope={scope}
                      userId={me.user.id}
                      workingDir={workingDir}
                      onOpen={(gallery) => navigate(`/galleries/${gallery.id}`)}
                    />
                  ) : (
                    !configError && <Typography>{getLabel("moments.loading")}</Typography>
                  )}
                </Box>
              }
            />
          ))}
          <Route
            path="/galleries/:id"
            element={
              <GalleryRoute
                userId={me.user.id}
                maxConcurrent={maxConcurrent}
                workingDir={workingDir ?? "/home"}
              />
            }
          />
          <Route path="/friends" element={<FriendsSettings />} />
          {["profile", "appearance", "gallery"].map((tab) => (
            <Route
              key={tab}
              path={`/settings/${tab}`}
              element={
                <Settings
                  view={settingsView}
                  onViewChange={(value) => navigate(`/settings/${value}`)}
                  username={me.user.login}
                  avatarUrl={avatar}
                  fallbackAvatarUrl={me.user.avatar_url}
                  mode={mode}
                  message={message}
                  onModeChange={onModeChange}
                  onSaveMode={(value) => void saveMode(value)}
                  maxConcurrent={maxConcurrent}
                  galleryMessage={galleryMessage}
                  onMaxConcurrentChange={(value) => {
                    galleryChanged.current = true;
                    setMaxConcurrent(value);
                    void saveGalleryConcurrency(value)
                      .then(() => setGalleryMessage(null))
                      .catch(() => setGalleryMessage(getLabel("language.saveFailed")));
                  }}
                />
              }
            />
          ))}
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Container>
    </MomentsLayout>
  );
};
