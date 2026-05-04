/**
 * Refreshes the browser session through the gateway-owned auth endpoint.
 *
 * This helper intentionally uses direct `fetch` instead of the app API builder
 * so refresh requests cannot recursively trigger the API provider's `401`
 * refresh handling.
 */
export const refreshAuth = (): Promise<Response> => {
  return fetch("/api/auth/refresh", {
    body: JSON.stringify({}),
    credentials: "include",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    method: "POST",
  });
};
