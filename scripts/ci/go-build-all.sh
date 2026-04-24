#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/logger.sh"

cd "$(git rev-parse --show-toplevel)"

mapfile -t modules < <(find . -name go.mod -not -path './vendor/*' | sort)

if [ "${#modules[@]}" -eq 0 ]; then
	log.info "No Go modules found."
	exit 0
fi

for module in "${modules[@]}"; do
	module_dir="$(dirname "$module")"
	"$SCRIPT_DIR/go-build-module.sh" "$module_dir"
done

log.info "Go build checks passed."
