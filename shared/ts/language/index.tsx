import {
  loadLocalSharedUi,
  readProfileDocument,
  saveLocalSharedUi,
  sharedUiSchema,
  writeProfileDocument,
} from "@lite-nas/profile";
import Box from "@mui/material/Box";
import FormControl from "@mui/material/FormControl";
import InputLabel from "@mui/material/InputLabel";
import MenuItem from "@mui/material/MenuItem";
import Select from "@mui/material/Select";
import Typography from "@mui/material/Typography";
import {
  createContext,
  type PropsWithChildren,
  type ReactElement,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { z } from "zod";
import danish from "./locales/da.json";
import english from "./locales/en.json";
import ukrainian from "./locales/uk.json";

/** Supported application languages, with English as the browser fallback. */
export const languageSchema = z.enum(["en", "uk", "da"]);
export type Language = z.infer<typeof languageSchema>;

/** Browser key shared by all LiteNAS applications on one origin. */
export const languageStorageKey = "lite-nas.ui.preferred-language";

/** Static dictionaries bundled into each application at build time. */
const dictionaries: Record<Language, unknown> = { en: english, uk: ukrainian, da: danish };

/** Resolves a dotted key, falling back to English and then the key itself. */
export const getLabel = (language: Language, key: string): string =>
  lookupLabel(dictionaries[language], key) ?? lookupLabel(english, key) ?? key;

/** Traverses one static dictionary without treating object prototypes as labels. */
const lookupLabel = (dictionary: unknown, key: string): string | undefined => {
  let value = dictionary;
  for (const segment of key.split(".")) {
    if (typeof value !== "object" || value === null || !Object.hasOwn(value, segment))
      return undefined;
    value = (value as Record<string, unknown>)[segment];
  }
  return typeof value === "string" ? value : undefined;
};

/** Returns a validated browser choice or English when no choice is stored. */
export const loadLocalLanguage = (): Language => {
  if (typeof window === "undefined") return "en";
  const parsed = languageSchema.safeParse(window.localStorage.getItem(languageStorageKey));
  return parsed.success ? parsed.data : "en";
};

/** Authenticated profile transport supplied by each app's existing API client. */
export interface LanguageTransport {
  read: (url: string) => Promise<Response>;
  write: (url: string, body: { document: z.infer<typeof sharedUiSchema> }) => Promise<Response>;
}

/** Language state and label lookup available to application components. */
export interface LanguageContextValue {
  language: Language;
  setLanguage: (value: Language) => void;
  getLabel: (key: string) => string;
  error: string | null;
}

const LanguageContext = createContext<LanguageContextValue | null>(null);

/** Shares language selection and syncs signed-in changes to the existing UI profile file. */
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: Keep local state and its profile synchronization together.
export const LanguageProvider = ({
  children,
  userId,
  transport,
}: PropsWithChildren<{ userId?: string | null; transport?: LanguageTransport }>): ReactElement => {
  const [language, setLanguageState] = useState<Language>(loadLocalLanguage);
  const [error, setError] = useState<string | null>(null);
  const selectionVersion = useRef(0);
  const saveQueue = useRef(Promise.resolve());
  const currentUserId = useRef(userId);
  currentUserId.current = userId;

  useEffect(() => {
    if (!userId || !transport) return;
    let active = true;
    const version = selectionVersion.current;
    void readProfileDocument(transport.read, "ui", sharedUiSchema)
      .then((document) => {
        if (!active || version !== selectionVersion.current || !document?.preferredLanguage) return;
        window.localStorage.setItem(languageStorageKey, document.preferredLanguage);
        setLanguageState(document.preferredLanguage);
      })
      .catch(() => undefined);
    return () => {
      active = false;
    };
  }, [userId, transport]);

  useEffect(() => {
    document.documentElement.lang = language;
  }, [language]);

  useEffect(() => {
    const onStorage = (event: StorageEvent): void => {
      if (event.key === languageStorageKey) setLanguageState(loadLocalLanguage());
    };
    window.addEventListener("storage", onStorage);
    return () => window.removeEventListener("storage", onStorage);
  }, []);

  const setLanguage = useCallback(
    (value: Language): void => {
      const validated = languageSchema.parse(value);
      selectionVersion.current += 1;
      const version = selectionVersion.current;
      window.localStorage.setItem(languageStorageKey, validated);
      setLanguageState(validated);
      setError(null);
      if (!userId || !transport) return;
      // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: Keep profile read, merge, write, and visible failure in one ordered transaction.
      const pending = saveQueue.current.then(async () => {
        if (currentUserId.current !== userId) return;
        try {
          const remote = await readProfileDocument(transport.read, "ui", sharedUiSchema);
          if (currentUserId.current !== userId) return;
          const document = sharedUiSchema.parse({
            preferredMode: remote?.preferredMode ?? loadLocalSharedUi()?.preferredMode ?? "dark",
            preferredLanguage: validated,
          });
          await writeProfileDocument(transport.write, "ui", sharedUiSchema, document);
          saveLocalSharedUi(document);
        } catch {
          if (selectionVersion.current === version)
            setError(getLabel(validated, "language.saveFailed"));
        }
      });
      saveQueue.current = pending;
    },
    [userId, transport],
  );

  const value = useMemo<LanguageContextValue>(
    () => ({
      language,
      setLanguage,
      getLabel: (key) => getLabel(language, key),
      error,
    }),
    [language, setLanguage, error],
  );
  return <LanguageContext.Provider value={value}>{children}</LanguageContext.Provider>;
};

/** Reads the active language context for labels and selection. */
export const useLanguage = (): LanguageContextValue => {
  const value = useContext(LanguageContext);
  if (!value) throw new Error("LanguageProvider is missing.");
  return value;
};

/** Compact selector shared by public and authenticated application headers. */
export const LanguageSelector = (): ReactElement => {
  const { language, setLanguage, getLabel, error } = useLanguage();
  return (
    <FormControl size="small" sx={{ minWidth: 130 }}>
      <InputLabel id="lite-nas-language-label">{getLabel("language.label")}</InputLabel>
      <Select
        data-testid="language-selector"
        labelId="lite-nas-language-label"
        label={getLabel("language.label")}
        value={language}
        onChange={(event) => {
          const parsed = languageSchema.safeParse(event.target.value);
          if (parsed.success) setLanguage(parsed.data);
        }}
      >
        <MenuItem value="en">
          <Box component="span" aria-hidden="true" sx={{ mr: 1 }}>
            🇬🇧
          </Box>
          {getLabel("language.english")}
        </MenuItem>
        <MenuItem value="uk">
          <Box component="span" aria-hidden="true" sx={{ mr: 1 }}>
            🇺🇦
          </Box>
          {getLabel("language.ukrainian")}
        </MenuItem>
        <MenuItem value="da">
          <Box component="span" aria-hidden="true" sx={{ mr: 1 }}>
            🇩🇰
          </Box>
          {getLabel("language.danish")}
        </MenuItem>
      </Select>
      {error && (
        <Typography color="error" variant="caption">
          {error}
        </Typography>
      )}
    </FormControl>
  );
};
