import type { FetchRequestBuilder, FetchRequestPayload } from "@helpers/fetch-request-builder";
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
 * Application-facing API request builder.
 *
 * This builder wraps a fetch request builder and exposes the same lifecycle
 * shape at the BFF API-action level. Transport details such as direct fetch,
 * cookies, refresh and retry stay inside the fetch builder dependency.
 */
export class ApiRequestBuilder {
  /**
   * Lifecycle builder used for application-level API hooks.
   */
  private readonly requestBuilder: RequestBuilder;

  /**
   * Creates an API request builder around a fetch builder dependency.
   */
  public constructor(private readonly fetchRequestBuilder: FetchRequestBuilder) {
    this.requestBuilder = new RequestBuilder(
      (options) => this.executeAction(options),
      () => this.readBeforeSendContext(),
    );
  }

  /**
   * Overrides the browser credential policy on the wrapped fetch request.
   */
  public credentials(credentials: RequestCredentials): this {
    this.fetchRequestBuilder.credentials(credentials);
    return this;
  }

  /**
   * Executes the API action and returns the final response.
   */
  public execute(options: RequestExecuteOptions = {}): Promise<Response> {
    return this.requestBuilder.execute(options);
  }

  /**
   * Adds or replaces an API request header.
   */
  public header(name: string, value: string): this {
    this.fetchRequestBuilder.header(name, value);
    return this;
  }

  /**
   * Sets the HTTP method on the wrapped fetch request.
   */
  public method(method: string): this {
    this.fetchRequestBuilder.method(method);
    return this;
  }

  /**
   * Registers a side-effect hook that runs immediately before the API action.
   */
  public onBeforeSend(handler: RequestBeforeSendHandler): this {
    this.requestBuilder.onBeforeSend(handler);
    return this;
  }

  /**
   * Registers a side-effect hook that runs after the API action completes.
   */
  public onEnded(handler: RequestEndedHandler): this {
    this.requestBuilder.onEnded(handler);
    return this;
  }

  /**
   * Registers a handler for failed API action responses.
   */
  public onError(handler: RequestErrorHandler): this {
    this.requestBuilder.onError(handler);
    return this;
  }

  /**
   * Registers a handler for successful API action responses.
   */
  public onSuccess(handler: RequestSuccessHandler): this {
    this.requestBuilder.onSuccess(handler);
    return this;
  }

  /**
   * Sets the API request payload.
   */
  public payload(payload: FetchRequestPayload): this {
    this.fetchRequestBuilder.payload(payload);
    return this;
  }

  /**
   * Backwards-compatible alias for payload.
   */
  public body(payload: FetchRequestPayload): this {
    return this.payload(payload);
  }

  /**
   * Executes the wrapped fetch builder as an API action.
   *
   * The fetch builder may mark a response as non-notifiable when transport auth
   * redirects to login. In that case API-level hooks are skipped.
   */
  private async executeAction(
    options: RequestExecuteOptions,
  ): Promise<{ kind: "error" | "success"; notifyHooks: boolean; response: Response }> {
    const result = await this.fetchRequestBuilder.executeForWrapper(options);

    return {
      kind: result.response.ok ? "success" : "error",
      notifyHooks: result.notifyHooks,
      response: result.response,
    };
  }

  /**
   * Reads request metadata from the wrapped fetch builder for API hooks.
   */
  private readBeforeSendContext(): RequestBeforeSendContext {
    return this.fetchRequestBuilder.readBeforeSendContext();
  }
}

/**
 * Creates an API request builder around the supplied fetch builder.
 */
export const createApiRequestBuilder = (
  fetchRequestBuilder: FetchRequestBuilder,
): ApiRequestBuilder => {
  return new ApiRequestBuilder(fetchRequestBuilder);
};
