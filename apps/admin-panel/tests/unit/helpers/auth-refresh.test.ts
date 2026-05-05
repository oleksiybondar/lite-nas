import { refreshAuth } from "@helpers/auth-refresh";

describe("refreshAuth", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(null, { status: 200 })));
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  test("posts an empty refresh request with browser credentials", async () => {
    const response = await refreshAuth();

    expect(response.status).toBe(200);
    expect(fetch).toHaveBeenCalledWith("/api/auth/refresh", {
      body: JSON.stringify({}),
      credentials: "include",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      method: "POST",
    });
  });
});
