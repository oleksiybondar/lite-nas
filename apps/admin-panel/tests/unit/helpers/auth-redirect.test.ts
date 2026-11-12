import { authDestination, loginDestination, loginRedirectPath } from "@lite-nas/profile";

test.each([
  { value: "/system?tab=network#daily", expected: "/system?tab=network#daily" },
  { value: "https://evil.example/path", expected: "/" },
  { value: "//evil.example/path", expected: "/" },
  { value: "/\\evil.example/path", expected: "/" },
  { value: "/%5Cevil.example/path", expected: "/" },
  { value: "/%2Fevil.example/path", expected: "/" },
  { value: "/login?redirect_to=%2Fsystem", expected: "/" },
  { value: "/login/", expected: "/" },
  { value: "/system\n/alerts", expected: "/" },
])("auth return target $value resolves to $expected", ({ value, expected }) => {
  expect(authDestination(value)).toBe(expected);
});

test("login link encodes the protected route", () => {
  expect(loginRedirectPath("/system?tab=network#daily")).toBe(
    "/login?redirect_to=%2Fsystem%3Ftab%3Dnetwork%23daily",
  );
});

test("query return target survives a login page reload", () => {
  expect(loginDestination("?redirect_to=%2Fsystem%3Ftab%3Dnetwork", "/old")).toBe(
    "/system?tab=network",
  );
});
