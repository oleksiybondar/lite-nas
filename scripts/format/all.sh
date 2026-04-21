#!/usr/bin/env bash
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

formatters=(
  "markdown"
  "js-ts"
  "go"
  "shell"
)

for formatter in "${formatters[@]}"; do
  printf '\n==> Formatting %s\n' "$formatter"
  "scripts/format/${formatter}.sh"
done

printf '\nFormatting complete.\n'
