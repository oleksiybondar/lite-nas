# LiteNAS theme

`@lite-nas/theme` owns the MUI palette, typography, component defaults, theme
registry, and browser preference format shared by the admin panel and File
Explorer. The admin panel stores user choices under
`lite-nas.admin-panel.theme.settings`. Both apps read that key from the same
origin. The default source resolves to the dark theme; the OS source follows
`prefers-color-scheme`; the user source uses the saved mode and template.

Apps render the result of `createAppTheme(templateName, mode)` inside MUI's
`ThemeProvider` and `CssBaseline`. App-specific presentation belongs in the
app, while palette or component defaults used by both apps belong here.
