import { monitoringPollingSettingsSchema } from "@schemas/monitoring/monitoring-polling-settings";
import { z } from "zod";

/** Runtime contract for the admin panel's own `.lite-nas/admin-panel.json`. */
export const adminSettingsSchema = z.object({
  theme: z.object({
    mode: z.enum(["light", "dark"]),
    source: z.enum(["default", "os", "user"]),
    templateName: z.literal("default"),
  }),
  monitoring: z.record(z.string(), monitoringPollingSettingsSchema),
});
