# Admin Panel — Applications and User Profile Task

## Goal

Add an Applications area that opens the separately deployed File Explorer and,
when enabled, Moments. Add user profile support backed by File Service so the
user's avatar and preferred UI settings persist in the user's `.lite-nas`
directory. Preserve the existing admin-panel authentication and theme flows
while making the shared user preference available to embedded applications.

## Required behavior

- Add an Applications navigation area. File Explorer is always deployed and
  listed. Moments is shown only when its app/service is enabled in LiteNAS
  configuration. Do not infer availability from a failed browser request.
- Embed each app at its Nginx path under `/apps/`; the app may also be opened
  directly in a new tab where its deployment supports that mode.
- Keep the current admin-panel theme as an app-specific setting. Persist the
  user's preferred UI settings through File Service in the user's
  `.lite-nas` directory. Establish the shared settings document shape before
  implementation and validate it at the frontend boundary with Zod.
- Support a user avatar stored in that same per-user directory. Read and write
  it through authenticated File Service operations; never accept a user path
  from the browser as the authority for which profile directory to modify.
- Make the preferred UI settings available to embedded Moments as a fallback.
  Moments-specific settings take precedence over the shared user preference.
- Reuse the existing `lite-nas-at` browser session. Do not create a second
  admin-panel token store or copy the token into iframe messages or URLs.

## Implementation instructions

1. Inspect the existing navigation tree, app theme storage, user identity
   response, File Service upload/read/listing contracts, and frontend API
   provider before changing integration points.
2. Define the stable per-user `.lite-nas` paths and the preferred UI settings
   schema. If File Service cannot safely create the directory or write/read
   these files under the authenticated user's authority, update its task and
   requirements first; do not bypass it with direct host filesystem access.
3. Implement profile loading and saving through the existing API provider so
   refresh and unauthorized handling remain consistent. Validate payloads
   before sending them.
4. Add user-facing avatar and appearance controls using existing admin-panel
   component organization, navigation patterns, and test selectors.
5. Add the Applications navigation and embedded app routes. Keep the route
   tree separate from the navigation tree and show service availability using
   the established service configuration source.
6. Document how the shared settings are exposed to Moments without exposing
   access tokens. Ensure theme changes made in Moments are stored by Moments
   under its own preference key.

## Acceptance criteria

- File Explorer is always listed and opens at `/apps/file-explorer`.
- Moments is listed and opens at `/apps/moments` only when enabled.
- Avatar and preferred UI settings persist in the authenticated user's
  `.lite-nas` directory through File Service.
- A Moments-specific UI setting overrides the shared preferred UI setting;
  absent an override, Moments can read the shared preference.
- No access token is passed through a URL, iframe message, or browser-readable
  storage.
- Add focused app tests for navigation visibility, profile persistence/error
  handling, and settings precedence. Follow the admin-panel test conventions.
