import type { ApiRequestBuilder } from "@helpers/api-request-builder";

/**
 * Payload accepted by the app API helpers.
 *
 * Use plain objects or arrays for JSON DTO calls. Use native `BodyInit` values
 * such as `FormData`, `URLSearchParams`, or strings when an endpoint needs a
 * non-JSON body.
 */
export type ApiPayload = BodyInit | Record<string, unknown> | unknown[] | null;

/**
 * React-facing API client exposed by `ApiProvider`.
 *
 * The methods are intentionally small wrappers around the request builder:
 * `get`, `post`, `put`, and `delete` preselect the HTTP method, apply the app's
 * default JSON headers, and return a chainable API action.
 *
 * When a request returns `401`, the provider attempts the BFF refresh endpoint.
 * If refresh succeeds, the original request is retried once and that response is
 * returned. If refresh also returns `401`, the app navigates to `/login` and the
 * original unauthorized response is returned to the caller.
 *
 * Example:
 *
 * ```ts
 * const { post } = useApi();
 * const response = await post("/api/shares", { name: "media" }).execute();
 *
 * if (response.ok) {
 *   const body = await response.json();
 * }
 * ```
 */
export type ApiContextValue = {
  /**
   * Builds a `DELETE` request action.
   */
  delete: (url: string) => ApiRequestBuilder;
  /**
   * Builds a `GET` request action.
   */
  get: (url: string) => ApiRequestBuilder;
  /**
   * Builds a `POST` request action with an optional payload.
   */
  post: (url: string, payload?: ApiPayload) => ApiRequestBuilder;
  /**
   * Builds a `PUT` request action with an optional payload.
   */
  put: (url: string, payload?: ApiPayload) => ApiRequestBuilder;
};
