import { z } from "zod";

/**
 * Runtime schema for one supported monitoring polling mode.
 */
export const monitoringPollingModeSchema = z.enum(["history", "snapshot"]);

/**
 * Runtime schema for persisted positive-integer monitoring settings.
 */
export const monitoringPollingPositiveIntegerSchema = z.number().int().positive();

/**
 * Runtime schema for one complete persisted monitoring polling settings object.
 */
export const monitoringPollingSettingsSchema = z.object({
  historyIntervalMs: monitoringPollingPositiveIntegerSchema,
  historyResetGapMs: monitoringPollingPositiveIntegerSchema,
  maxRecords: monitoringPollingPositiveIntegerSchema,
  mode: monitoringPollingModeSchema,
  snapshotIntervalMs: monitoringPollingPositiveIntegerSchema,
});

/**
 * Runtime schema for editable string-backed monitoring polling settings form state.
 */
export const monitoringPollingSettingsFormSchema = z.object({
  historyIntervalMs: z.coerce.number().int().positive(),
  historyResetGapMs: z.coerce.number().int().positive(),
  maxRecords: z.coerce.number().int().positive(),
  mode: monitoringPollingModeSchema,
  snapshotIntervalMs: z.coerce.number().int().positive(),
});
