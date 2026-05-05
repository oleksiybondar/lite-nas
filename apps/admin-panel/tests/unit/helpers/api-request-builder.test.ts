import { createApiRequestBuilder } from "@helpers/api-request-builder";
import {
  FetchRequestBuilder,
  type FetchRequestExecutionResult,
} from "@helpers/fetch-request-builder";
import type { RequestBeforeSendContext, RequestExecuteOptions } from "@helpers/request-builder";

describe("ApiRequestBuilder configuration", () => {
  test("delegates request configuration to the wrapped fetch builder", () => {
    const fetchBuilder = newStubFetchRequestBuilder(responseWithStatus(200));

    createApiRequestBuilder(fetchBuilder)
      .method("POST")
      .credentials("include")
      .header("X-Test", "1")
      .payload({ name: "media" })
      .body({ name: "backup" });

    expect(fetchBuilder.calls).toEqual([
      "method:POST",
      "credentials:include",
      "header:X-Test:1",
      'payload:{"name":"media"}',
      'payload:{"name":"backup"}',
    ]);
  });
});

describe("ApiRequestBuilder success lifecycle", () => {
  test("runs API lifecycle hooks around the wrapped fetch execution", async () => {
    const response = responseWithStatus(200);
    const handledResponse = responseWithStatus(201);
    const fetchBuilder = newStubFetchRequestBuilder(response);
    const beforeSend = vi.fn();
    const ended = vi.fn();

    const result = await createApiRequestBuilder(fetchBuilder)
      .onBeforeSend(beforeSend)
      .onSuccess((nextResponse) => {
        expect(nextResponse).toBe(response);
        return handledResponse;
      })
      .onEnded(ended)
      .execute();

    expect(result).toBe(handledResponse);
    expect(fetchBuilder.executeOptions).toEqual({});
    expect(beforeSend).toHaveBeenCalledWith(fetchBuilder.beforeSendContext);
    expect(ended).toHaveBeenCalledWith({ response: handledResponse });
  });
});

describe("ApiRequestBuilder error lifecycle", () => {
  test("runs API error hooks for notifiable failed fetch results", async () => {
    const response = responseWithStatus(500);
    const handledResponse = responseWithStatus(502);
    const fetchBuilder = newStubFetchRequestBuilder(response);
    const onEnded = vi.fn();

    const result = await createApiRequestBuilder(fetchBuilder)
      .onError(({ response: nextResponse }) => {
        expect(nextResponse).toBe(response);
        return handledResponse;
      })
      .onEnded(onEnded)
      .execute();

    expect(result).toBe(handledResponse);
    expect(onEnded).toHaveBeenCalledWith({ response: handledResponse });
  });
});

describe("ApiRequestBuilder skipped lifecycle", () => {
  test("skips API hooks for non-notifiable fetch results", async () => {
    const response = responseWithStatus(401);
    const fetchBuilder = newStubFetchRequestBuilder(response, false);
    const onError = vi.fn();
    const onEnded = vi.fn();

    const result = await createApiRequestBuilder(fetchBuilder)
      .onError(onError)
      .onEnded(onEnded)
      .execute();

    expect(result).toBe(response);
    expect(onError).not.toHaveBeenCalled();
    expect(onEnded).not.toHaveBeenCalled();
  });
});

type StubFetchRequestBuilder = FetchRequestBuilder & {
  beforeSendContext: RequestBeforeSendContext;
  calls: string[];
  executeOptions: RequestExecuteOptions | null;
};

const newStubFetchRequestBuilder = (
  response: Response,
  notifyHooks = true,
): StubFetchRequestBuilder => {
  const beforeSendContext = {
    headers: new Headers({ Accept: "application/json" }),
    method: "GET",
    url: "/api/test",
  };
  const calls: string[] = [];

  const stub = new FetchRequestBuilder("/api/test") as StubFetchRequestBuilder;
  stub.beforeSendContext = beforeSendContext;
  stub.calls = calls;
  stub.executeOptions = null;

  vi.spyOn(stub, "credentials").mockImplementation((credentials: RequestCredentials) => {
    calls.push(`credentials:${credentials}`);
    return stub;
  });
  vi.spyOn(stub, "execute").mockResolvedValue(response);
  vi.spyOn(stub, "executeForWrapper").mockImplementation(
    async (options: RequestExecuteOptions = {}): Promise<FetchRequestExecutionResult> => {
      stub.executeOptions = options;
      return { notifyHooks, response };
    },
  );
  vi.spyOn(stub, "header").mockImplementation((name: string, value: string) => {
    calls.push(`header:${name}:${value}`);
    return stub;
  });
  vi.spyOn(stub, "method").mockImplementation((method: string) => {
    calls.push(`method:${method}`);
    return stub;
  });
  vi.spyOn(stub, "payload").mockImplementation((payload: unknown) => {
    calls.push(`payload:${JSON.stringify(payload)}`);
    return stub;
  });
  vi.spyOn(stub, "body").mockImplementation((payload: unknown) => {
    calls.push(`payload:${JSON.stringify(payload)}`);
    return stub;
  });
  vi.spyOn(stub, "readBeforeSendContext").mockReturnValue(beforeSendContext);

  return stub;
};

const responseWithStatus = (status: number): Response => {
  return new Response(null, { status });
};
