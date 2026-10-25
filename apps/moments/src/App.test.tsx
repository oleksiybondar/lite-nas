// @vitest-environment jsdom
import { act, cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";
import { App } from "./App";
import { loadMe, login, logout, request, saveSharedMode } from "./api";
import type { Me } from "./schemas";

vi.mock("./api", () => ({
  loadMe: vi.fn(),
  login: vi.fn(),
  logout: vi.fn(),
  request: vi.fn(),
  saveSharedMode: vi.fn(),
}));

const me: Me = {
  authenticated: true,
  auth_type: "password",
  user: { id: "alice", login: "alice" },
  roles: ["owner"],
  scopes: ["read"],
};

/** Gives the shell unavailable optional resources while keeping auth controlled. */
function renderMomentsApp(): void {
  vi.stubGlobal("IS_REACT_ACT_ENVIRONMENT", true);
  vi.stubGlobal(
    "fetch",
    vi.fn(async () => new Response("", { status: 404 })),
  );
  vi.mocked(request).mockResolvedValue(new Response("", { status: 404 }));
  render(<App />);
}

/** Submits the visible credential form through the browser login action. */
async function submitLogin(): Promise<void> {
  await screen.findByText("Sign in to Moments");
  fireEvent.change(screen.getByLabelText(/Login/i), { target: { value: "alice" } });
  fireEvent.change(screen.getByLabelText(/Password/i), { target: { value: "secret" } });
  await act(async () => {
    fireEvent.click(screen.getByRole("button", { name: "Sign in" }));
  });
}

/** Chooses one authenticated navigation action through the shared user menu. */
async function chooseUserAction(testId: string): Promise<void> {
  fireEvent.click(await screen.findByRole("button", { name: "User menu" }));
  await act(async () => {
    fireEvent.click(screen.getByTestId(testId));
  });
}

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  vi.clearAllMocks();
});

// Requirements: moments/FR-001. Signing in replaces the public form with the gallery shell.
test("successful sign in opens the authenticated Moments home", async () => {
  vi.mocked(loadMe).mockResolvedValueOnce(null).mockResolvedValueOnce(me);
  vi.mocked(login).mockResolvedValue(undefined);
  renderMomentsApp();
  await submitLogin();
  expect(await screen.findByTestId("moments-home")).toBeTruthy();
});

// Requirements: moments/FR-001. Failed credentials leave the user on the form.
test("failed sign in shows a visible error", async () => {
  vi.mocked(loadMe).mockResolvedValue(null);
  vi.mocked(login).mockRejectedValue(new Error("invalid credentials"));
  renderMomentsApp();
  await submitLogin();
  expect(await screen.findByText("Sign in failed.")).toBeTruthy();
});

// Requirements: moments/FR-002. Authenticated users can reach application settings.
test("user menu opens Moments appearance settings", async () => {
  vi.mocked(loadMe).mockResolvedValue(me);
  renderMomentsApp();
  await chooseUserAction("user-menu-application-settings-link");
  await waitFor(() => expect(screen.getByTestId("preferences-theme-settings-title")).toBeTruthy());
});

// Requirements: moments/FR-002. Authenticated users can inspect their profile.
test("user menu opens the shared profile screen", async () => {
  vi.mocked(loadMe).mockResolvedValue(me);
  renderMomentsApp();
  await chooseUserAction("user-menu-profile-link");
  expect(screen.getByTestId("preferences-profile-page")).toBeTruthy();
});

// Requirements: moments/FR-001. A successful logout returns to the public sign-in form.
test("logging out closes the authenticated shell", async () => {
  vi.mocked(loadMe).mockResolvedValue(me);
  vi.mocked(logout).mockResolvedValue(undefined);
  renderMomentsApp();
  await chooseUserAction("user-menu-logout-button");
  expect(await screen.findByTestId("moments-login")).toBeTruthy();
});

// Requirements: moments/FR-001. A failed logout keeps the user informed.
test("failed logout reports an error", async () => {
  vi.mocked(loadMe).mockResolvedValue(me);
  vi.mocked(logout).mockRejectedValue(new Error("logout unavailable"));
  renderMomentsApp();
  await chooseUserAction("user-menu-logout-button");
  await chooseUserAction("user-menu-application-settings-link");
  expect(await screen.findByText("Sign out failed.")).toBeTruthy();
});

// Requirements: moments/FR-002. Applying a light appearance saves the user preference.
test("appearance changes are saved to the shared user profile", async () => {
  vi.mocked(loadMe).mockResolvedValue(me);
  vi.mocked(saveSharedMode).mockResolvedValue(undefined);
  renderMomentsApp();
  await chooseUserAction("user-menu-application-settings-link");
  fireEvent.click(screen.getByRole("switch", { name: "Dark" }));
  await act(async () => {
    fireEvent.click(screen.getByRole("button", { name: "Apply" }));
  });
  await waitFor(() => expect(saveSharedMode).toHaveBeenCalledWith("light"));
});

// Requirements: moments/FR-002. Failed preference writes leave a visible warning.
test("appearance save failure reports its local fallback", async () => {
  vi.mocked(loadMe).mockResolvedValue(me);
  vi.mocked(saveSharedMode).mockRejectedValue(new Error("File Service unavailable"));
  renderMomentsApp();
  await chooseUserAction("user-menu-application-settings-link");
  fireEvent.click(screen.getByRole("switch", { name: "Dark" }));
  await act(async () => {
    fireEvent.click(screen.getByRole("button", { name: "Apply" }));
  });
  expect(
    await screen.findByText("Saved in this browser, but the user file could not be updated."),
  ).toBeTruthy();
});
