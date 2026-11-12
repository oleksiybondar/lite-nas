import {
  journalQuerySchema,
  journalResponseSchema,
  journalVacuumSchema,
} from "@schemas/management/journal";
import { describe, expect, it } from "vitest";

describe("journal browser schemas", () => {
  it("rejects option-shaped unit filters", () => {
    const result = journalQuerySchema.safeParse({
      since: "",
      until: "",
      message: "",
      unit: "--help",
      priority: "",
      cursor: "",
      limit: 50,
    });
    expect(result.success).toBe(false);
  });

  it("accepts a bounded structured chunk", () => {
    const result = journalResponseSchema.safeParse({
      success: true,
      data: {
        entries: [{ cursor: "s=one", timestamp: "123", message: "ready" }],
        next_cursor: "s=one",
        has_more: true,
      },
    });
    expect(result.success).toBe(true);
  });

  it("rejects zero-day vacuum retention", () => {
    expect(journalVacuumSchema.safeParse({ days: 0 }).success).toBe(false);
  });
});
