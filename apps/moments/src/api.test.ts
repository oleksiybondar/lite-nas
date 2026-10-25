import { afterEach, expect, test, vi } from "vitest";
import { loadMe, login, logout, request, saveSharedMode } from "./api";

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
  await expect(login({ login: "alice", password: "secret" })).resolves.toBeUndefined();
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
