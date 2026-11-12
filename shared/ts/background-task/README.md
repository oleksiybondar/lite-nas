# Shared background tasks

`@lite-nas/background-task` provides a React context for loading and polling a
service-owned long-running task, plus reusable start/cancel buttons and a
progress indicator. The package does not assume how a service starts, cancels,
or retries work. Callers pass those API operations to the context's `execute`
method and provide localized labels to the shared controls.

Each provider needs a stable `taskKey`, a task adapter with `loadLatest`,
`poll`, `getId`, and `isActive`, and localized load/refresh failure messages.
Memoize the adapter so rendering child state does not restart the polling loop.
Polling continues while a task is active, including while the child view is
temporarily hidden; set `enabled` to control latest-task loading when the view
opens. For a task with no measurable percentage, pass `null` to
`BackgroundTaskProgress` to show an indeterminate progress bar.

The provider is scoped to one task identity. Mount a separate provider for
independent tasks, and key it by the service-owned resource and task kind when
those change.
