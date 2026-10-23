# Shared settings controls

`@lite-nas/settings` exports controlled monitoring history and process snapshot
polling cards. Applications supply settings, a save callback, labels, and a
scope key. The cards validate draft intervals with Zod and expose the same
stable selectors used by the admin panel. The application owns local storage,
profile file persistence, and any monitoring provider.
