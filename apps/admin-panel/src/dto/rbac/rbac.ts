/**
 * Supported role names used by admin-panel authorization checks.
 */
export const ADMIN_ROLE_GROUP = ["admin", "sudo"] as const;

/**
 * Operator-capable roles accepted by operator-only feature gates.
 */
export const OPERATOR_ROLE_GROUP = ["lite-nas-operator"] as const;

/**
 * Security-capable roles accepted by security-only feature gates.
 */
export const SECURITY_ROLE_GROUP = ["lite-nas-security"] as const;

/**
 * Role collection accepted by RBAC helper predicates.
 */
export type RbacRole = string;

/**
 * Scope collection accepted by RBAC helper predicates.
 */
export type RbacScope = string;

/**
 * Browser-facing RBAC state and authorization guards derived from auth state.
 */
export type RbacContextValue = {
  /**
   * Current authenticated principal roles, or an empty list when unavailable.
   */
  roles: RbacRole[];
  /**
   * Current authenticated principal scopes, or an empty list when unavailable.
   */
  scopes: RbacScope[];
  /**
   * Returns whether the current principal satisfies administrator access.
   *
   * This guard grants access to principals holding any role from
   * `ADMIN_ROLE_GROUP`. When auth state is anonymous or cleared, it fails
   * closed and returns `false`.
   */
  requireAdmin: () => boolean;
  /**
   * Returns whether the current principal satisfies operator access.
   *
   * This guard grants access to principals holding any role from
   * `OPERATOR_ROLE_GROUP` and also accepts administrator principals through the
   * admin fallback path. When auth state is anonymous or cleared, it fails
   * closed and returns `false`.
   */
  requireOperator: () => boolean;
  /**
   * Returns whether the current principal satisfies security access.
   *
   * This guard grants access to principals holding any role from
   * `SECURITY_ROLE_GROUP` and also accepts administrator principals through the
   * admin fallback path. When auth state is anonymous or cleared, it fails
   * closed and returns `false`.
   */
  requireSecurity: () => boolean;
};
