# Shared settings controls

`PreferencesPageHeader` and `PreferencesTabs` give applications the same
preferences heading and tab styling. `PreferencesTabs` accepts ordinary MUI
`Tab` children, so apps can keep either local state or route links. Its `inline`
layout fills the page width; `sidebar` places the same tabs in a responsive
settings sidebar. Applications own the tab values, destinations, and content.

`@lite-nas/settings` exports controlled monitoring history and process snapshot
polling cards. Applications supply settings, a save callback, labels, and a
scope key. The cards validate draft intervals with Zod and expose the same
stable selectors used by the admin panel. The application owns local storage,
profile file persistence, and any monitoring provider.
