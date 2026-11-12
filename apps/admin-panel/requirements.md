# Admin Panel Software Requirements

Requirement IDs use `admin-panel-app/<TYPE>-<NNN>`. These requirements cover
the Applications, dashboard, and user-preference behavior of the authenticated
admin panel.

## Functional Requirements

### admin-panel-app/FR-001 Open installed applications

The sidebar MUST expose an Applications category with File Explorer as a child.
The Applications page MUST show File Explorer as one keyboard-accessible,
whole-card link to `/applications/file-explorer`, without a separate Open
button. The admin route MUST render the separately served
`/apps/file-explorer/` Nginx mount in an iframe. The `/apps/` mount MUST remain
separate from React routes.

### admin-panel-app/FR-002 Navigate from the dashboard

The dashboard MUST present whole-card links to the available System Metrics,
Services, Applications, and Preferences pages. Each card MUST have a visible
focus state and a stable semantic selector and path. Card descriptions MUST
describe working destinations rather than future placeholders.

### admin-panel-app/FR-003 Manage the current user's avatar

The profile page MUST accept a PNG or JPEG image up to 5 MiB, validate it
before upload, and write it through `PUT /f/profile/avatar` using the existing
authenticated API provider. It MUST show the saved avatar when available and
fall back to the existing user avatar or initials when absent. Failed uploads
MUST show an error with the File Service HTTP status when one is available.

### admin-panel-app/FR-004 Persist shared and app settings

The shared `ui.json` document MUST contain a validated `preferredMode` of
`light` or `dark`. The admin-specific `admin-panel.json` document MUST contain
the theme source, mode, template, and monitoring polling settings by scope.
Authenticated reads MUST use File Service before local storage and copy valid
file values into local storage. Missing or unavailable files MUST fall back to
validated local values. Without an authenticated session, preferences MUST
remain usable through local storage. Changes MUST update local storage and,
when authenticated, write the appropriate file through File Service. A file
write failure MUST be visible while retaining the local copy.

### admin-panel-app/FR-005 Reuse settings controls

The profile and avatar panel MUST come from `@lite-nas/profile`; appearance
controls and their panel MUST come from `@lite-nas/theme`; and monitoring and
process polling forms MUST come from `@lite-nas/settings`. The admin panel
MUST supply its own authenticated transport, current values, and save actions
to these shared controls.

## Interface and Security Requirements

### admin-panel-app/IR-001 Use fixed File Service profile routes

The browser MUST use `/f/profile/ui`, `/f/profile/admin-panel`, and
`/f/profile/avatar` for profile reads and writes. It MUST NOT submit a home
path to select the profile owner. The request layer MUST validate submitted
settings and avatar files at runtime with Zod.

### admin-panel-app/SR-001 Keep authentication in the browser session

The admin panel MUST reuse the existing `lite-nas-at` cookie and API provider
refresh behavior. It MUST NOT store access tokens in local storage or pass
them through iframe messages or URLs.

## Operational and Testability Requirements

### admin-panel-app/OR-001 Support local File Service development

The Vite development server MUST forward `/f/` requests to File Service,
remove the `/f` proxy prefix, and preserve the browser Host so File Service's
same-origin mutation check can validate the request.

### admin-panel-app/TR-001 Verify navigation and persistence

Frontend tests MUST cover dashboard and Applications destinations, the iframe
mount, avatar validation and upload failures, and authenticated-file-first
settings with local-storage fallback.
