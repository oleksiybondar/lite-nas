import { ApiContext } from "@contexts/api-context";
import type { ApiContextValue } from "@dto/api/api";
import type { AuthMeDTO } from "@dto/auth/auth";
import type { ApiRequestBuilder } from "@helpers/api-request-builder";
import { useAuth } from "@hooks/useAuth";
import { AuthProvider } from "@providers/AuthProvider";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import type { PropsWithChildren, ReactElement } from "react";

describe("AuthProvider initialization", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  test("loads current user state once during initialization", async () => {
    renderWithApi(<AuthConsumer />, responseWithJson(200, { data: authenticatedMe }));

    await waitFor(() => {
      expect(screen.getByTestId("auth-state")).toHaveTextContent("ready:yes:admin");
    });
  });

  test("resets auth state when current user loading fails", async () => {
    renderWithApi(<AuthConsumer />, responseWithStatus(500));

    await waitFor(() => {
      expect(screen.getByTestId("auth-state")).toHaveTextContent("ready:no:anonymous");
    });
  });
});

describe("AuthProvider commands", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  test("logs in with raw fetch and reloads current user state", async () => {
    const fetchMock = vi.fn().mockResolvedValueOnce(
      responseWithJson(200, {
        data: {
          authenticated: true,
          user: { id: "admin-id" },
        },
      }),
    );
    vi.stubGlobal("fetch", fetchMock);

    renderWithApi(
      <AuthConsumer />,
      responseWithStatus(401),
      responseWithJson(200, { data: authenticatedMe }),
    );
    await screen.findByText("ready:no:anonymous");
    fireEvent.click(screen.getByRole("button", { name: "Login" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith("/api/auth/login", {
        body: JSON.stringify({ login: "admin", password: "passw" }),
        credentials: "include",
        headers: expect.any(Headers),
        method: "POST",
      });
    });
    await waitFor(() => {
      expect(screen.getByTestId("auth-state")).toHaveTextContent("ready:yes:admin");
    });
  });

  test("logs out with raw fetch and clears auth state", async () => {
    const fetchMock = vi.fn().mockResolvedValueOnce(responseWithStatus(204));
    vi.stubGlobal("fetch", fetchMock);

    renderWithApi(<AuthConsumer />, responseWithJson(200, { data: authenticatedMe }));
    await screen.findByText("ready:yes:admin");
    fireEvent.click(screen.getByRole("button", { name: "Logout" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith("/api/auth/logout", {
        body: JSON.stringify({}),
        credentials: "include",
        headers: expect.any(Headers),
        method: "POST",
      });
    });
    await waitFor(() => {
      expect(screen.getByTestId("auth-state")).toHaveTextContent("ready:no:anonymous");
    });
  });
});

/**
 * Consumer used to exercise AuthProvider through the public hook contract.
 */
const AuthConsumer = (): ReactElement => {
  const { isAuthInited, isAuthenticated, login, logout, me } = useAuth();

  return (
    <>
      <span data-testid="auth-state">
        {isAuthInited ? "ready" : "loading"}:{isAuthenticated ? "yes" : "no"}:
        {me?.user.login ?? "anonymous"}
      </span>
      <button
        onClick={() => {
          void login("admin", "passw");
        }}
        type="button"
      >
        Login
      </button>
      <button
        onClick={() => {
          void logout();
        }}
        type="button"
      >
        Logout
      </button>
    </>
  );
};

/**
 * Renders AuthProvider with a fake app API provider for `/api/auth/me`.
 */
const renderWithApi = (component: ReactElement, ...responses: Response[]) => {
  return render(<ApiProviderStub responses={responses}>{component}</ApiProviderStub>);
};

/**
 * Minimal API provider stub that serves queued `get(...).execute()` responses.
 */
const ApiProviderStub = ({ children, responses }: PropsWithChildren<{ responses: Response[] }>) => {
  const queuedResponses = [...responses];
  const api: ApiContextValue = {
    delete: createUnsupportedApiBuilder,
    get: () => createApiBuilderStub(queuedResponses.shift() ?? responseWithStatus(500)),
    post: createUnsupportedApiBuilder,
    put: createUnsupportedApiBuilder,
  };

  return (
    <ApiContext.Provider value={api}>
      <AuthProvider>{children}</AuthProvider>
    </ApiContext.Provider>
  );
};

/**
 * Creates an API request builder stub for AuthProvider's current-user loader.
 */
const createApiBuilderStub = (response: Response): ApiRequestBuilder => {
  return {
    execute: async () => response,
  } as unknown as ApiRequestBuilder;
};

/**
 * Fails tests that accidentally call an unsupported fake API method.
 */
const createUnsupportedApiBuilder = (): ApiRequestBuilder => {
  throw new Error("Unsupported API method in AuthProvider test.");
};

/**
 * Authenticated current-user fixture returned by `/api/auth/me`.
 */
const authenticatedMe: AuthMeDTO = {
  authenticated: true,
  auth_type: "password",
  roles: ["admin"],
  scopes: ["admin"],
  user: {
    id: "admin-id",
    login: "admin",
  },
};

/**
 * Creates a minimal response with a JSON body.
 */
const responseWithJson = (status: number, body: unknown): Response => {
  return new Response(JSON.stringify(body), {
    headers: { "Content-Type": "application/json" },
    status,
  });
};

/**
 * Creates a minimal response with the supplied status.
 */
const responseWithStatus = (status: number): Response => {
  return new Response(null, { status });
};
