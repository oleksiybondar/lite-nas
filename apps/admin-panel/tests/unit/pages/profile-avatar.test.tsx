import { AuthContext } from "@contexts/auth-context";
import { PreferencesProfilePage } from "@pages/PreferencesProfilePage";
import { ApiProvider } from "@providers/ApiProvider";
import { fireEvent, render, screen } from "@testing-library/react";

/** Renders the profile page with one authenticated account and the real API adapter. */
const renderProfile = (): void => {
  render(
    <ApiProvider>
      <AuthContext.Provider
        value={{
          isAuthInited: true,
          isAuthenticated: true,
          me: {
            authenticated: true,
            auth_type: "password",
            roles: [],
            scopes: [],
            user: { id: "alice", login: "alice" },
          },
          login: vi.fn(),
          logout: vi.fn(),
        }}
      >
        <PreferencesProfilePage />
      </AuthContext.Provider>
    </ApiProvider>,
  );
};

describe("admin-panel-app/FR-003 profile avatar upload", () => {
  afterEach(() => vi.unstubAllGlobals());

  test("uploads a selected PNG to the authenticated fixed profile route", async () => {
    const fetchMock = vi
      .fn()
      .mockImplementation(
        async (_url: string, init: RequestInit) =>
          new Response(null, { status: init.method === "PUT" ? 201 : 404 }),
      );
    vi.stubGlobal("fetch", fetchMock);
    renderProfile();
    const file = new File(["png bytes"], "face.png", { type: "image/png" });
    fireEvent.change(screen.getByTestId("profile-avatar-file-input"), {
      target: { files: [file] },
    });

    await screen.findByText("Avatar saved.");
    expect(
      fetchMock.mock.calls.some(
        ([url, init]) => url === "/f/profile/avatar" && init.method === "PUT",
      ),
    ).toBe(true);
  });

  test("rejects unsupported image types before upload", async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 404 }));
    vi.stubGlobal("fetch", fetchMock);
    renderProfile();
    const file = new File(["gif bytes"], "face.gif", { type: "image/gif" });
    fireEvent.change(screen.getByTestId("profile-avatar-file-input"), {
      target: { files: [file] },
    });

    await screen.findByText("Choose a PNG or JPEG image.");
    expect(fetchMock.mock.calls.some(([_url, init]) => init.method === "PUT")).toBe(false);
  });

  test("shows the File Service status when avatar upload fails", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(null, { status: 503 })));
    renderProfile();
    fireEvent.change(screen.getByTestId("profile-avatar-file-input"), {
      target: { files: [new File(["png bytes"], "face.png", { type: "image/png" })] },
    });

    expect(
      await screen.findByText("Avatar could not be saved. Avatar upload failed (503)."),
    ).toBeInTheDocument();
  });
});
