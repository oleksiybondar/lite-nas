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
import { router } from "@routes/router";
import type { PropsWithChildren, ReactElement } from "react";

const loginPath = "/login";
const unauthorizedStatus = 401;

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
 * successful refresh retries the original request once. A refresh `401` means
 * the session cannot be recovered, so the router moves the user to `/login`.
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
 * and exposed to the API layer. When refresh fails with 401, the app navigates
 * to the login route and marks the response as non-notifiable for API hooks.
 */
const onUnauthorizedHandler = async ({
  response,
  retry,
}: FetchUnauthorizedContext): Promise<FetchUnauthorizedResult> => {
  const refreshResponse = await refreshAuth();

  if (!refreshResponse.ok) {
    if (refreshResponse.status === unauthorizedStatus) {
      void router.navigate(loginPath);
    }
    return {
      notifyHooks: refreshResponse.status !== unauthorizedStatus,
      response,
    };
  }

  return {
    notifyHooks: true,
    response: await retry(),
  };
};
