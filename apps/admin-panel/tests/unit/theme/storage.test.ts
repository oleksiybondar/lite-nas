import { defaultThemeSettings, loadThemeSettings, saveThemeSettings } from "@theme/manager/storage";

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
});
