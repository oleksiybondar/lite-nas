import { loginRequestSchema } from "@schemas/auth/login";

describe("loginRequestSchema", () => {
  test("accepts local login names", () => {
    expect(
      loginRequestSchema.parse({
        login: "john.doe",
        password: "1234",
      }),
    ).toEqual({
      login: "john.doe",
      password: "1234",
    });
  });

  test("accepts email login names", () => {
    expect(
      loginRequestSchema.parse({
        login: "john@example.com",
        password: "1234",
      }),
    ).toEqual({
      login: "john@example.com",
      password: "1234",
    });
  });

  test("rejects invalid login payloads", () => {
    expect(() =>
      loginRequestSchema.parse({
        login: "john doe",
        password: "123",
      }),
    ).toThrow();
  });
});
