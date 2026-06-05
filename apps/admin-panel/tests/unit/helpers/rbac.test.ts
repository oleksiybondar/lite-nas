import {
  requireAdmin,
  requireOperator,
  requireRole,
  requireScope,
  requireSecurity,
} from "@helpers/rbac";

describe("RBAC helper primitives", () => {
  test("requireRole returns false when actual roles are empty", () => {
    expect(requireRole(["admin"], [])).toBe(false);
  });

  test("requireRole returns true when any acceptable role matches", () => {
    expect(requireRole(["admin", "sudo"], ["user", "sudo"])).toBe(true);
  });

  test("requireScope returns false when actual scopes are empty", () => {
    expect(requireScope(["metrics:read"], [])).toBe(false);
  });

  test("requireScope returns true when any acceptable scope matches", () => {
    expect(requireScope(["metrics:read", "alerts:read"], ["alerts:read"])).toBe(true);
  });
});

describe("RBAC domain guards", () => {
  test("requireAdmin accepts sudo as an administrator role", () => {
    expect(requireAdmin(["sudo"])).toBe(true);
  });

  test("requireOperator falls back to administrator roles", () => {
    expect(requireOperator(["admin"])).toBe(true);
  });

  test("requireSecurity accepts an explicit security role", () => {
    expect(requireSecurity(["lite-nas-security"])).toBe(true);
  });

  test("requireSecurity fails closed for unrelated roles", () => {
    expect(requireSecurity(["viewer"])).toBe(false);
  });
});
