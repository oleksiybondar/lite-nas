/**
 * Refreshes the browser session through the gateway-owned auth endpoint.
 *
 * This helper intentionally uses direct `fetch` instead of the app API builder
 * so refresh requests cannot recursively trigger the API provider's `401`
 * refresh handling.
 */
let pendingRefresh: Promise<Response> | null = null;

export const refreshAuth = (): Promise<Response> => {
  if (pendingRefresh) return pendingRefresh;
  pendingRefresh = (async () => {
    try {
      return await fetch("/api/auth/refresh", {
        body: JSON.stringify({}),
        credentials: "include",
        headers: {
          Accept: "application/json",
          "Content-Type": "application/json",
        },
        method: "POST",
      });
    } finally {
      pendingRefresh = null;
    }
  })();
  return pendingRefresh;
};
