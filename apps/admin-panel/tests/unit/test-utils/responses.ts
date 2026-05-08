/**
 * Creates a minimal response with the supplied status.
 */
export const responseWithStatus = (status: number): Response => {
  return new Response(null, { status });
};

/**
 * Creates a JSON response fixture.
 */
export const responseWithJson = (status: number, body: unknown): Response => {
  return new Response(JSON.stringify(body), {
    headers: { "Content-Type": "application/json" },
    status,
  });
};
