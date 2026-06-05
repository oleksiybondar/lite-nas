import { RbacContext } from "@contexts/rbac-context";
import type { AuthMeDTO } from "@dto/auth/auth";
import type { RbacContextValue, RbacRole, RbacScope } from "@dto/rbac/rbac";
import {
  requireAdmin as requireAdminByRole,
  requireOperator as requireOperatorByRole,
  requireSecurity as requireSecurityByRole,
} from "@helpers/rbac";
import { useAuth } from "@hooks/useAuth";
import type { PropsWithChildren, ReactElement } from "react";

/**
 * Provides RBAC state and authorization guards derived from the current auth
 * session.
 *
 * This provider intentionally owns no independent role or scope state. It
 * derives normalized values from `AuthProvider` so authorization decisions
 * clear immediately when the authenticated principal is removed.
 */
export const RbacProvider = ({ children }: PropsWithChildren): ReactElement => {
  const value = useRbacProviderValue();

  return <RbacContext.Provider value={value}>{children}</RbacContext.Provider>;
};

/**
 * Builds the value exposed by `RbacContext`.
 *
 * The returned guard functions close over the current normalized role and
 * scope snapshots so feature components can perform stable, domain-level
 * checks without duplicating auth-shape knowledge.
 */
const useRbacProviderValue = (): RbacContextValue => {
  const { isAuthenticated, me } = useAuth();
  const roles = getRoles(isAuthenticated, me);
  const scopes = getScopes(isAuthenticated, me);

  return {
    roles,
    scopes,
    requireAdmin: () => requireAdminByRole(roles),
    requireOperator: () => requireOperatorByRole(roles),
    requireSecurity: () => requireSecurityByRole(roles),
  };
};

/**
 * Returns normalized roles for the current auth state.
 *
 * Anonymous, incomplete, or cleared auth state maps to an empty role list so
 * all downstream guards fail closed.
 */
const getRoles = (isAuthenticated: boolean, me: AuthMeDTO | null): RbacRole[] => {
  if (!isAuthenticated || me === null || !me.authenticated) {
    return [];
  }

  return me.roles;
};

/**
 * Returns normalized scopes for the current auth state.
 *
 * Anonymous, incomplete, or cleared auth state maps to an empty scope list so
 * all downstream checks fail closed.
 */
const getScopes = (isAuthenticated: boolean, me: AuthMeDTO | null): RbacScope[] => {
  if (!isAuthenticated || me === null || !me.authenticated) {
    return [];
  }

  return me.scopes;
};
