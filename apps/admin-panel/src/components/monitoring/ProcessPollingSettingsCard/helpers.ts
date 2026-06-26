import { z } from "zod";
import type { ProcessPollingSettingsDraft } from "./types";

export const processPollingSettingsFormSchema = z.object({
  snapshotIntervalMs: z.coerce.number().int().positive(),
});

/**
 * Builds the editable draft state from one process polling interval value.
 */
export const createProcessPollingDraft = (
  snapshotIntervalMs: number,
): ProcessPollingSettingsDraft => {
  return { snapshotIntervalMs: String(snapshotIntervalMs) };
};
