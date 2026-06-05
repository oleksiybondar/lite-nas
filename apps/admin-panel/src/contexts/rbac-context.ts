import type { RbacContextValue } from "@dto/rbac/rbac";
import { createContext } from "react";

/**
 * Context for role- and scope-based authorization derived from auth state.
 *
 * Consumers should use `useRbac` so missing provider wiring fails with a clear
 * local error instead of leaking `undefined` checks into feature components.
 * The context exposes only domain guards and normalized role/scope state, while
 * generic matching helpers remain internal to the RBAC layer.
 */
export const RbacContext = createContext<RbacContextValue | undefined>(undefined);
