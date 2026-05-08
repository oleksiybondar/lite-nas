import { useApi } from "@hooks/useApi";
import { ApiProvider } from "@providers/ApiProvider";
import { router } from "@routes/router";
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

describe("ApiProvider unauthorized handling", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  test("refreshes and retries protected requests after unauthorized responses", async () => {
    const fetchMock = mockFetch(
      responseWithStatus(401),
      responseWithStatus(204),
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

  test("navigates to login when session refresh is unauthorized", async () => {
    const navigate = vi.spyOn(router, "navigate").mockResolvedValue(undefined);
    mockFetch(responseWithStatus(401), responseWithStatus(401));

    render(
      <ApiProvider>
        <ApiConsumer />
      </ApiProvider>,
    );
    fireEvent.click(screen.getByRole("button", { name: "Run protected request" }));

    await waitFor(() => {
      expect(navigate).toHaveBeenCalledWith("/login");
    });
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
