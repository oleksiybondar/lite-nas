import { afterEach, expect, test, vi } from "vitest";
import {
  loadGalleryConcurrency,
  loadMe,
  loadMomentsConfig,
  login,
  logout,
  momentsApiPath,
  request,
  saveGalleryConcurrency,
  saveSharedMode,
} from "./api";

afterEach(() => vi.unstubAllGlobals());

test("embedded Moments uses its dedicated API mount", () => {
  vi.stubGlobal("window", { location: { pathname: "/apps/moments/" } });
  expect(momentsApiPath("/api/galleries")).toBe("/apps/moments/api/galleries");
});

test("Moments picker configuration accepts a clean absolute working directory", async () => {
  vi.stubGlobal(
    "fetch",
    vi
      .fn()
      .mockResolvedValue(
        new Response(JSON.stringify({ workingDir: "/NASPool/Photos" }), { status: 200 }),
      ),
  );
  await expect(loadMomentsConfig()).resolves.toBe("/NASPool/Photos");
});

test("Moments picker configuration rejects a relative working directory", async () => {
  vi.stubGlobal(
    "fetch",
    vi
      .fn()
      .mockResolvedValue(new Response(JSON.stringify({ workingDir: "Photos" }), { status: 200 })),
  );
  await expect(loadMomentsConfig()).rejects.toThrow();
});

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

test("simultaneous expired requests share one cookie refresh", async () => {
  let completeRefresh: (response: Response) => void = () => undefined;
  const refresh = new Promise<Response>((resolve) => {
    completeRefresh = resolve;
  });
  const fetchMock = vi.fn((path: string) => {
    if (path.endsWith("/auth/refresh")) return refresh;
    return Promise.resolve(
      new Response(null, { status: fetchMock.mock.calls.length <= 2 ? 401 : 200 }),
    );
  });
  vi.stubGlobal("fetch", fetchMock);
  const requests = Promise.all([request("/f/one"), request("/f/two")]);
  await vi.waitFor(() => expect(fetchMock.mock.calls.length).toBe(3));
  completeRefresh(
    new Response(JSON.stringify({ data: { authenticated: true, user: { id: "1" } } }), {
      status: 200,
    }),
  );
  await requests;
  expect(fetchMock.mock.calls.filter(([path]) => path === "/api/auth/refresh")).toHaveLength(1);
});

// Requirements: moments/FR-001. Login sends a validated credential pair.
test("login accepts a valid authenticated session response", async () => {
  vi.stubGlobal(
    "fetch",
    vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ data: { authenticated: true, user: { id: "alice" } } }), {
        status: 200,
      }),
    ),
  );
  await expect(login({ login: "alice", password: "secret" })).resolves.toBe("/");
});

test("login sends the requested return path", async () => {
  const fetchMock = vi.fn().mockResolvedValue(
    new Response(
      JSON.stringify({
        data: { authenticated: true, user: { id: "alice" }, redirect_to: "/galleries/saved" },
      }),
    ),
  );
  vi.stubGlobal("fetch", fetchMock);
  await login({ login: "alice", password: "secret" }, "/galleries/saved");
  expect(fetchMock.mock.calls[0][0]).toBe("/api/auth/login?redirect_to=%2Fgalleries%2Fsaved");
});

test("login uses the validated return path echoed by the backend", async () => {
  vi.stubGlobal(
    "fetch",
    vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          data: { authenticated: true, user: { id: "alice" }, redirect_to: "/galleries/saved" },
        }),
      ),
    ),
  );
  await expect(login({ login: "alice", password: "secret" }, "/galleries/saved")).resolves.toBe(
    "/galleries/saved",
  );
});

// Requirements: moments/FR-001. Invalid credentials never reach the transport.
test("login rejects malformed credentials before requesting auth", async () => {
  const fetchMock = vi.fn();
  vi.stubGlobal("fetch", fetchMock);
  await expect(login({ login: "bad name", password: "x" })).rejects.toThrow();
});

