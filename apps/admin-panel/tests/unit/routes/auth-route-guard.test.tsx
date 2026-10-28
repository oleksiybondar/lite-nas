import { AuthContext } from "@contexts/auth-context";
import type { AuthContextValue } from "@dto/auth/auth";
import { languageStorageKey } from "@lite-nas/language";
import { LoginPage } from "@pages/LoginPage";
import { AuthRouteGuard } from "@routes/AuthRouteGuard";
import { fireEvent, screen, waitFor } from "@testing-library/react";
import { renderWithLanguage } from "@tests/unit/test-utils/language";
import { TestMemoryRouter } from "@tests/unit/test-utils/router";
import type { ReactElement } from "react";
import { MemoryRouter, Route, Routes, useLocation } from "react-router-dom";

afterEach(() => window.localStorage.removeItem(languageStorageKey));

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

  test("keeps the protected URL in redirect_to when routing to login", async () => {
    renderWithLanguage(
      <MemoryRouter initialEntries={["/system?tab=network"]}>
        <AuthContext.Provider value={createAuthContextValue({ isAuthenticated: false })}>
          <RouteLocation />
          <Routes>
            <Route element={<AuthRouteGuard />}>
              <Route element={<LoginPage />} path="/login" />
              <Route element={<span>System page</span>} path="/system" />
            </Route>
          </Routes>
        </AuthContext.Provider>
      </MemoryRouter>,
    );
    expect(
      await screen.findByText("/login?redirect_to=%2Fsystem%3Ftab%3Dnetwork"),
    ).toBeInTheDocument();
  });

  test("renders protected route content for authenticated sessions", () => {
    renderGuard({ isAuthenticated: true });

    expect(screen.getByText("Protected content")).toBeInTheDocument();
  });

  test("an already restored session returns from login to the protected route", async () => {
    renderWithLanguage(
      <MemoryRouter initialEntries={[{ pathname: "/login", state: { from: "/system" } }]}>
        <AuthContext.Provider value={createAuthContextValue({ isAuthenticated: true })}>
          <Routes>
            <Route element={<AuthRouteGuard />}>
              <Route element={<LoginPage />} path="/login" />
              <Route element={<span>System page</span>} path="/system" />
            </Route>
          </Routes>
        </AuthContext.Provider>
      </MemoryRouter>,
    );
    expect(await screen.findByText("System page")).toBeInTheDocument();
  });
});

describe("LoginPage", () => {
  test("uses the browser language before authentication", () => {
    window.localStorage.setItem(languageStorageKey, "uk");
    renderWithAuth(<LoginPage />);
    expect(screen.getByRole("heading", { name: "Увійти" })).toBeInTheDocument();
  });

  test("submits credentials through the auth context", async () => {
    const login = vi.fn().mockResolvedValue(responseWithStatus(200));

    renderWithAuth(<LoginPage />, { login });
    fireEvent.change(screen.getByLabelText(/Login/i), { target: { value: "admin" } });
    fireEvent.change(screen.getByLabelText(/Password/i), { target: { value: "passw" } });
    fireEvent.click(screen.getByRole("button", { name: "Sign in" }));

    await waitFor(() => {
      expect(login).toHaveBeenCalledWith("admin", "passw", "/");
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

  test("returns to the protected route after successful login", async () => {
    const login = vi.fn().mockResolvedValue(responseWithStatus(200));
    renderWithLanguage(
      <MemoryRouter initialEntries={[{ pathname: "/login", state: { from: "/system" } }]}>
        <AuthContext.Provider value={createAuthContextValue({ login })}>
          <Routes>
            <Route element={<LoginPage />} path="/login" />
            <Route element={<span>System page</span>} path="/system" />
          </Routes>
        </AuthContext.Provider>
      </MemoryRouter>,
    );
    fireEvent.change(screen.getByLabelText(/Login/i), { target: { value: "admin" } });
    fireEvent.change(screen.getByLabelText(/Password/i), { target: { value: "passw" } });
    fireEvent.click(screen.getByRole("button", { name: "Sign in" }));
    expect(await screen.findByText("System page")).toBeInTheDocument();
  });
});

describe("LoginPage return paths", () => {
  test("returns to redirect_to after login without router state", async () => {
    const login = vi.fn().mockResolvedValue(responseWithStatus(200));
    renderWithLanguage(
      <MemoryRouter initialEntries={["/login?redirect_to=%2Fsystem"]}>
        <AuthContext.Provider value={createAuthContextValue({ login })}>
          <Routes>
            <Route element={<LoginPage />} path="/login" />
            <Route element={<span>System page</span>} path="/system" />
          </Routes>
        </AuthContext.Provider>
      </MemoryRouter>,
    );
    fireEvent.change(screen.getByLabelText(/Login/i), { target: { value: "admin" } });
    fireEvent.change(screen.getByLabelText(/Password/i), { target: { value: "passw" } });
    fireEvent.click(screen.getByRole("button", { name: "Sign in" }));
    expect(await screen.findByText("System page")).toBeInTheDocument();
  });

  test("uses the backend-validated destination after successful login", async () => {
    const login = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: { authenticated: true, user: { id: "admin-id" }, redirect_to: "/system" },
        }),
        { status: 200 },
      ),
    );
    renderWithLanguage(
      <MemoryRouter initialEntries={["/login?redirect_to=%2Fpreferences"]}>
        <AuthContext.Provider value={createAuthContextValue({ login })}>
          <Routes>
            <Route element={<LoginPage />} path="/login" />
            <Route element={<span>System page</span>} path="/system" />
          </Routes>
        </AuthContext.Provider>
      </MemoryRouter>,
    );
    fireEvent.change(screen.getByLabelText(/Login/i), { target: { value: "admin" } });
    fireEvent.change(screen.getByLabelText(/Password/i), { target: { value: "passw" } });
    fireEvent.click(screen.getByRole("button", { name: "Sign in" }));
    expect(await screen.findByText("System page")).toBeInTheDocument();
  });
});

/**
 * Renders the route guard with a protected child route.
 */
const renderGuard = (overrides: Partial<AuthContextValue> = {}) => {
  return renderWithAuth(
    <Routes>
      <Route element={<AuthRouteGuard />}>
        <Route element={<LoginPage />} path="/login" />
        <Route element={<span>Protected content</span>} path="/" />
      </Route>
    </Routes>,
    overrides,
  );
};

/** Exposes the in-memory browser route for redirect assertions. */
const RouteLocation = (): ReactElement => {
  const location = useLocation();
  return <span>{location.pathname + location.search}</span>;
};

/**
 * Renders a component under a controlled auth context value.
 */
const renderWithAuth = (component: ReactElement, overrides: Partial<AuthContextValue> = {}) => {
  return renderWithLanguage(
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
