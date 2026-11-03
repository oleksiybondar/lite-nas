import { ApiContext } from "@contexts/api-context";
import type { ApiPayload } from "@dto/api/api";
import { type ApiRequestBuilder, createApiRequestBuilder } from "@helpers/api-request-builder";
import { refreshAuth } from "@helpers/auth-refresh";
import {
  createFetchRequestBuilder,
  type FetchRequestBuilder,
  type FetchUnauthorizedContext,
  type FetchUnauthorizedResult,
} from "@helpers/fetch-request-builder";
import { authSessionResponseSchema } from "@schemas/auth/session";
import type { PropsWithChildren, ReactElement } from "react";

const unauthorizedStatus = 401;
const sessionExpiredEvent = "lite-nas-admin-session-expired";

/** Clears stale auth state so the route guard redirects to login. */
const expireSession = (): void => {
  window.dispatchEvent(new Event(sessionExpiredEvent));
};

/**
 * Provides the app API client to React components.
 *
 * The provider owns LiteNAS BFF-specific behavior: every request created through
 * these methods receives the shared unauthorized handler, which attempts session
 * refresh before sending the user to the login route.
 */
export const ApiProvider = ({ children }: PropsWithChildren): ReactElement => {
  /**
   * Builds a `DELETE` request through the app API pipeline.
   */
  const del = (url: string): ApiRequestBuilder => {
    return makeApiRequest(url).method("DELETE");
  };

  /**
   * Builds a `GET` request through the app API pipeline.
   */
  const get = (url: string): ApiRequestBuilder => {
    return makeApiRequest(url).method("GET");
  };

  /**
   * Builds a `POST` request through the app API pipeline.
   *
   * Payloads are optional so endpoints with empty POST bodies can use the same
   * method without forcing callers to pass placeholder objects.
   */
  const post = (url: string, payload?: ApiPayload): ApiRequestBuilder => {
    const request = makeApiRequest(url).method("POST");

    if (payload !== undefined) {
      request.payload(payload);
    }

    return request;
  };

  /**
   * Builds a `PUT` request through the app API pipeline.
   */
  const put = (url: string, payload?: ApiPayload): ApiRequestBuilder => {
    const request = makeApiRequest(url).method("PUT");

    if (payload !== undefined) {
      request.payload(payload);
    }

    return request;
  };

  return (
    <ApiContext.Provider value={{ delete: del, get, post, put }}>{children}</ApiContext.Provider>
  );
};

/**
 * Creates a request builder with the app's shared unauthorized response flow.
 *
 * A `401` from the original API request attempts `/api/auth/refresh`. A
 * successful refresh retries the original request once. Failed renewal clears
 * auth state so the route guard moves the user to `/login`.
 */
const makeApiRequest = (url: string): ApiRequestBuilder => {
  return createApiRequestBuilder(makeFetchRequest(url));
};

/**
 * Creates a protected fetch builder for gateway-backed API calls.
 */
const makeFetchRequest = (url: string): FetchRequestBuilder => {
  return createFetchRequestBuilder(url, {
    unauthorizedHandler: onUnauthorizedHandler,
  }).credentials("include");
};

/**
 * Handles unauthorized protected fetch responses for gateway-backed requests.
 *
 * A 401 response attempts to refresh the browser session through the gateway
 * refresh endpoint. When refresh succeeds, the original request is retried once
 * and exposed to the API layer. When refresh fails, the route guard navigates
 * to login and the response is non-notifiable for API hooks.
 */
const onUnauthorizedHandler = async ({
  response,
  retry,
}: FetchUnauthorizedContext): Promise<FetchUnauthorizedResult> => {
  let refreshResponse: Response;
  try {
    refreshResponse = await refreshAuth();
  } catch {
    expireSession();
    return { notifyHooks: false, response };
  }

  if (!refreshResponse.ok) {
    expireSession();
    return {
      notifyHooks: false,
      response,
    };
  }

  try {
    authSessionResponseSchema.parse(await refreshResponse.clone().json());
  } catch {
    expireSession();
    return { notifyHooks: false, response };
  }

  const retried = await retry();
  if (retried.status === unauthorizedStatus) {
    expireSession();
  }

  return {
    notifyHooks: retried.status !== unauthorizedStatus,
    response: retried,
  };
};
