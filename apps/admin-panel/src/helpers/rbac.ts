import {
  ADMIN_ROLE_GROUP,
  OPERATOR_ROLE_GROUP,
  type RbacRole,
  type RbacScope,
  SECURITY_ROLE_GROUP,
} from "@dto/rbac/rbac";

/**
 * Returns whether any acceptable role exists in the supplied role set.
 *
 * This helper is the shared primitive for RBAC role checks. It fails closed
 * when either collection is empty so cleared auth state cannot accidentally
 * satisfy a guard.
 */
export const requireRole = (
  acceptableRoles: readonly RbacRole[],
  actualRoles: readonly RbacRole[],
): boolean => {
  if (acceptableRoles.length === 0 || actualRoles.length === 0) {
    return false;
  }

  return acceptableRoles.some((acceptableRole) => actualRoles.includes(acceptableRole));
};

/**
 * Returns whether any acceptable scope exists in the supplied scope set.
 *
 * This helper mirrors `requireRole` for scope-based checks and also fails
 * closed when either collection is empty.
 */
export const requireScope = (
  acceptableScopes: readonly RbacScope[],
  actualScopes: readonly RbacScope[],
): boolean => {
  if (acceptableScopes.length === 0 || actualScopes.length === 0) {
    return false;
  }

  return acceptableScopes.some((acceptableScope) => actualScopes.includes(acceptableScope));
};

/**
 * Returns whether the supplied roles satisfy administrator access.
 *
 * Administrators are modeled as unix-style superusers and therefore satisfy
 * higher-level role checks that delegate through this helper.
 */
export const requireAdmin = (actualRoles: readonly RbacRole[]): boolean => {
  return requireRole(ADMIN_ROLE_GROUP, actualRoles);
};

/**
 * Returns whether the supplied roles satisfy operator access.
 *
 * Operator access accepts principals with an explicit operator role and also
 * accepts administrator principals through the admin fallback path.
 */
export const requireOperator = (actualRoles: readonly RbacRole[]): boolean => {
  return requireRole(OPERATOR_ROLE_GROUP, actualRoles) || requireAdmin(actualRoles);
};

/**
 * Returns whether the supplied roles satisfy security access.
 *
 * Security access accepts principals with an explicit security role and also
 * accepts administrator principals through the admin fallback path.
 */
export const requireSecurity = (actualRoles: readonly RbacRole[]): boolean => {
  return requireRole(SECURITY_ROLE_GROUP, actualRoles) || requireAdmin(actualRoles);
};
