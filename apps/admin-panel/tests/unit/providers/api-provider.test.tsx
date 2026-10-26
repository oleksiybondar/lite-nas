import { useApi } from "@hooks/useApi";
import { ApiProvider } from "@providers/ApiProvider";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import type { ReactElement } from "react";

describe("ApiProvider request methods", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  test("builds method-specific requests through the context API", async () => {
    const fetchMock = mockFetch(responseWithStatus(200), responseWithStatus(201));

    render(
      <ApiProvider>
        <ApiConsumer />
      </ApiProvider>,
    );
    fireEvent.click(screen.getByRole("button", { name: "Run API requests" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
    expect(fetchMock).toHaveBeenNthCalledWith(1, "/api/items", {
      credentials: "include",
      headers: expect.any(Headers),
      method: "GET",
    });
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/items", {
      body: JSON.stringify({ name: "media" }),
      credentials: "include",
      headers: expect.any(Headers),
      method: "POST",
    });
  });
});

// biome-ignore lint/complexity/noExcessiveLinesPerFunction: renewal and concurrent renewal cover the same provider contract.
describe("ApiProvider unauthorized handling", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  test("refreshes and retries protected requests after unauthorized responses", async () => {
    const fetchMock = mockFetch(
      responseWithStatus(401),
      sessionResponse(),
      responseWithStatus(200),
    );

    render(
      <ApiProvider>
        <ApiConsumer />
      </ApiProvider>,
    );
    fireEvent.click(screen.getByRole("button", { name: "Run protected request" }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(3);
    });
    expect(fetchMock).toHaveBeenNthCalledWith(2, "/api/auth/refresh", {
      body: JSON.stringify({}),
      credentials: "include",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      method: "POST",
    });
  });

  test("expires auth state when session refresh is unauthorized", async () => {
    const dispatch = vi.spyOn(window, "dispatchEvent");
    mockFetch(responseWithStatus(401), responseWithStatus(401));

    render(
      <ApiProvider>
        <ApiConsumer />
      </ApiProvider>,
    );
    fireEvent.click(screen.getByRole("button", { name: "Run protected request" }));

    await waitFor(() => {
      expect(dispatch).toHaveBeenCalledWith(
        expect.objectContaining({ type: "lite-nas-admin-session-expired" }),
      );
    });
  });

  test("expires auth state when a renewed request is still unauthorized", async () => {
    const dispatch = vi.spyOn(window, "dispatchEvent");
    mockFetch(responseWithStatus(401), sessionResponse(), responseWithStatus(401));

    render(
      <ApiProvider>
        <ApiConsumer />
      </ApiProvider>,
    );
    fireEvent.click(screen.getByRole("button", { name: "Run protected request" }));

    await waitFor(() => {
      expect(dispatch).toHaveBeenCalledWith(
        expect.objectContaining({ type: "lite-nas-admin-session-expired" }),
      );
    });
  });

  test("expires auth state when refresh returns malformed session data", async () => {
    const dispatch = vi.spyOn(window, "dispatchEvent");
    mockFetch(responseWithStatus(401), new Response("{}", { status: 200 }));
    render(
      <ApiProvider>
        <ApiConsumer />
      </ApiProvider>,
    );
    fireEvent.click(screen.getByRole("button", { name: "Run protected request" }));
    await waitFor(() =>
      expect(dispatch).toHaveBeenCalledWith(
        expect.objectContaining({ type: "lite-nas-admin-session-expired" }),
      ),
    );
  });

  test("simultaneous unauthorized requests share one gateway refresh", async () => {
    let completeRefresh: (response: Response) => void = () => undefined;
    const refresh = new Promise<Response>((resolve) => {
      completeRefresh = resolve;
    });
    // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: the test transport returns a held refresh response and staged protected responses.
    const fetchMock = vi.fn((path: string) => {
      if (path === "/api/auth/refresh") return refresh;
      return Promise.resolve(responseWithStatus(fetchMock.mock.calls.length <= 2 ? 401 : 200));
    });
    vi.stubGlobal("fetch", fetchMock);

    render(
      <ApiProvider>
        <ApiConsumer />
      </ApiProvider>,
    );
    fireEvent.click(screen.getByRole("button", { name: "Run concurrent requests" }));
    await waitFor(() => expect(fetchMock.mock.calls.length).toBe(3));
    completeRefresh(sessionResponse());
    await waitFor(() => expect(fetchMock.mock.calls.length).toBe(5));
    expect(fetchMock.mock.calls.filter(([path]) => path === "/api/auth/refresh")).toHaveLength(1);
  });
});

/**
 * Test consumer that exercises ApiProvider through the public hook contract.
 */
const ApiConsumer = (): ReactElement => {
  const { get, post } = useApi();

  return (
    <>
      <button
        onClick={() => {
          void get("/api/items").execute();
          void post("/api/items", { name: "media" }).execute();
        }}
        type="button"
      >
        Run API requests
      </button>
      <button
        onClick={() => {
          void get("/api/protected").execute();
        }}
        type="button"
      >
        Run protected request
      </button>
      <button
        onClick={() => {
          void get("/api/first").execute();
          void get("/api/second").execute();
        }}
        type="button"
      >
        Run concurrent requests
      </button>
    </>
  );
};

/**
 * Installs a fetch mock that returns responses in call order.
 */
const mockFetch = (...responses: Response[]) => {
  const fetchMock = vi.fn();

  for (const response of responses) {
    fetchMock.mockResolvedValueOnce(response);
  }

  vi.stubGlobal("fetch", fetchMock);
  return fetchMock;
};

/**
 * Creates a minimal fetch response with the supplied status.
 */
const responseWithStatus = (status: number): Response => {
  return new Response(null, { status });
};

/** Creates the gateway session envelope issued after cookie rotation. */
const sessionResponse = (): Response =>
  new Response(JSON.stringify({ data: { authenticated: true, user: { id: "admin-id" } } }), {
    status: 200,
  });
