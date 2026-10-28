# LiteNAS browser profile contract

File Service owns `/home/<verified Unix user>/.lite-nas`. Browsers never submit
the home path for profile writes. Authenticated `GET` and `PUT` operations at
`/f/profile/ui`, `/f/profile/admin-panel`, `/f/profile/moments`, and
`/f/profile/avatar` address the current account's `ui.json`,
`admin-panel.json`, `moments.json`, and `avatar.image`.

`ui.json` contains `preferredMode` (`light` or `dark`) and may contain
`preferredLanguage` (`en`, `uk`, or `da`). Older files without a language remain
valid. The admin
panel keeps its theme source/template and per-resource monitoring timing values
in `admin-panel.json`. Moments keeps its gallery preferences under the
`gallery` group in `moments.json`. The reusable Zod schema and file-read/write
helpers are exported by `@lite-nas/profile`. `loadSharedUiPreference` reads the file first,
copies a valid response to local storage, and falls back to the browser copy.
`ProfilePreferencesPanel` supplies the reusable profile and avatar controls.
Applications pass their own authenticated request function and upload callback;
the admin panel uses its API provider. Neither this package
nor embedded application messages carry tokens.

The package also exports avatar upload validation and authenticated avatar
read/write helpers, so Moments can reuse the same fixed file contract later.

On an authenticated load, a valid profile file wins and is copied to local
storage. If the file is absent or unavailable, the app uses its validated local
storage value, then its default. File Explorer reads `ui.json` directly using
the shared browser cookie and the same origin. Moments reads its app-specific
preferences from `moments.json`, then its browser fallback. Shared appearance
preferences remain in `ui.json`.

Appearance controls and the complete appearance preferences panel are exported
by `@lite-nas/theme`. Monitoring and process polling settings cards are exported
by `@lite-nas/settings`. They accept current values and change callbacks, so
each application can choose its own provider and profile file.

`avatar.image` accepts complete PNG or JPEG files up to 5 MiB. File Service
publishes it as `0644` and serves authenticated cross-user reads through the
fixed `/f/home/<username>/.lite-nas/avatar.image` path. Other files keep normal
per-user read authorization.
