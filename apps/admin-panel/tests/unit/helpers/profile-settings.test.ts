import type { ApiContextValue } from "@dto/api/api";
import { loadProfileSettings, saveSharedMode } from "@helpers/profile-settings";
import {
  avatarFileSchema,
  loadLocalSharedUi,
  readProfileDocument,
  sharedUiSchema,
} from "@lite-nas/profile";
import { loadThemeSettings } from "@theme/manager/storage";

/** Wraps a controlled response in the API action shape used by profile helpers. */
const getApi = (response: (url: string) => Response): Pick<ApiContextValue, "get"> => ({
  get: (url) => ({ execute: async () => response(url) }) as ReturnType<ApiContextValue["get"]>,
});

describe("admin-panel-app/FR-004 profile file precedence", () => {
  beforeEach(() => window.localStorage.clear());

  test("replicates authenticated preferred mode to local storage", async () => {
    await loadProfileSettings(
      getApi((url) =>
        url.endsWith("/ui")
          ? Response.json({ preferredMode: "light" })
          : new Response(null, { status: 404 }),
      ),
    );

    expect(loadLocalSharedUi()).toEqual({ preferredMode: "light" });
  });

  test("replicates authenticated admin theme to local storage", async () => {
    await loadProfileSettings(
      getApi((url) =>
        url.endsWith("/admin-panel")
          ? Response.json({
              theme: { mode: "light", source: "user", templateName: "default" },
              monitoring: {},
            })
          : new Response(null, { status: 404 }),
      ),
    );

    expect(loadThemeSettings().mode).toBe("light");
  });

  test("keeps the local preference when the file is unavailable", async () => {
    window.localStorage.setItem(
      "lite-nas.ui.preferred-mode",
      JSON.stringify({ preferredMode: "dark" }),
    );
    await loadProfileSettings(getApi(() => new Response(null, { status: 503 })));

    expect(loadLocalSharedUi()?.preferredMode).toBe("dark");
  });
});

describe("admin-panel-app/FR-004 profile file validation and saves", () => {
  beforeEach(() => window.localStorage.clear());

  test("rejects malformed settings returned by File Service", async () => {
    await expect(
      readProfileDocument(
        async () => Response.json({ preferredMode: "purple" }),
        "ui",
        sharedUiSchema,
      ),
    ).rejects.toThrow();
  });

  test("validates and writes the preferred mode before caching it", async () => {
    let submitted: unknown;
    const api: Pick<ApiContextValue, "put"> = {
      put: (_url, body) => {
        submitted = body;
        return { execute: async () => new Response(null, { status: 201 }) } as ReturnType<
          ApiContextValue["put"]
        >;
      },
    };
    await saveSharedMode(api, "light");

    expect(submitted).toEqual({ document: { preferredMode: "light" } });
  });

  test("keeps the local preferred mode when the profile write fails", async () => {
    const api: Pick<ApiContextValue, "put"> = {
      put: () =>
        ({ execute: async () => new Response(null, { status: 503 }) }) as ReturnType<
          ApiContextValue["put"]
        >,
    };
    await saveSharedMode(api, "light").catch(() => undefined);

    expect(loadLocalSharedUi()?.preferredMode).toBe("light");
  });

  test("accepts JPEG files whose browser MIME type is missing", () => {
    const image = new File(["jpeg bytes"], "portrait.jpg");

    expect(avatarFileSchema.safeParse(image).success).toBe(true);
  });
});
