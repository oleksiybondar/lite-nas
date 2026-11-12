import { AuthContext } from "@contexts/auth-context";
import { LanguageProvider } from "@lite-nas/language";
import { PreferencesProfilePage } from "@pages/PreferencesProfilePage";
import { ApiProvider } from "@providers/ApiProvider";
import { fireEvent, render, screen } from "@testing-library/react";

const drawImageMock = vi.fn();

/** Renders the profile page with one authenticated account and the real API adapter. */
const renderProfile = (): void => {
  render(
    <LanguageProvider>
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
      </ApiProvider>
    </LanguageProvider>,
  );
};

/** Completes the crop dialog using jsdom's mocked canvas image. */
const confirmCrop = async (): Promise<void> => {
  const cropArea = await screen.findByTestId("profile-avatar-crop-area");
  const image = cropArea.querySelector("img");
  if (!image) throw new Error("The crop preview image was not rendered.");
  fireEvent.load(image);
  fireEvent.click(screen.getByTestId("profile-avatar-crop-confirm"));
};

/** Verifies that a cropped PNG is uploaded to the fixed profile route. */
const testCropsAndUploadsAvatar = async (): Promise<void> => {
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
  await confirmCrop();
  await screen.findByText("Avatar saved.");
  expect(
    fetchMock.mock.calls.some(
      ([url, init]) => url === "/f/profile/avatar" && init.method === "PUT",
    ),
  ).toBe(true);
};

/** Verifies that unsupported image types are rejected before upload. */
const testRejectsUnsupportedAvatar = async (): Promise<void> => {
  const fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 404 }));
  vi.stubGlobal("fetch", fetchMock);
  renderProfile();
  const file = new File(["gif bytes"], "face.gif", { type: "image/gif" });
  fireEvent.change(screen.getByTestId("profile-avatar-file-input"), {
    target: { files: [file] },
  });
  await screen.findByText("Choose a valid PNG or JPEG image.");
  expect(fetchMock.mock.calls.some(([_url, init]) => init.method === "PUT")).toBe(false);
};

/** Verifies that upload failures display the File Service response. */
const testReportsAvatarUploadFailure = async (): Promise<void> => {
  vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(null, { status: 503 })));
  renderProfile();
  fireEvent.change(screen.getByTestId("profile-avatar-file-input"), {
    target: { files: [new File(["png bytes"], "face.png", { type: "image/png" })] },
  });
  await confirmCrop();
  expect(
    await screen.findByText("Avatar could not be saved. Avatar upload failed (503)."),
  ).toBeInTheDocument();
};

/** Verifies a wider source image is cropped from within its bounds and centered. */
const testUsesCenteredSourceCoordinates = async (): Promise<void> => {
  vi.stubGlobal("fetch", vi.fn().mockResolvedValue(new Response(null, { status: 201 })));
  renderProfile();
  fireEvent.change(screen.getByTestId("profile-avatar-file-input"), {
    target: { files: [new File(["png bytes"], "face.png", { type: "image/png" })] },
  });
  await confirmCrop();
  expect(drawImageMock).toHaveBeenCalledWith(
    expect.any(HTMLImageElement),
    80,
    0,
    480,
    480,
    0,
    0,
    512,
    512,
  );
};

describe("admin-panel-app/FR-003 profile avatar upload", () => {
  beforeEach(() => {
    vi.stubGlobal("URL", {
      createObjectURL: vi.fn(() => "blob:profile-avatar"),
      revokeObjectURL: vi.fn(),
    });
    drawImageMock.mockClear();
    vi.spyOn(HTMLCanvasElement.prototype, "getContext").mockReturnValue({
      drawImage: drawImageMock,
    } as unknown as CanvasRenderingContext2D);
    vi.spyOn(HTMLCanvasElement.prototype, "toBlob").mockImplementation((callback) => {
      callback(new Blob(["cropped image"], { type: "image/png" }));
    });
    Object.defineProperties(HTMLImageElement.prototype, {
      naturalWidth: { configurable: true, get: () => 640 },
      naturalHeight: { configurable: true, get: () => 480 },
    });
  });
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
    Reflect.deleteProperty(HTMLImageElement.prototype, "naturalWidth");
    Reflect.deleteProperty(HTMLImageElement.prototype, "naturalHeight");
  });

  test(
    "crops a selected PNG before uploading it to the fixed profile route",
    testCropsAndUploadsAvatar,
  );
  test("rejects unsupported image types before upload", testRejectsUnsupportedAvatar);
  test("shows the File Service status when avatar upload fails", testReportsAvatarUploadFailure);
  test("uses centered source coordinates for a wider image", testUsesCenteredSourceCoordinates);
});
