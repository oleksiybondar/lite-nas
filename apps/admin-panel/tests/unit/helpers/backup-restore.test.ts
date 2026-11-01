import { backupPath, changedLines } from "@helpers/backup-restore";
import { describe, expect, it } from "vitest";

describe("backup merge helpers", () => {
  it("marks a changed result line", () => {
    expect([...changedLines("one\ntwo\nthree", "one\nupdated\nthree")]).toEqual([1]);
  });

  it("keeps following lines clear after an insertion", () => {
    expect([...changedLines("one\nthree", "one\ntwo\nthree")]).toEqual([1]);
  });

  it("marks the adjacent result line after a deletion", () => {
    expect([...changedLines("one\ntwo\nthree", "one\nthree")]).toEqual([1]);
  });

  it("encodes inventory-selected route operands", () => {
    expect(backupPath("app name", "v/1")).toBe(
      "/api/backup-restore/applications/app%20name/versions/v%2F1",
    );
  });
});
