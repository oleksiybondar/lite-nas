import { defaultThemeSettings, loadThemeSettings, saveThemeSettings } from "@theme/manager/storage";

const themeSettingsStorageKey = "lite-nas.admin-panel.theme.settings";

describe("theme settings storage", () => {
  beforeEach(() => {
    window.localStorage.clear();
  });

  test("defaults to dark mode", () => {
    expect(loadThemeSettings()).toEqual(defaultThemeSettings);
    expect(defaultThemeSettings.mode).toBe("dark");
  });

  test("loads saved settings", () => {
    saveThemeSettings({
      mode: "light",
      source: "user",
      templateName: "default",
    });

    expect(loadThemeSettings()).toEqual({
      mode: "light",
      source: "user",
      templateName: "default",
    });
  });

  test("uses defaults when saved settings are not valid JSON", () => {
    window.localStorage.setItem(themeSettingsStorageKey, "{");

    expect(loadThemeSettings()).toEqual(defaultThemeSettings);
  });

  test("uses defaults when saved settings are missing fields", () => {
    window.localStorage.setItem(
      themeSettingsStorageKey,
      JSON.stringify({
        mode: "light",
        source: "user",
      }),
    );

    expect(loadThemeSettings()).toEqual(defaultThemeSettings);
  });

  test("uses defaults when saved settings contain invalid values", () => {
    window.localStorage.setItem(
      themeSettingsStorageKey,
      JSON.stringify({
        mode: "invalid",
        source: "invalid",
        templateName: "invalid",
      }),
    );

    expect(loadThemeSettings()).toEqual(defaultThemeSettings);
  });
});