// Requirements: moments/FR-001. Logout propagates successful revocation.
test("logout accepts a successful revocation response", async () => {
  vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(null, { status: 200 })));
  await expect(logout()).resolves.toBeUndefined();
});

// Requirements: moments/FR-002. Shared appearance changes use the profile route.
test("shared appearance writes a validated profile document", async () => {
  vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(null, { status: 200 })));
  await expect(saveSharedMode("light")).resolves.toBeUndefined();
});

test("Moments gallery settings load from the authenticated profile file", async () => {
  vi.stubGlobal(
    "fetch",
    vi
      .fn()
      .mockResolvedValue(
        new Response(JSON.stringify({ gallery: { maxConcurrentImageLoads: 12 } }), { status: 200 }),
      ),
  );
  await expect(loadGalleryConcurrency()).resolves.toBe(12);
});

test("Moments gallery settings default to ten when the profile file is absent", async () => {
  vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(null, { status: 404 })));
  await expect(loadGalleryConcurrency()).resolves.toBe(10);
});

test("Moments gallery settings use a validated browser value when the profile file is absent", async () => {
  vi.stubGlobal("window", { localStorage: { getItem: () => "12" } });
  vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(null, { status: 404 })));
  await expect(loadGalleryConcurrency()).resolves.toBe(12);
});

test("Moments gallery settings write a grouped document through File Service", async () => {
  const fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 201 }));
  vi.stubGlobal("fetch", fetchMock);
  await saveGalleryConcurrency(12);
  expect(fetchMock).toHaveBeenCalledWith(
    "/f/profile/moments",
    expect.objectContaining({
      method: "PUT",
      body: JSON.stringify({ document: { gallery: { maxConcurrentImageLoads: 12 } } }),
    }),
  );
});

test("Moments gallery settings reject invalid concurrency before writing", async () => {
  const fetchMock = vi.fn();
  vi.stubGlobal("fetch", fetchMock);
  await expect(saveGalleryConcurrency(21)).rejects.toThrow();
});

// Requirements: moments/FR-001. Auth failures cannot trigger recursive refresh.
test("auth route rejection returns without another refresh attempt", async () => {
  const fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 401 }));
  vi.stubGlobal("fetch", fetchMock);
  await request("/api/auth/login");
  expect(fetchMock).toHaveBeenCalledTimes(1);
});

// Requirements: moments/FR-001. Failed rotation leaves a protected request unauthorized.
test("protected request stays unauthorized when refresh fails", async () => {
  vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(null, { status: 401 })));
  await expect(request("/f/profile/ui")).resolves.toHaveProperty("status", 401);
});

// Requirements: moments/FR-001. An expired browser session resolves to no current user.
test("current user is absent after refresh also expires", async () => {
  vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(null, { status: 401 })));
  await expect(loadMe()).resolves.toBeNull();
});

// Requirements: moments/FR-001. A rotated session restores the current user.
test("current user loads after refreshing an expired access cookie", async () => {
  vi.stubGlobal(
    "fetch",
    vi
      .fn()
      .mockResolvedValueOnce(new Response(null, { status: 401 }))
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ data: { authenticated: true, user: { id: "alice" } } }), {
          status: 200,
        }),
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            data: {
              authenticated: true,
              auth_type: "jwt",
              user: { id: "alice", login: "alice" },
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

// Requirements: moments/FR-001. Revocation errors remain visible to the caller.
test("logout rejects a failed revocation response", async () => {
  vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(null, { status: 503 })));
  await expect(logout()).rejects.toThrow("Sign out failed.");
});

// Requirements: moments/FR-001. An unavailable current-user endpoint is an error.
test("current user rejects a service failure", async () => {
  vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(null, { status: 503 })));
  await expect(loadMe()).rejects.toThrow("Could not load your session.");
});

// Requirements: moments/FR-001. Rejected login responses do not create a session.
test("login rejects failed credential verification", async () => {
  vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(null, { status: 401 })));
  await expect(login({ login: "alice", password: "secret" })).rejects.toThrow("Sign in failed.");
});
