#!/usr/bin/env bash
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

mapfile -t files < <(find . -type f \( \
  -name '*.js' -o -name '*.jsx' -o -name '*.ts' -o -name '*.tsx' -o -name '*.json' -o -name '*.jsonc' \) \
  -not -path './node_modules/*' \
  -not -path './dist/*' \
  -not -path './build/*')

if [ "${#files[@]}" -eq 0 ]; then
  echo "No JS/TS/JSON files found."
  exit 0
fi

npx --no-install biome check --write .
