/**
 * Context passed to a request before-send handler.
 *
 * This context is read-only by convention. Lifecycle hooks are intended for side
 * effects such as starting loaders, not for mutating request execution.
 */
export type RequestBeforeSendContext = {
  /**
   * Headers that will be sent with the request when the action is HTTP-backed.
   */
  headers: Headers;
  /**
   * HTTP method that will be sent with the request when the action is
   * HTTP-backed.
   */
  method: string;
  /**
   * URL or action target that will be executed.
   */
  url: string;
};

/**
 * Context passed to a request ended handler.
 */
export type RequestEndedContext =
  | {
      /**
       * Error thrown before a response was produced.
       */
      error: unknown;
      /**
       * Response is unavailable when execution fails before a response exists.
       */
      response?: never;
    }
  | {
      /**
       * Final response returned by `execute` after success or error handling.
       */
      response: Response;
      /**
       * Error is unavailable when the request completed with a response.
       */
      error?: never;
    };

/**
 * Context passed to a request error handler.
 */
export type RequestErrorContext = {
  /**
   * Non-2xx response returned by the action.
   */
  response: Response;
  /**
   * Replays the action once without invoking the same error handler.
   */
  retry: () => Promise<Response>;
};

/**
 * Internal execution options used when a handler replays an action.
 */
export type RequestExecuteOptions = {
  /**
   * Prevents a retry response from recursively entering the same error handler.
   */
  skipErrorHandler?: boolean;
};

/**
 * Result returned by a request action before lifecycle hooks are applied.
 */
export type RequestActionOutcome = {
  /**
   * Whether the response should be processed as success or error.
   */
  kind: "error" | "success";
  /**
   * Whether this outcome should notify lifecycle hooks.
   */
  notifyHooks: boolean;
  /**
   * Response produced by the action.
   */
  response: Response;
};

/**
 * Function that executes the underlying action for a request builder.
 */
export type RequestActionExecutor = (
  options: RequestExecuteOptions,
) => Promise<RequestActionOutcome>;

/**
 * Function that reads before-send metadata for a request builder.
 */
export type RequestBeforeSendContextReader = () => RequestBeforeSendContext;

/**
 * Hook called immediately before action execution.
 */
export type RequestBeforeSendHandler = (context: RequestBeforeSendContext) => Promise<void> | void;

/**
 * Hook called after request completion.
 */
export type RequestEndedHandler = (context: RequestEndedContext) => Promise<void> | void;

/**
 * Optional replacement response returned by success and error hooks.
 */
type RequestHandlerResponse = Promise<Response | undefined> | Response | undefined;

/**
 * Hook called for failed action responses.
 *
 * Returning a response replaces the response that `execute` resolves with.
 */
export type RequestErrorHandler = (context: RequestErrorContext) => RequestHandlerResponse;

/**
 * Hook called for successful action responses.
 *
 * Returning a response replaces the response that `execute` resolves with.
 */
export type RequestSuccessHandler = (response: Response) => RequestHandlerResponse;

/**
 * Generic response lifecycle builder.
 *
 * This class owns hook ordering and response replacement semantics. It does not
 * know whether the action is direct fetch, BFF-backed API work, or another
 * response-producing action.
 */
export class RequestBuilder {
  /**
   * Optional hook for observing the action immediately before execution.
   */
  private beforeSendHandler?: RequestBeforeSendHandler;

  /**
   * Optional hook for observing action completion.
   */
  private endedHandler?: RequestEndedHandler;

  /**
   * Optional hook for failed responses.
   */
  private errorHandler?: RequestErrorHandler;

  /**
   * Optional hook for successful responses.
   */
  private successHandler?: RequestSuccessHandler;

  /**
   * Creates a lifecycle builder around an injected action executor.
   *
   * Parameters:
   * - `executeAction`: runs the underlying action and reports how hooks should
   *   classify the result.
   * - `readBeforeSendContext`: reads metadata for before-send hooks.
   */
  public constructor(
    private readonly executeAction: RequestActionExecutor,
    private readonly readBeforeSendContext: RequestBeforeSendContextReader,
  ) {}

  /**
   * Executes the action and returns the final response.
   */
  public async execute(options: RequestExecuteOptions = {}): Promise<Response> {
    await this.handleBeforeSend();

    try {
      const outcome = await this.executeAction(options);

      if (!outcome.notifyHooks) {
        return outcome.response;
      }

      const response = await this.handleOutcome(outcome, options);
      await this.handleEnded({ response });

      return response;
    } catch (error) {
      await this.handleEnded({ error });
      throw error;
    }
  }

  /**
   * Registers a side-effect hook that runs immediately before action execution.
   */
  public onBeforeSend(handler: RequestBeforeSendHandler): this {
    this.beforeSendHandler = handler;
    return this;
  }

  /**
   * Registers a side-effect hook that runs after action execution.
   */
  public onEnded(handler: RequestEndedHandler): this {
    this.endedHandler = handler;
    return this;
  }

  /**
   * Registers a handler for failed responses.
   */
  public onError(handler: RequestErrorHandler): this {
    this.errorHandler = handler;
    return this;
  }

  /**
   * Registers a handler for successful responses.
   */
  public onSuccess(handler: RequestSuccessHandler): this {
    this.successHandler = handler;
    return this;
  }

  /**
   * Runs the before-send hook with action metadata.
   */
  private async handleBeforeSend(): Promise<void> {
    await this.beforeSendHandler?.(this.readBeforeSendContext());
  }

  /**
   * Runs the ended hook with final completion metadata.
   */
  private async handleEnded(context: RequestEndedContext): Promise<void> {
    await this.endedHandler?.(context);
  }

  /**
   * Resolves a failed response through the registered error handler.
   */
  private async handleError(response: Response): Promise<Response> {
    const handledResponse = await this.errorHandler?.({
      response,
      retry: () => this.execute({ skipErrorHandler: true }),
    });

    return handledResponse ?? response;
  }

  /**
   * Routes an action outcome through the appropriate response hook.
   */
  private async handleOutcome(
    outcome: RequestActionOutcome,
    options: RequestExecuteOptions,
  ): Promise<Response> {
    if (outcome.kind === "success") {
      return await this.handleSuccess(outcome.response);
    }

    if (options.skipErrorHandler === true || this.errorHandler === undefined) {
      return outcome.response;
    }

    return await this.handleError(outcome.response);
  }

  /**
   * Resolves a successful response through the registered success handler.
   */
  private async handleSuccess(response: Response): Promise<Response> {
    const handledResponse = await this.successHandler?.(response);

    return handledResponse ?? response;
  }
}
