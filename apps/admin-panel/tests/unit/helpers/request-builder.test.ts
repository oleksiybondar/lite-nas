import { RequestBuilder, type RequestEndedContext } from "@helpers/request-builder";

const beforeSendContext = {
  headers: new Headers({ Accept: "application/json" }),
  method: "GET",
  url: "/api/test",
};

describe("RequestBuilder success lifecycle", () => {
  test("runs success lifecycle hooks and returns the handled response", async () => {
    const response = responseWithStatus(200);
    const handledResponse = responseWithStatus(201);
    const events: string[] = [];
    const endedContexts: RequestEndedContext[] = [];

    const result = await new RequestBuilder(
      async () => {
        events.push("execute");
        return { kind: "success", notifyHooks: true, response };
      },
      () => beforeSendContext,
    )
      .onBeforeSend(({ method, url }) => {
        events.push(`before:${method}:${url}`);
      })
      .onSuccess((nextResponse) => {
        events.push(`success:${nextResponse.status}`);
        return handledResponse;
      })
      .onEnded((context) => {
        events.push("ended");
        endedContexts.push(context);
      })
      .execute();

    expect(result).toBe(handledResponse);
    expect(events).toEqual(["before:GET:/api/test", "execute", "success:200", "ended"]);
    expect(endedContexts).toEqual([{ response: handledResponse }]);
  });
});

describe("RequestBuilder error lifecycle", () => {
  test("runs error hooks and supports retry replacement", async () => {
    const failedResponse = responseWithStatus(500);
    const retryResponse = responseWithStatus(200);
    let attempts = 0;
    const nextOutcome = (skipErrorHandler: boolean) => {
      attempts += 1;
      if (skipErrorHandler) {
        return successOutcome(retryResponse);
      }

      return errorOutcome(failedResponse);
    };

    const result = await new RequestBuilder(
      async ({ skipErrorHandler = false }) => nextOutcome(skipErrorHandler),
      () => beforeSendContext,
    )
      .onError(async ({ retry }) => {
        return await retry();
      })
      .execute();

    expect(result).toBe(retryResponse);
    expect(attempts).toBe(2);
  });
});

describe("RequestBuilder completion lifecycle", () => {
  test("skips hooks for non-notifiable outcomes", async () => {
    const response = responseWithStatus(401);
    const onEnded = vi.fn();

    const result = await new RequestBuilder(
      async () => {
        return { kind: "error", notifyHooks: false, response };
      },
      () => beforeSendContext,
    )
      .onEnded(onEnded)
      .execute();

    expect(result).toBe(response);
    expect(onEnded).not.toHaveBeenCalled();
  });

  test("runs ended hook with thrown execution errors", async () => {
    const error = new Error("network failed");
    const endedContexts: RequestEndedContext[] = [];

    await expect(
      new RequestBuilder(
        async () => {
          throw error;
        },
        () => beforeSendContext,
      )
        .onEnded((context) => {
          endedContexts.push(context);
        })
        .execute(),
    ).rejects.toThrow(error);

    expect(endedContexts).toEqual([{ error }]);
  });
});

const responseWithStatus = (status: number): Response => {
  return new Response(null, { status });
};

const errorOutcome = (response: Response) => {
  return { kind: "error", notifyHooks: true, response } as const;
};

const successOutcome = (response: Response) => {
  return { kind: "success", notifyHooks: true, response } as const;
};
