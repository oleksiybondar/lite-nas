// @vitest-environment jsdom
import { cleanup, fireEvent, screen, waitFor } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";
import { FriendsPage } from "./FriendsPage";
import { listFriendsPage, loadFriendAvatar, saveFriend } from "./moments-data";
import { renderWithLanguage } from "./test-utils/language";

const friendFixture = {
  username: "ari-example",
  firstName: "Ari",
  lastName: "Example",
  nickname: "",
  createdAt: "2026-10-26T00:00:00Z",
  updatedAt: "2026-10-26T00:00:00Z",
};

vi.mock("./moments-data", () => ({
  listFriendsPage: vi.fn(),
  loadFriendAvatar: vi.fn(),
  removeFriend: vi.fn(),
  saveFriend: vi.fn(),
}));

afterEach(() => {
  cleanup();
  vi.clearAllMocks();
});

// Requirements: moments-backend/FR-010. The modal saves user-entered login labels.
test("add friend modal saves a login with first name, last name, and nickname", async () => {
  vi.mocked(listFriendsPage).mockResolvedValue({
    items: [],
    metadata: { page: 1, size: 12, total_count: 0, total_pages: 0 },
  });
  vi.mocked(saveFriend).mockResolvedValue(undefined);
  renderWithLanguage(<FriendsPage />);
  fireEvent.click(screen.getByRole("button", { name: "Add friend" }));
  fireEvent.change(screen.getByRole("textbox", { name: "Username" }), {
    target: { value: "unknown-login" },
  });
  fireEvent.change(screen.getByRole("textbox", { name: "First name" }), {
    target: { value: "Ari" },
  });
  fireEvent.change(screen.getByRole("textbox", { name: "Last name" }), {
    target: { value: "Example" },
  });
  fireEvent.change(screen.getByRole("textbox", { name: "Nickname" }), {
    target: { value: "Ace" },
  });
  fireEvent.click(screen.getByRole("button", { name: "Add friend" }));
  await waitFor(() =>
    expect(saveFriend).toHaveBeenCalledWith({
      username: "unknown-login",
      firstName: "Ari",
      lastName: "Example",
      nickname: "Ace",
    }),
  );
});

// Requirements: moments-backend/FR-010. Each saved login triggers an eligible avatar attempt.
test("saved friend avatar is requested through the fixed File Service path helper", async () => {
  vi.mocked(listFriendsPage).mockResolvedValue({
    items: [{ ...friendFixture, username: "not-a-real-account", nickname: "Imaginary" }],
    metadata: { page: 1, size: 12, total_count: 1, total_pages: 1 },
  });
  vi.mocked(loadFriendAvatar).mockResolvedValue(null);
  renderWithLanguage(<FriendsPage />);
  await waitFor(() => expect(loadFriendAvatar).toHaveBeenCalledWith("not-a-real-account"));
});

test("friends search sends the entered term to the paged list request", async () => {
  vi.mocked(listFriendsPage).mockResolvedValue({
    items: [friendFixture],
    metadata: { page: 1, size: 12, total_count: 1, total_pages: 1 },
  });
  renderWithLanguage(<FriendsPage />);
  fireEvent.change(screen.getByRole("textbox", { name: "Search friends" }), {
    target: { value: "ari" },
  });
  await waitFor(() => expect(listFriendsPage).toHaveBeenCalledWith(1, 12, "ari"));
});

test("friends pagination requests the selected server page", async () => {
  vi.mocked(listFriendsPage).mockResolvedValue({
    items: [friendFixture],
    metadata: { page: 1, size: 12, total_count: 13, total_pages: 2 },
  });
  renderWithLanguage(<FriendsPage />);
  fireEvent.click(await screen.findByRole("button", { name: "Go to page 2" }));
  await waitFor(() => expect(listFriendsPage).toHaveBeenCalledWith(2, 12, ""));
});
