import {
  backupTaskStartResponseSchema,
  backupTaskStatusResponseSchema,
  ufwStateSchema,
  versionListResponseSchema,
} from "@schemas/backup-restore";
import { describe, expect, it } from "vitest";

const state = {
  schemaVersion: 1,
  enabled: true,
  logging: "low",
  defaults: { incoming: "deny", outgoing: "allow", routed: "deny" },
  rules: [{ kind: "rule", action: "allow", arguments: ["443/tcp"] }],
};

describe("UFW merge JSON", () => {
  it("accepts a complete CLI rule document", () => {
    expect(ufwStateSchema.safeParse(state).success).toBe(true);
  });

  it("rejects option-shaped rule operands", () => {
    expect(
      ufwStateSchema.safeParse({ ...state, rules: [{ ...state.rules[0], arguments: ["--force"] }] })
        .success,
    ).toBe(false);
  });

  it("rejects missing enabled state", () => {
    const missingEnabled: Record<string, unknown> = { ...state };
    delete missingEnabled.enabled;
    expect(ufwStateSchema.safeParse(missingEnabled).success).toBe(false);
  });
});

describe("background backup task responses", () => {
  const taskID = "0123456789abcdef0123456789abcdef";
  const timestamp = "2026-11-10T12:00:00Z";

  it("accepts a queued task identifier", () => {
    expect(
      backupTaskStartResponseSchema.safeParse({
        success: true,
        timestamp,
        data: { task_id: taskID, state: "queued" },
      }).success,
    ).toBe(true);
  });

  it("rejects malformed task identifiers", () => {
    expect(
      backupTaskStartResponseSchema.safeParse({
        success: true,
        timestamp,
        data: { task_id: "../../etc/passwd", state: "queued" },
      }).success,
    ).toBe(false);
  });
});

describe("background backup task progress", () => {
  const taskID = "0123456789abcdef0123456789abcdef";
  const timestamp = "2026-11-10T12:00:00Z";

  it("accepts a valid progress snapshot", () => {
    expect(
      backupTaskStatusResponseSchema.safeParse({
        success: true,
        timestamp,
        data: {
          task_id: taskID,
          application: "all",
          state: "running",
          started_at: timestamp,
          updated_at: timestamp,
          progress: { percent: 50, message: "Processed example" },
          results: [{ application: "example", status: "completed", version: "v1" }],
        },
      }).success,
    ).toBe(true);
  });

  it("rejects progress above 100 percent", () => {
    expect(
      backupTaskStatusResponseSchema.safeParse({
        success: true,
        timestamp,
        data: {
          task_id: taskID,
          application: "all",
          state: "running",
          started_at: timestamp,
          updated_at: timestamp,
          progress: { percent: 101 },
        },
      }).success,
    ).toBe(false);
  });
});

describe("backup snapshot inventory", () => {
  it("accepts snapshot names alongside version identifiers", () => {
    expect(
      versionListResponseSchema.safeParse({
        success: true,
        timestamp: "2026-11-10T12:00:00Z",
        data: {
          versions: ["v1"],
          snapshots: [{ version: "v1", name: "nightly" }],
        },
      }).success,
    ).toBe(true);
  });
});
