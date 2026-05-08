import {
  type RequestBeforeSendContext,
  type RequestBeforeSendHandler,
  RequestBuilder,
  type RequestEndedHandler,
  type RequestErrorHandler,
  type RequestExecuteOptions,
  type RequestSuccessHandler,
} from "@helpers/request-builder";

/**
 * Body values accepted by `FetchRequestBuilder.payload`.
 *
 * Plain objects and arrays are treated as JSON payloads. Native `BodyInit`
 * values are passed through so callers can still send browser-supported body
 * formats when an endpoint is not JSON-based.
 */
export type FetchRequestPayload = BodyInit | Record<string, unknown> | unknown[] | null;

/**
 * Context passed to protected fetch unauthorized handlers.
 */
export type FetchUnauthorizedContext = {
  /**
   * Unauthorized response returned by the original protected request.
   */
  response: Response;
  /**
   * Replays the original fetch once without re-entering unauthorized handling.
   */
  retry: () => Promise<Response>;
};

/**
 * Result returned by protected fetch unauthorized handlers.
 */
export type FetchUnauthorizedResult = {
  /**
   * Whether upper request layers should run their lifecycle hooks.
   */
  notifyHooks: boolean;
  /**
   * Final response that public callers receive.
   */
  response: Response;
};

/**
 * Handler for protected fetch requests that receive `401`.
 */
export type FetchUnauthorizedHandler = (
  context: FetchUnauthorizedContext,
) => Promise<FetchUnauthorizedResult> | FetchUnauthorizedResult;

/**
 * Result consumed by wrappers around a fetch request builder.
 */
export type FetchRequestExecutionResult = {
  /**
   * Whether upper request layers should run their lifecycle hooks.
   */
  notifyHooks: boolean;
  /**
   * Final fetch response.
   */
  response: Response;
};

/**
 * Options for creating a fetch request builder.
 */
export type FetchRequestBuilderOptions = {
  /**
   * Optional handler for protected `401` responses.
   */
  unauthorizedHandler?: FetchUnauthorizedHandler;
};

/**
 * Fetch-specific request builder.
 *
 * This builder owns direct browser `fetch` execution, request shape, credentials
 * and optional BFF unauthorized handling. It composes `RequestBuilder` for
 * lifecycle hooks instead of duplicating hook ordering.
 */
export class FetchRequestBuilder {
  /**
   * Prepared request body.
   *
   * `undefined` means no body was configured; `null` is an explicit fetch body.
   */
  private bodyValue?: BodyInit | null;

  /**
   * Browser credential policy sent with the request.
   */
  private credentialsValue?: RequestCredentials;

  /**
   * Headers sent with the request.
   */
  private headersValue = new Headers({ Accept: "application/json" });

  /**
   * HTTP method used by the pending request.
   */
  private methodValue = "GET";

  /**
   * Lifecycle builder used for fetch-level hooks.
   */
  private readonly requestBuilder: RequestBuilder;

  /**
   * Optional protected-request handler for `401`.
   */
  private readonly unauthorizedHandler: FetchUnauthorizedHandler | undefined;

  /**
   * Creates a fetch request builder for a fixed URL.
   */
  public constructor(
    private readonly urlValue: string,
    options: FetchRequestBuilderOptions = {},
  ) {
    this.unauthorizedHandler = options.unauthorizedHandler;
    this.requestBuilder = new RequestBuilder(
      (executeOptions) => this.executeAction(executeOptions),
      () => this.readBeforeSendContext(),
    );
  }

  /**
   * Sets the browser credential policy for cookie-backed requests.
   */
  public credentials(credentials: RequestCredentials): this {
    this.credentialsValue = credentials;
    return this;
  }

  /**
   * Executes the fetch request and returns the final response.
   */
  public execute(options: RequestExecuteOptions = {}): Promise<Response> {
    return this.requestBuilder.execute(options);
  }

  /**
   * Executes the fetch request and returns metadata for wrapping builders.
   */
  public async executeForWrapper(
    options: RequestExecuteOptions = {},
  ): Promise<FetchRequestExecutionResult> {
    const outcome = await this.executeAction(options);

    return {
      notifyHooks: outcome.notifyHooks,
      response: outcome.response,
    };
  }

