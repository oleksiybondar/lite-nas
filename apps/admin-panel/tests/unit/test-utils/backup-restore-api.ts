import { vi } from "vitest";

const timestamp = "2026-11-10T12:00:00Z";

/** Creates a successful gateway response accepted by backup restore schemas. */
export const backupResponse = (data: unknown): Response =>
  new Response(JSON.stringify({ success: true, timestamp, data }), { status: 200 });

/** Describes one predictable API response in a test's route table. */
export type ApiResponseRoute = {
  matches: (path: string) => boolean;
  respond: () => Response | Promise<Response>;
};

/** Matches one exact API path. */
export const exactApiRoute = (
  path: string,
  respond: ApiResponseRoute["respond"],
): ApiResponseRoute => ({
  matches: (candidate) => candidate === path,
  respond,
});

/** Matches an API path by suffix for resource-specific endpoint families. */
export const suffixApiRoute = (
  suffix: string,
  respond: ApiResponseRoute["respond"],
): ApiResponseRoute => ({
  matches: (candidate) => candidate.endsWith(suffix),
  respond,
});

/** Builds the mocked GET adapter from ordered routes, returning an empty response for misses. */
export const mockBackupGet = (routes: readonly ApiResponseRoute[]) =>
  vi.fn((path: string) => ({
    execute: async () => {
      const route = routes.find((candidate) => candidate.matches(path));
      return route ? route.respond() : backupResponse([]);
    },
  }));
