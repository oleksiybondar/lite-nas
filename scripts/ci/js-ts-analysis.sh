#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/helpers/logger.sh
source "$SCRIPT_DIR/../helpers/logger.sh"

cd "$(git rev-parse --show-toplevel)"

mapfile -t biome_files < <(find . -type f \( \
  -name '*.js' -o -name '*.jsx' -o -name '*.ts' -o -name '*.tsx' -o -name '*.json' -o -name '*.jsonc' \) \
  -not -path './node_modules/*' \
  -not -path './dist/*' \
  -not -path './build/*')

if [ "${#biome_files[@]}" -eq 0 ]; then
  log.info "No JS/TS/JSON files found."
else
  log.pushTask "Running Biome"
  npx --no-install biome ci .
  log.popTask
fi

mapfile -t duplicate_files < <(find . -type f \( -name '*.js' -o -name '*.jsx' -o -name '*.ts' -o -name '*.tsx' \) \
  -not -path './node_modules/*' \
  -not -path './dist/*' \
  -not -path './build/*')

if [ "${#duplicate_files[@]}" -eq 0 ]; then
  log.info "No JS/TS files found for duplication detection."
  exit 0
fi

log.pushTask "Running JS/TS duplication detection"
npx --no-install jscpd .
log.popTask
