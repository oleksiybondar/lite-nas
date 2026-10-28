import {
  getLabel,
  LanguageProvider,
  LanguageSelector,
  type LanguageTransport,
  languageStorageKey,
  loadLocalLanguage,
  useLanguage,
} from "@lite-nas/language";
// @vitest-environment jsdom
import { act, cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import type { ReactElement } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import danish from "../../../shared/ts/language/locales/da.json";
import english from "../../../shared/ts/language/locales/en.json";
import ukrainian from "../../../shared/ts/language/locales/uk.json";

/** Lists leaf paths so locale files can be checked for the same label keys. */
const labelPaths = (value: Record<string, unknown>, prefix = ""): string[] =>
  Object.entries(value).flatMap(([key, entry]) => {
    const path = prefix ? `${prefix}.${key}` : key;
    return typeof entry === "string" ? [path] : labelPaths(entry as Record<string, unknown>, path);
  });

/** Exposes the language action without relying on Material UI's select internals. */
const LanguageProbe = (): ReactElement => {
  const { language, setLanguage } = useLanguage();
  return (
    <button type="button" onClick={() => setLanguage("uk")}>
      {language}
    </button>
  );
};

afterEach(() => {
  cleanup();
  window.localStorage.clear();
});

describe("browser language", () => {
  it("falls back to English when the browser value is invalid", () => {
    window.localStorage.setItem(languageStorageKey, "unsupported");
    expect(loadLocalLanguage()).toBe("en");
  });

  it("traverses a locale dictionary by dotted key", () => {
    expect(getLabel("uk", "common.signIn")).toBe("Увійти");
  });

  it("changes the anonymous browser language through the visible selector", () => {
    render(
      <LanguageProvider>
        <LanguageSelector />
      </LanguageProvider>,
    );
    fireEvent.mouseDown(screen.getByRole("combobox", { name: "Language" }));
    fireEvent.click(
      within(screen.getByRole("listbox")).getByRole("option", { name: "Українська" }),
    );
    expect(window.localStorage.getItem(languageStorageKey)).toBe("uk");
  });
});

describe("language flags", () => {
  it.each([
    ["en", "🇬🇧"],
    ["uk", "🇺🇦"],
    ["da", "🇩🇰"],
  ])("shows the %s flag in the selected value", (language, flag) => {
    window.localStorage.setItem(languageStorageKey, language);
    render(
      <LanguageProvider>
        <LanguageSelector />
      </LanguageProvider>,
    );
    expect(screen.getByRole("combobox").textContent).toContain(flag);
  });
});

describe("browser tab synchronization", () => {
  it("applies a language change from another tab", () => {
    render(
      <LanguageProvider>
        <LanguageProbe />
      </LanguageProvider>,
    );
    window.localStorage.setItem(languageStorageKey, "da");
    fireEvent(window, new StorageEvent("storage", { key: languageStorageKey }));
    expect(screen.getByRole("button").textContent).toBe("da");
  });
});

describe("locale dictionaries", () => {
  it.each([
    ["Ukrainian", ukrainian],
    ["Danish", danish],
  ])("keeps every English label key in %s", (_name, dictionary) => {
    expect(labelPaths(dictionary).sort()).toEqual(labelPaths(english).sort());
  });
});

describe("saved language precedence", () => {
  it("uses a signed-in user's saved language before the browser choice", async () => {
    window.localStorage.setItem(languageStorageKey, "da");
    const transport: LanguageTransport = {
      read: vi.fn(async () => Response.json({ preferredMode: "dark", preferredLanguage: "uk" })),
      write: vi.fn(async () => new Response(null, { status: 201 })),
    };
    render(
      <LanguageProvider userId="alice" transport={transport}>
        <LanguageProbe />
      </LanguageProvider>,
    );
    await waitFor(() => expect(screen.getByRole("button").textContent).toBe("uk"));
  });

  it("keeps the login language when the user's UI file is missing", async () => {
    window.localStorage.setItem(languageStorageKey, "da");
    const transport: LanguageTransport = {
      read: vi.fn(async () => new Response(null, { status: 404 })),
      write: vi.fn(async () => new Response(null, { status: 201 })),
    };
    await act(async () => {
      render(
        <LanguageProvider userId="alice" transport={transport}>
          <LanguageProbe />
        </LanguageProvider>,
      );
    });
    expect(screen.getByRole("button").textContent).toBe("da");
  });
});

describe("profile language validation", () => {
  it("keeps the browser language when the user file contains an unsupported locale", async () => {
    window.localStorage.setItem(languageStorageKey, "da");
    const transport: LanguageTransport = {
      read: vi.fn(async () => Response.json({ preferredMode: "dark", preferredLanguage: "pl" })),
      write: vi.fn(async () => new Response(null, { status: 201 })),
    };
    await act(async () => {
      render(
        <LanguageProvider userId="alice" transport={transport}>
          <LanguageProbe />
        </LanguageProvider>,
      );
    });
    expect(screen.getByRole("button").textContent).toBe("da");
  });
});

describe("authenticated language changes", () => {
  it("keeps the existing theme mode when saving a language change", async () => {
    let saved: unknown;
    const transport: LanguageTransport = {
      read: vi.fn(async () => Response.json({ preferredMode: "light" })),
      write: vi.fn(async (_url, body) => {
        saved = body;
        return new Response(null, { status: 201 });
      }),
    };
    render(
      <LanguageProvider userId="alice" transport={transport}>
        <LanguageProbe />
      </LanguageProvider>,
    );
    fireEvent.click(screen.getByRole("button"));
    await waitFor(() =>
      expect(saved).toEqual({ document: { preferredMode: "light", preferredLanguage: "uk" } }),
    );
  });

  it("shows a profile write failure after language selection", async () => {
    const transport: LanguageTransport = {
      read: vi.fn(async () => Response.json({ preferredMode: "dark" })),
      write: vi.fn(async () => new Response(null, { status: 503 })),
    };
    render(
      <LanguageProvider userId="alice" transport={transport}>
        <LanguageSelector />
      </LanguageProvider>,
    );
    fireEvent.mouseDown(screen.getByRole("combobox", { name: "Language" }));
    fireEvent.click(
      within(screen.getByRole("listbox")).getByRole("option", { name: "Українська" }),
    );
    expect(
      await screen.findByText(
        "Збережено в цьому браузері, але не вдалося оновити файл користувача.",
      ),
    ).toBeTruthy();
  });
});
