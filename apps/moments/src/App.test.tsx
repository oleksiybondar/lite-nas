// @vitest-environment jsdom
import { languageStorageKey } from "@lite-nas/language";
import { act, cleanup, fireEvent, screen, waitFor } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";
import { App } from "./App";
import {
  loadGalleryConcurrency,
  loadMe,
  loadMomentsConfig,
  login,
  logout,
  request,
  saveGalleryConcurrency,
  saveSharedMode,
} from "./api";
import type { Me } from "./schemas";
import { renderWithLanguage as render } from "./test-utils/language";

vi.mock("./api", () => ({
  sessionExpiredEvent: "lite-nas-moments-session-expired",
  loadMomentsConfig: vi.fn(),
  loadGalleryConcurrency: vi.fn(),
  loadMe: vi.fn(),
  login: vi.fn(),
  logout: vi.fn(),
  request: vi.fn(),
  saveSharedMode: vi.fn(),
  saveGalleryConcurrency: vi.fn(),
}));

const me: Me = {
  authenticated: true,
  auth_type: "password",
  user: { id: "alice", login: "alice" },
  roles: ["owner"],
  scopes: ["read"],
};

/** Gives the shell unavailable optional resources while keeping auth controlled. */
function renderMomentsApp(concurrency = 10): void {
  vi.mocked(loadMomentsConfig).mockResolvedValue("/home/oleksiybondar");
  vi.mocked(loadGalleryConcurrency).mockResolvedValue(concurrency);
  vi.mocked(saveGalleryConcurrency).mockResolvedValue(undefined);
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
  window.localStorage.removeItem("moments-image-concurrency");
  window.localStorage.removeItem(languageStorageKey);
  window.history.replaceState({}, "", "/");
});

test.each([
  "/galleries/00000000-0000-4000-8000-000000000001",
  "/apps/moments/galleries/00000000-0000-4000-8000-000000000001",
])("a signed-in user opening %s stays on the gallery route", async (path) => {
  window.history.replaceState({}, "", path);
  vi.mocked(loadMe).mockResolvedValue(me);
  renderMomentsApp();
  expect(await screen.findByText("Could not open this gallery.")).toBeTruthy();
});

test("My collections navigation updates the browser URL", async () => {
  vi.mocked(loadMe).mockResolvedValue(me);
  renderMomentsApp();
  const button = await screen.findByRole("button", { name: "My collections" });
  await act(async () => {
    fireEvent.click(button);
  });
  expect(window.location.pathname).toBe("/mine");
});

// Requirements: moments/FR-001. Signing in replaces the public form with the gallery shell.
test("successful sign in opens the authenticated Moments home", async () => {
  vi.mocked(loadMe).mockResolvedValueOnce(null).mockResolvedValueOnce(me);
  vi.mocked(login).mockResolvedValue("/");
  renderMomentsApp();
  await submitLogin();
  expect(await screen.findByTestId("moments-home")).toBeTruthy();
});

test("an anonymous gallery link keeps its destination on the login URL", async () => {
  const galleryPath = "/galleries/00000000-0000-4000-8000-000000000001";
  window.history.replaceState({}, "", galleryPath);
  vi.mocked(loadMe).mockResolvedValue(null);
  renderMomentsApp();
  await screen.findByTestId("moments-login");
  expect(window.location.search).toBe(`?redirect_to=${encodeURIComponent(galleryPath)}`);
});

test("the login page uses the language selected before sign-in", async () => {
  window.localStorage.setItem(languageStorageKey, "da");
  vi.mocked(loadMe).mockResolvedValue(null);
  renderMomentsApp();
  expect(await screen.findByRole("heading", { name: "Log ind på Moments" })).toBeTruthy();
});

test("sign in returns to the gallery from a reloadable login URL", async () => {
  const galleryPath = "/galleries/00000000-0000-4000-8000-000000000001";
  window.history.replaceState({}, "", `/login?redirect_to=${encodeURIComponent(galleryPath)}`);
  vi.mocked(loadMe).mockResolvedValueOnce(null).mockResolvedValueOnce(me);
  vi.mocked(login).mockResolvedValue(galleryPath);
  renderMomentsApp();
  await submitLogin();
  await screen.findByText("Could not open this gallery.");
  expect(window.location.pathname).toBe(galleryPath);
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

test("settings sidebar opens gallery preferences", async () => {
  vi.mocked(loadMe).mockResolvedValue(me);
  renderMomentsApp();
  await chooseUserAction("user-menu-application-settings-link");
  fireEvent.click(screen.getByRole("tab", { name: "Gallery" }));
  expect(screen.getByLabelText("Maximum concurrent gallery image loads")).toBeTruthy();
});

// Requirements: moments-backend/FR-010. Users can maintain their own saved-login list.
test("header friends entry opens the private friends page", async () => {
  vi.mocked(loadMe).mockResolvedValue(me);
  renderMomentsApp();
  fireEvent.click(await screen.findByTestId("moments-friends-button"));
  expect(screen.getByTestId("moments-friends-page")).toBeTruthy();
  expect(
    await screen.findByText(
      "Save logins in your private contact list. Missing accounts and unavailable avatars look the same.",
    ),
  ).toBeTruthy();
});

test("Moments title returns from settings to the home feed", async () => {
  vi.mocked(loadMe).mockResolvedValue(me);
  renderMomentsApp();
  await chooseUserAction("user-menu-application-settings-link");
  fireEvent.click(screen.getByRole("link", { name: "Moments" }));
  expect(await screen.findByTestId("moments-home")).toBeTruthy();
});

test("gallery settings save image concurrency to the user profile", async () => {
  vi.mocked(loadMe).mockResolvedValue(me);
  renderMomentsApp();
  await chooseUserAction("user-menu-application-settings-link");
  fireEvent.click(screen.getByRole("tab", { name: "Gallery" }));
  fireEvent.change(screen.getByLabelText("Maximum concurrent gallery image loads"), {
    target: { value: "12" },
  });
  await waitFor(() => expect(saveGalleryConcurrency).toHaveBeenCalledWith(12));
});

test("gallery settings load the user's saved concurrency", async () => {
  vi.mocked(loadMe).mockResolvedValue(me);
  renderMomentsApp(12);
  await chooseUserAction("user-menu-application-settings-link");
  fireEvent.click(screen.getByRole("tab", { name: "Gallery" }));
  await waitFor(() =>
    expect(screen.getByLabelText("Maximum concurrent gallery image loads")).toHaveProperty(
      "value",
      "12",
    ),
  );
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

test("an expired session returns Moments to sign in", async () => {
  vi.mocked(loadMe).mockResolvedValue(me);
  renderMomentsApp();
  await screen.findByTestId("moments-home");
  fireEvent(window, new Event("lite-nas-moments-session-expired"));
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
