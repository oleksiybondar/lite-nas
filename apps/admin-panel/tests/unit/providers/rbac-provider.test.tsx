import { AuthContext } from "@contexts/auth-context";
import type { AuthContextValue, AuthMeDTO } from "@dto/auth/auth";
import { useRbac } from "@hooks/useRbac";
import { RbacProvider } from "@providers/RbacProvider";
import { render, screen } from "@testing-library/react";
import type { ReactElement } from "react";

describe("RbacProvider", () => {
  test("derives roles, scopes, and security access from the authenticated principal", () => {
    renderWithAuth(<RbacConsumer />, {
      isAuthenticated: true,
      me: createMe({ roles: ["lite-nas-security"], scopes: ["alerts:read"] }),
    });

    expect(screen.getByTestId("rbac-state")).toHaveTextContent(
      "roles:lite-nas-security|scopes:alerts:read|admin:no|operator:no|security:yes",
    );
  });

  test("grants operator and security access through administrator fallback", () => {
    renderWithAuth(<RbacConsumer />, {
      isAuthenticated: true,
      me: createMe({ roles: ["admin"], scopes: ["admin"] }),
    });

    expect(screen.getByTestId("rbac-state")).toHaveTextContent(
      "roles:admin|scopes:admin|admin:yes|operator:yes|security:yes",
    );
  });

  test("fails all guards immediately when auth state is cleared", () => {
    renderWithAuth(<RbacConsumer />, {
      isAuthenticated: false,
      me: null,
    });

    expect(screen.getByTestId("rbac-state")).toHaveTextContent(
      "roles:none|scopes:none|admin:no|operator:no|security:no",
    );
  });
});

/**
 * Consumer used to exercise RbacProvider through the public hook contract.
 */
const RbacConsumer = (): ReactElement => {
  const { requireAdmin, requireOperator, requireSecurity, roles, scopes } = useRbac();

  return (
    <span data-testid="rbac-state">
      roles:{roles.join(",") || "none"}|scopes:{scopes.join(",") || "none"}|admin:
      {requireAdmin() ? "yes" : "no"}|operator:{requireOperator() ? "yes" : "no"}|security:
      {requireSecurity() ? "yes" : "no"}
    </span>
  );
};

/**
 * Renders a component under controlled auth and RBAC providers.
 */
const renderWithAuth = (component: ReactElement, overrides: Partial<AuthContextValue> = {}) => {
  return render(
    <AuthContext.Provider value={createAuthContextValue(overrides)}>
      <RbacProvider>{component}</RbacProvider>
    </AuthContext.Provider>,
  );
};

/**
 * Creates a complete auth context value for RBAC-provider tests.
 */
const createAuthContextValue = (overrides: Partial<AuthContextValue> = {}): AuthContextValue => {
  return {
    isAuthInited: true,
    isAuthenticated: false,
    login: vi.fn().mockResolvedValue(new Response(null, { status: 200 })),
    logout: vi.fn().mockResolvedValue(new Response(null, { status: 204 })),
    me: null,
    ...overrides,
  };
};

/**
 * Creates a current-user payload fixture for RBAC tests.
 */
const createMe = (overrides: Partial<AuthMeDTO> = {}): AuthMeDTO => {
  return {
    authenticated: true,
    auth_type: "password",
    roles: [],
    scopes: [],
    user: {
      id: "user-id",
      login: "user",
    },
    ...overrides,
  };
};
