#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/helpers/logger.sh
source "$SCRIPT_DIR/../helpers/logger.sh"

cd "$(git rev-parse --show-toplevel)"

mapfile -t files < <(find . -type f \( -name '*.sh' -o -name '*.bash' -o -name '*.zsh' \) \
  -not -path './node_modules/*' \
  -not -path './dist/*' \
  -not -path './build/*')

if [ "${#files[@]}" -eq 0 ]; then
  log.info "No shell files found."
  exit 0
fi

log.pushTask "Running shfmt autofix"
shfmt -w "${files[@]}"
log.popTask
