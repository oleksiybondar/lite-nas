import { resolveThemeSettings } from "@theme/manager/resolveThemeSettings";
import type { ThemeSettings } from "@theme/types";

describe("resolveThemeSettings", () => {
  test("uses dark mode when source is default", () => {
    const settings: ThemeSettings = {
      mode: "light",
      source: "default",
      templateName: "default",
    };

    expect(resolveThemeSettings(settings, "light")).toEqual({
      mode: "dark",
      templateName: "default",
    });
  });

  test("uses the operating system mode when source is os", () => {
    const settings: ThemeSettings = {
      mode: "light",
      source: "os",
      templateName: "default",
    };

    expect(resolveThemeSettings(settings, "dark")).toEqual({
      mode: "dark",
      templateName: "default",
    });
  });

  test("uses the stored user settings when source is user", () => {
    const settings: ThemeSettings = {
      mode: "light",
      source: "user",
      templateName: "default",
    };

    expect(resolveThemeSettings(settings, "dark")).toEqual({
      mode: "light",
      templateName: "default",
    });
  });
});
