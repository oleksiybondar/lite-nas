import { ApiContext } from "@contexts/api-context";
import { AuthContext } from "@contexts/auth-context";
import { useAcknowledgeAlert } from "@domain/alerts/hooks/useAcknowledgeAlert";
import type { ApiContextValue } from "@dto/api/api";
import type { ApiRequestBuilder } from "@helpers/api-request-builder";
import { QueryProvider } from "@providers/QueryProvider";
import { renderHook, waitFor } from "@testing-library/react";
import type { PropsWithChildren, ReactNode } from "react";

describe("useAcknowledgeAlert", () => {
  test("posts an empty action body to the expected acknowledge endpoint", async () => {
    const post = createPostStub();
    const { result } = renderHook(() => useAcknowledgeAlert("system"), {
      wrapper: createWrapper(post),
    });

    result.current.mutate({ id: "sysmail_90031349" });

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(post).toHaveBeenCalledWith("/api/alerts/system/sysmail_90031349/acknowledge", {});
  });
});

/**
 * Creates a stable `post` stub that resolves with one successful action response.
 */
const createPostStub = (): ApiContextValue["post"] => {
  return vi.fn((url: string, payload?: unknown) =>
    createApiBuilderStub(url, payload, responseWithJson(200, createActionResponseBody())),
  );
};

/**
 * Creates a wrapper with auth, app API context, and query provider.
 */
const createWrapper = (
  post: ApiContextValue["post"],
): ((props: PropsWithChildren) => ReactNode) => {
  return ({ children }: PropsWithChildren): ReactNode => {
    const api: ApiContextValue = {
      delete: createUnsupportedApiBuilder,
      get: createUnsupportedApiBuilder,
      post,
      put: createUnsupportedApiBuilder,
    };

    return (
      <AuthContext.Provider
        value={{
          isAuthInited: true,
          isAuthenticated: true,
          login: vi.fn(),
          logout: vi.fn(),
          me: {
            authenticated: true,
            auth_type: "password",
            roles: ["admin"],
            scopes: ["admin"],
            user: {
              full_name: "Admin User",
              id: "admin-id",
              login: "admin",
            },
          },
        }}
      >
        <ApiContext.Provider value={api}>
          <QueryProvider>{children}</QueryProvider>
        </ApiContext.Provider>
      </AuthContext.Provider>
    );
  };
};

/**
 * Creates an API request builder stub for acknowledge-hook tests.
 */
const createApiBuilderStub = (
  url: string,
  payload: unknown,
  response: Response,
): ApiRequestBuilder => {
  return {
    execute: async () => response,
    method: vi.fn(() => {
      throw new Error(
        `Unexpected method override for ${url} with payload ${JSON.stringify(payload)}.`,
      );
    }),
  } as unknown as ApiRequestBuilder;
};

/**
 * Fails tests that accidentally call an unsupported fake API method.
 */
const createUnsupportedApiBuilder = (): ApiRequestBuilder => {
  throw new Error("Unsupported API method in acknowledge hook test.");
};

/**
 * Creates a minimal successful action-response payload.
 */
const createActionResponseBody = () => {
  return {
    message: "alert acknowledged",
    success: true,
    timestamp: "2026-06-06T16:12:11.941182868+02:00",
  };
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
