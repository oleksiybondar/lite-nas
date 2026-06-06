import { useRbac } from "@hooks/useRbac";

type MockRbacAccessInput = {
  requireOperator: boolean;
  requireSecurity: boolean;
};

/**
 * Mocks the RBAC hook with the minimal access contract used by unit tests.
 */
export const mockRbacAccess = ({ requireOperator, requireSecurity }: MockRbacAccessInput): void => {
  vi.mocked(useRbac).mockReturnValue({
    requireAdmin: () => false,
    requireOperator: () => requireOperator,
    requireSecurity: () => requireSecurity,
    roles: [],
    scopes: [],
  });
};
