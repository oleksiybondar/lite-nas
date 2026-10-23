import { afterEach, expect, test, vi } from "vitest";
import { loadMe, request } from "./api";

afterEach(() => vi.unstubAllGlobals());

// Requirements: moments/FR-001
test("current user response is validated before use", async () => {
  vi.stubGlobal(
    "fetch",
    vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: {
            authenticated: true,
            auth_type: "jwt",
            user: { id: "1", login: "alice" },
            roles: [],
            scopes: [],
          },
        }),
        { status: 200 },
      ),
    ),
  );
  await expect(loadMe()).resolves.toMatchObject({ user: { login: "alice" } });
});

// Requirements: moments/FR-001
test("malformed current user response is rejected", async () => {
  vi.stubGlobal(
    "fetch",
    vi
      .fn()
      .mockResolvedValue(
        new Response(JSON.stringify({ data: { authenticated: true } }), { status: 200 }),
      ),
  );
  await expect(loadMe()).rejects.toThrow();
});

// Requirements: moments/FR-004
test("profile request retries once after refreshing an expired access cookie", async () => {
  const fetchMock = vi
    .fn()
    .mockResolvedValueOnce(new Response(null, { status: 401 }))
    .mockResolvedValueOnce(
      new Response(JSON.stringify({ data: { authenticated: true, user: { id: "1" } } }), {
        status: 200,
      }),
    )
    .mockResolvedValueOnce(new Response("{}", { status: 200 }));
  vi.stubGlobal("fetch", fetchMock);
  await expect(request("/f/profile/ui")).resolves.toHaveProperty("status", 200);
});