  /**
   * Adds or replaces a request header.
   */
  public header(name: string, value: string): this {
    this.headersValue.set(name, value);
    return this;
  }

  /**
   * Sets the HTTP method used by the request.
   */
  public method(method: string): this {
    this.methodValue = method;
    return this;
  }

  /**
   * Registers a side-effect hook that runs immediately before fetch execution.
   */
  public onBeforeSend(handler: RequestBeforeSendHandler): this {
    this.requestBuilder.onBeforeSend(handler);
    return this;
  }

  /**
   * Registers a side-effect hook that runs after fetch execution.
   */
  public onEnded(handler: RequestEndedHandler): this {
    this.requestBuilder.onEnded(handler);
    return this;
  }

  /**
   * Registers a handler for failed fetch responses.
   */
  public onError(handler: RequestErrorHandler): this {
    this.requestBuilder.onError(handler);
    return this;
  }

  /**
   * Registers a handler for successful fetch responses.
   */
  public onSuccess(handler: RequestSuccessHandler): this {
    this.requestBuilder.onSuccess(handler);
    return this;
  }

  /**
   * Sets the request payload.
   */
  public payload(payload: FetchRequestPayload): this {
    if (payload === null) {
      this.bodyValue = null;
      return this;
    }

    if (isBodyInit(payload)) {
      this.bodyValue = payload;
      return this;
    }

    this.bodyValue = JSON.stringify(payload);
    this.headersValue.set("Content-Type", "application/json");
    return this;
  }

  /**
   * Backwards-compatible alias for payload.
   */
  public body(payload: FetchRequestPayload): this {
    return this.payload(payload);
  }

  /**
   * Executes direct fetch and applies optional unauthorized handling.
   */
  private async executeAction(
    options: RequestExecuteOptions,
  ): Promise<{ kind: "error" | "success"; notifyHooks: boolean; response: Response }> {
    const response = await this.executeFetch();

    if (
      response.status === unauthorizedStatus &&
      options.skipErrorHandler !== true &&
      this.unauthorizedHandler !== undefined
    ) {
      const result = await this.unauthorizedHandler({
        response,
        retry: () => this.executeFetch(),
      });

      return {
        kind: result.response.ok ? "success" : "error",
        notifyHooks: result.notifyHooks,
        response: result.response,
      };
    }

    return {
      kind: response.ok ? "success" : "error",
      notifyHooks: true,
      response,
    };
  }

  /**
   * Performs the browser fetch call.
   */
  private async executeFetch(): Promise<Response> {
    return await fetch(this.urlValue, this.buildRequestInit());
  }

  /**
   * Builds the `RequestInit` object without undefined optional properties.
   */
  private buildRequestInit(): RequestInit {
    const requestInit: RequestInit = {
      headers: this.headersValue,
      method: this.methodValue,
    };

    if (this.credentialsValue !== undefined) {
      requestInit.credentials = this.credentialsValue;
    }

    if (this.bodyValue !== undefined) {
      requestInit.body = this.bodyValue;
    }

    return requestInit;
  }

  /**
   * Reads final request metadata for lifecycle hooks.
   */
  public readBeforeSendContext(): RequestBeforeSendContext {
    return {
      headers: this.headersValue,
      method: this.methodValue,
      url: this.urlValue,
    };
  }
}

/**
 * Creates a fetch request builder.
 */
export const createFetchRequestBuilder = (
  url: string,
  options: FetchRequestBuilderOptions = {},
): FetchRequestBuilder => {
  return new FetchRequestBuilder(url, options);
};

/**
 * Detects browser-native body values that should not be JSON serialized.
 */
const isBodyInit = (payload: FetchRequestPayload): payload is BodyInit => {
  return (
    typeof payload === "string" ||
    payload instanceof Blob ||
    payload instanceof FormData ||
    payload instanceof URLSearchParams ||
    payload instanceof ArrayBuffer ||
    ArrayBuffer.isView(payload) ||
    payload instanceof ReadableStream
  );
};

/**
 * HTTP status used by protected endpoints when the access token is invalid.
 */
const unauthorizedStatus = 401;
