import type { ApiContextValue } from "@dto/api/api";
import type { MonitoringPollingSettings } from "@dto/monitoring/monitoring-polling-settings";
import {
  loadMonitoringPollingSettings,
  saveMonitoringPollingSettings,
} from "@helpers/monitoring-polling-settings-storage";
import { loadLocalLanguage } from "@lite-nas/language";
import {
  loadLocalSharedUi,
  loadSharedUiPreference,
  readProfileDocument,
  saveLocalSharedUi,
  sharedUiSchema,
  writeProfileDocument,
} from "@lite-nas/profile";
import { adminSettingsSchema } from "@schemas/profile/admin-settings";
import { loadThemeSettings, saveThemeSettings } from "@theme/manager/storage";
import type { ThemeSettings } from "@theme/types";
import type { z } from "zod";

/** Admin panel polling domains persisted in the app-specific profile file. */
export const monitoringScopes = [
  "system-metrics",
  "network-metrics",
  "disk-metrics",
  "zfs-metrics",
  "process-metrics",
  "service-metrics",
] as const;

type AdminSettings = z.infer<typeof adminSettingsSchema>;

/** Replicates authenticated app settings for use when the profile file is unavailable. */
const replicateAdminSettings = (admin: AdminSettings): void => {
  saveThemeSettings(admin.theme);
  for (const scope of monitoringScopes) {
    const settings = admin.monitoring[scope];
    if (settings) saveMonitoringPollingSettings(scope, settings);
  }
};

/** Loads authenticated files first and replicates valid settings to local storage. */
export const loadProfileSettings = async (
  api: Pick<ApiContextValue, "get">,
): Promise<AdminSettings | null> => {
  const request = (url: string): Promise<Response> => api.get(url).execute();
  await loadSharedUiPreference(request);
  const admin = await readProfileDocument(request, "admin-panel", adminSettingsSchema).catch(
    () => null,
  );
  if (admin) replicateAdminSettings(admin);
  return admin;
};

/** Builds the app-specific document from validated local storage values. */
export const collectAdminSettings = (): AdminSettings => {
  const monitoring: Record<string, MonitoringPollingSettings> = {};
  for (const scope of monitoringScopes) {
    monitoring[scope] = loadMonitoringPollingSettings(scope);
  }
  return adminSettingsSchema.parse({ theme: loadThemeSettings(), monitoring });
};

let saveQueue = Promise.resolve();

/** Serializes app-setting saves so rapid changes cannot replace newer data. */
export const saveAdminSettings = (api: Pick<ApiContextValue, "put">): Promise<void> => {
  const pending = saveQueue.then(async () => {
    await writeProfileDocument(
      (url, body) => api.put(url, body).execute(),
      "admin-panel",
      adminSettingsSchema,
      collectAdminSettings(),
    );
  });
  saveQueue = pending.catch(() => undefined);
  return pending;
};

/** Saves the shared mode locally first, then persists it to the authenticated profile. */
export const saveSharedMode = async (
  api: Pick<ApiContextValue, "put">,
  mode: ThemeSettings["mode"],
): Promise<void> => {
  const value = sharedUiSchema.parse({
    preferredMode: mode,
    preferredLanguage: loadLocalLanguage(),
  });
  saveLocalSharedUi(value);
  await writeProfileDocument(
    (url, body) => api.put(url, body).execute(),
    "ui",
    sharedUiSchema,
    value,
  );
};

/** Uses the browser preference when an authenticated file is absent or unavailable. */
export const localPreferredMode = (): ThemeSettings["mode"] | null => {
  return loadLocalSharedUi()?.preferredMode ?? null;
};
