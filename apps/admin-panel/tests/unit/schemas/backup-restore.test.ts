import { ufwStateSchema } from "@schemas/backup-restore";
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
