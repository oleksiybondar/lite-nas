import { monitoringPollingSettingsSchema } from "@schemas/monitoring/monitoring-polling-settings";
import { z } from "zod";

/** Converts a saved four-column widget to the current six-column layout. */
const migrateWidget = (widget: {
  id: string;
  kind: "system" | "zfs" | "cpu" | "cpu-cores" | "ram" | "zfs-io" | "zfs-errors";
  width: 3 | 6 | 12;
  height: 1 | 2;
}) => {
  let kind: "cpu" | "zfs-io" | "cpu-cores" | "ram" | "zfs-errors";
  if (widget.kind === "system") kind = "cpu";
  else if (widget.kind === "zfs") kind = "zfs-io";
  else kind = widget.kind;
  let width = 6;
  if (widget.width === 3) width = 2;
  if (widget.width === 6) width = 3;
  return { id: widget.id, kind, width, height: widget.height };
};

/** Six-column dashboard document, including migration from saved four-column layouts. */
export const dashboardSchema = z
  .union([
    z.object({
      version: z.literal(2),
      widgets: z
        .array(
          z.object({
            id: z.string(),
            kind: z.enum([
              "cpu",
              "cpu-cores",
              "ram",
              "zfs-io",
              "zfs-errors",
              "network",
              "disk",
              "processes",
              "services",
            ]),
            width: z.number().int().min(1).max(6),
            height: z.union([z.literal(1), z.literal(2)]).default(1),
          }),
        )
        .max(16),
    }),
    z.object({
      version: z.literal(1),
      widgets: z
        .array(
          z.object({
            id: z.string(),
            kind: z.enum(["system", "zfs", "cpu", "cpu-cores", "ram", "zfs-io", "zfs-errors"]),
            width: z.union([z.literal(3), z.literal(6), z.literal(12)]),
            height: z.union([z.literal(1), z.literal(2)]).default(1),
          }),
        )
        .max(16),
    }),
  ])
  .transform((document) =>
    document.version === 2
      ? document
      : {
          version: 2 as const,
          widgets: document.widgets.map(migrateWidget),
        },
  );

/** Runtime contract for the admin panel's own `.lite-nas/admin-panel.json`. */
export const adminSettingsSchema = z.object({
  dashboard: dashboardSchema.optional(),
  theme: z.object({
    mode: z.enum(["light", "dark"]),
    source: z.enum(["default", "os", "user"]),
    templateName: z.literal("default"),
  }),
  monitoring: z.record(z.string(), monitoringPollingSettingsSchema),
});
