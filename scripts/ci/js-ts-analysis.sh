#!/usr/bin/env bash
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

mapfile -t biome_files < <(find . -type f \( \
  -name '*.js' -o -name '*.jsx' -o -name '*.ts' -o -name '*.tsx' -o -name '*.json' -o -name '*.jsonc' \) \
  -not -path './node_modules/*' \
  -not -path './dist/*' \
  -not -path './build/*')

if [ "${#biome_files[@]}" -eq 0 ]; then
  echo "No JS/TS/JSON files found."
else
  npx --no-install biome ci .
fi

mapfile -t duplicate_files < <(find . -type f \( -name '*.js' -o -name '*.jsx' -o -name '*.ts' -o -name '*.tsx' \) \
  -not -path './node_modules/*' \
  -not -path './dist/*' \
  -not -path './build/*')

if [ "${#duplicate_files[@]}" -eq 0 ]; then
  echo "No JS/TS files found for duplication detection."
  exit 0
fi

npx --no-install jscpd .
