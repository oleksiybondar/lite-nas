# Shared application language

`@lite-nas/language` bundles English (`en`), Ukrainian (`uk`), and Danish (`da`)
JSON dictionaries. English is the default. Components inside `LanguageProvider`
use `useLanguage().getLabel("section.key")` for visible text and
`LanguageSelector` for switching languages.

The provider uses `lite-nas.ui.preferred-language` in local storage before sign
in. An authenticated app passes the current user ID and its own profile
`read`/`write` transport. The saved `preferredLanguage` in `/f/profile/ui`
then takes precedence. Older UI files without this field keep the browser
choice. A selection is applied and stored locally immediately; authenticated
selections also update `ui.json` while preserving `preferredMode`. A write
failure is shown by the selector and leaves the local choice active.

Add a key to all three locale files before using it. A missing translation
falls back to the English value, then the key, so gaps remain visible during
development.
