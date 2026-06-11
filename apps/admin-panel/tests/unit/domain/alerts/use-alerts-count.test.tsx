import { ApiContext } from "@contexts/api-context";
import { useAlertsCount } from "@domain/alerts/hooks/useAlertsCount";
import type { ApiContextValue } from "@dto/api/api";
import type { ApiRequestBuilder } from "@helpers/api-request-builder";
import { QueryProvider } from "@providers/QueryProvider";
import { renderHook, waitFor } from "@testing-library/react";
import type { PropsWithChildren, ReactNode } from "react";

describe("useAlertsCount", () => {
  describe("request execution", () => {
    test("loads the system alerts count from the expected endpoint", async () => {
      const get = createGetStub(3);
      const { result } = renderHook(() => useAlertsCount("system"), {
        wrapper: createWrapper(get),
      });

      await waitForSuccess(result);

      expect(get).toHaveBeenCalledWith("/api/alerts/system/unacknowledged/count");
      expect(result.current.data).toEqual({ count: 3 });
    });

    test("loads the security alerts count from the expected endpoint", async () => {
      const get = createGetStub(7);
      const { result } = renderHook(() => useAlertsCount("security"), {
        wrapper: createWrapper(get),
      });

      await waitForSuccess(result);

      expect(get).toHaveBeenCalledWith("/api/alerts/security/unacknowledged/count");
      expect(result.current.data).toEqual({ count: 7 });
    });
  });

  describe("query enablement", () => {
    test("does not execute when disabled", async () => {
      const get = createGetStub(1);
      const { result } = renderHook(() => useAlertsCount("system", { enabled: false }), {
        wrapper: createWrapper(get),
      });

      await waitFor(() => {
        expect(result.current.fetchStatus).toBe("idle");
      });

      expect(get).not.toHaveBeenCalled();
      expect(result.current.data).toBeUndefined();
    });
  });
});

/**
 * Creates a stable `get` stub that resolves with one alert-count response.
 */
const createGetStub = (count: number): ApiContextValue["get"] => {
  return vi.fn((url: string) => createApiBuilderStub(url, createAlertCountResponse(count)));
};

/**
 * Waits until a query-hook result reaches a successful state.
 */
const waitForSuccess = async (
  result: ReturnType<typeof renderHook<ReturnType<typeof useAlertsCount>, unknown>>["result"],
): Promise<void> => {
  await waitFor(() => {
    expect(result.current.isSuccess).toBe(true);
  });
};

/**
 * Creates a successful alert-count response body.
 */
const createAlertCountResponse = (count: number): Response => {
  return responseWithJson(200, {
    data: { count },
    success: true,
    timestamp: "2026-06-04T10:00:00Z",
  });
};

/**
 * Creates a wrapper with the app API context and query provider.
 */
const createWrapper = (get: ApiContextValue["get"]): ((props: PropsWithChildren) => ReactNode) => {
  return ({ children }: PropsWithChildren): ReactNode => {
    const api: ApiContextValue = {
      delete: createUnsupportedApiBuilder,
      get,
      post: createUnsupportedApiBuilder,
      put: createUnsupportedApiBuilder,
    };

    return (
      <ApiContext.Provider value={api}>
        <QueryProvider>{children}</QueryProvider>
      </ApiContext.Provider>
    );
  };
};

/**
 * Creates an API request builder stub for query-hook tests.
 */
const createApiBuilderStub = (url: string, response: Response): ApiRequestBuilder => {
  return {
    execute: async () => response,
    method: vi.fn(() => {
      throw new Error(`Unexpected method override for ${url}.`);
    }),
  } as unknown as ApiRequestBuilder;
};

/**
 * Fails tests that accidentally call an unsupported fake API method.
 */
const createUnsupportedApiBuilder = (): ApiRequestBuilder => {
  throw new Error("Unsupported API method in alerts query test.");
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
