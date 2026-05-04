# Admin Panel Agent Instructions

These instructions apply to work under `apps/admin-panel/`.

They supplement the repository-level `AGENTS.md`. When these instructions
conflict with the repository-level file, these local instructions take
precedence.

## React Context Usage

- Prefer destructuring values returned by context hooks at the call site, for
  example `const { get } = useApi();` or `const { login } = useAuth();`.
- Avoid storing a whole context object only to call `context.method(...)` when
  the required members are known locally.
- Keep provider values focused on the stable contract exposed by the context,
  while implementation helpers can stay private to the provider module.
