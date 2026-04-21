#!/usr/bin/env bash
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

checks=(
  "markdown-analysis"
  "shell-analysis"
  "js-ts-analysis"
  "go-analysis"
)

for check in "${checks[@]}"; do
  printf '\n==> Running %s\n' "$check"
  "scripts/ci/${check}.sh"
done

printf '\nAll local CI checks passed.\n'
