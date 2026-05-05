import { createFetchRequestBuilder } from "@helpers/fetch-request-builder";

beforeEach(() => {
  vi.stubGlobal("fetch", vi.fn());
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("FetchRequestBuilder request assembly", () => {
  test("builds a JSON request and exposes fetch lifecycle hooks", async () => {
    const response = responseWithStatus(200);
    vi.mocked(fetch).mockResolvedValue(response);
    const beforeSend = vi.fn();
    const ended = vi.fn();

    const result = await createFetchRequestBuilder("/api/test")
      .method("POST")
      .credentials("include")
      .header("X-Test", "1")
      .payload({ name: "media" })
      .onBeforeSend(beforeSend)
      .onEnded(ended)
      .execute();

    expect(result).toBe(response);
    expect(fetch).toHaveBeenCalledWith("/api/test", {
      body: JSON.stringify({ name: "media" }),
      credentials: "include",
      headers: expect.any(Headers),
      method: "POST",
    });
    expect(requestHeaders()).toEqual({
      accept: "application/json",
      contentType: "application/json",
      test: "1",
    });
    expect(beforeSend).toHaveBeenCalledWith({
      headers: expect.any(Headers),
      method: "POST",
      url: "/api/test",
    });
    expect(ended).toHaveBeenCalledWith({ response });
  });
});

describe("FetchRequestBuilder payload handling", () => {
  test("passes native body values without JSON serialization", async () => {
    const body = new URLSearchParams({ name: "media" });
    vi.mocked(fetch).mockResolvedValue(responseWithStatus(204));

    await createFetchRequestBuilder("/api/test").method("POST").payload(body).execute();

    expect(vi.mocked(fetch).mock.calls[0]?.[1]).toMatchObject({
      body,
      method: "POST",
    });
  });
});

describe("FetchRequestBuilder unauthorized handling", () => {
  test("uses unauthorized handler and exposes retry response to wrappers", async () => {
    const unauthorizedResponse = responseWithStatus(401);
    const retryResponse = responseWithStatus(200);
    vi.mocked(fetch)
      .mockResolvedValueOnce(unauthorizedResponse)
      .mockResolvedValueOnce(retryResponse);

    const result = await createFetchRequestBuilder("/api/protected", {
      unauthorizedHandler: async ({ response, retry }) => {
        expect(response).toBe(unauthorizedResponse);
        return {
          notifyHooks: true,
          response: await retry(),
        };
      },
    }).executeForWrapper();

    expect(result).toEqual({
      notifyHooks: true,
      response: retryResponse,
    });
    expect(fetch).toHaveBeenCalledTimes(2);
  });

  test("does not enter unauthorized handler when error handling is skipped", async () => {
    const unauthorizedHandler = vi.fn();
    vi.mocked(fetch).mockResolvedValue(responseWithStatus(401));

    const response = await createFetchRequestBuilder("/api/protected", {
      unauthorizedHandler,
    }).execute({ skipErrorHandler: true });

    expect(response.status).toBe(401);
    expect(unauthorizedHandler).not.toHaveBeenCalled();
  });
});

const responseWithStatus = (status: number): Response => {
  return new Response(null, { status });
};

const requestHeaders = (): {
  accept: string | null;
  contentType: string | null;
  test: string | null;
} => {
  const headers = vi.mocked(fetch).mock.calls[0]?.[1]?.headers;

  if (!(headers instanceof Headers)) {
    throw new Error("Expected fetch headers to be Headers");
  }

  return {
    accept: headers.get("Accept"),
    contentType: headers.get("Content-Type"),
    test: headers.get("X-Test"),
  };
};
