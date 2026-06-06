import { AuthContext } from "@contexts/auth-context";
import type { AuthContextValue } from "@dto/auth/auth";
import { LoginPage } from "@pages/LoginPage";
import { AuthRouteGuard } from "@routes/AuthRouteGuard";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { TestMemoryRouter } from "@tests/unit/test-utils/router";
import type { ReactElement } from "react";
import { Route, Routes } from "react-router-dom";

describe("AuthRouteGuard", () => {
  test("renders loading state before auth initialization completes", () => {
    renderGuard({ isAuthInited: false });

    expect(screen.getByLabelText("Loading authentication state")).toBeInTheDocument();
  });

  test("renders the login page for anonymous sessions", () => {
    renderGuard({ isAuthenticated: false });

    expect(screen.getByRole("heading", { name: "Sign in" })).toBeInTheDocument();
    expect(screen.queryByText("Protected content")).not.toBeInTheDocument();
  });

  test("renders protected route content for authenticated sessions", () => {
    renderGuard({ isAuthenticated: true });

    expect(screen.getByText("Protected content")).toBeInTheDocument();
  });
});

describe("LoginPage", () => {
  test("submits credentials through the auth context", async () => {
    const login = vi.fn().mockResolvedValue(responseWithStatus(200));

    renderWithAuth(<LoginPage />, { login });
    fireEvent.change(screen.getByLabelText(/Login/i), { target: { value: "admin" } });
    fireEvent.change(screen.getByLabelText(/Password/i), { target: { value: "passw" } });
    fireEvent.click(screen.getByRole("button", { name: "Sign in" }));

    await waitFor(() => {
      expect(login).toHaveBeenCalledWith("admin", "passw");
    });
  });

  test("renders an error when login fails", async () => {
    const login = vi.fn().mockResolvedValue(responseWithStatus(401));

    renderWithAuth(<LoginPage />, { login });
    fireEvent.change(screen.getByLabelText(/Login/i), { target: { value: "admin" } });
    fireEvent.change(screen.getByLabelText(/Password/i), { target: { value: "wrong" } });
    fireEvent.click(screen.getByRole("button", { name: "Sign in" }));

    expect(await screen.findByText("Login failed.")).toBeInTheDocument();
  });
});

/**
 * Renders the route guard with a protected child route.
 */
const renderGuard = (overrides: Partial<AuthContextValue> = {}) => {
  return renderWithAuth(
    <Routes>
      <Route element={<AuthRouteGuard />}>
        <Route element={<span>Protected content</span>} path="/" />
      </Route>
    </Routes>,
    overrides,
  );
};

/**
 * Renders a component under a controlled auth context value.
 */
const renderWithAuth = (component: ReactElement, overrides: Partial<AuthContextValue> = {}) => {
  return render(
    <TestMemoryRouter>
      <AuthContext.Provider value={createAuthContextValue(overrides)}>
        {component}
      </AuthContext.Provider>
    </TestMemoryRouter>,
  );
};

/**
 * Creates a complete auth context value for guard and page tests.
 */
const createAuthContextValue = (overrides: Partial<AuthContextValue> = {}): AuthContextValue => {
  return {
    isAuthInited: true,
    isAuthenticated: false,
    login: vi.fn().mockResolvedValue(responseWithStatus(200)),
    logout: vi.fn().mockResolvedValue(responseWithStatus(204)),
    me: null,
    ...overrides,
  };
};

/**
 * Creates a minimal fetch response with the supplied status.
 */
const responseWithStatus = (status: number): Response => {
  return new Response(null, { status });
};
