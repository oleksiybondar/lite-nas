#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/common.sh"

cd "$LITE_NAS_REPO_ROOT"

log.requireCommand "npm" "Install Node.js/npm and retry."

with_coverage=0

for option in "$@"; do
	case "$option" in
	--with-coverage)
		with_coverage=1
		;;
	*)
		log.error "Unknown option for test-admin-panel.sh: $option"
		;;
	esac
done

log.pushTask "Testing admin-panel frontend"
if [ "$with_coverage" -eq 1 ]; then
	npm --prefix "$LITE_NAS_REPO_ROOT/apps/admin-panel" run test:unit:coverage
else
	npm --prefix "$LITE_NAS_REPO_ROOT/apps/admin-panel" run test:unit
fi
log.popTask
