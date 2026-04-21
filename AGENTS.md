# Agent Instructions

These instructions apply to all AI agents working in this repository, including
CLI-based agents and IDE-integrated agents.

## Git and Repository Operations

Agents may:

- Read repository state with commands such as `git status`, `git diff`,
  `git log`, and `git show`.
- Create, edit, and delete files as required by an explicit user request.
- Stage files with `git add` when preparing or inspecting the agent's own
  changes.
- Create a pull request only when there is already an appropriate pushed branch
  or pushed state available.

Agents must not:

- Create commits.
- Amend commits.
- Create tags.
- Push branches, tags, or commits.
- Force-push.
- Stage unrelated user changes.
- Run destructive Git commands such as `git reset --hard`, `git checkout --`,
  `git clean`, or equivalent history/worktree rewriting commands unless the
  user explicitly requests that exact operation.
- Push a branch only for the purpose of opening a pull request.

If a pull request would require pushing local changes first, do not create the
pull request. Instead, explain that a PR can be created after the user pushes the
branch or otherwise provides a pushed source branch.

## Change Handling

- Keep edits focused on the user's request.
- Preserve user changes already present in the worktree.
- Do not revert unrelated files.
- Report validation performed and any commands that could not be run.

## Tooling Conventions

- Use existing scripts and configs where available.
- Prefer `./scripts/install-dev-dependencies.sh` for local developer tooling.
- Prefer `./scripts/run-ci.sh` for the full local static-analysis check.
- Prefer scripts under `scripts/ci/` for CI-equivalent checks.
- Prefer scripts under `scripts/format/` for manual formatting.
