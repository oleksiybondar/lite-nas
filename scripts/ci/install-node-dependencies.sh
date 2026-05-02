#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/logger.sh"

cd "$(git rev-parse --show-toplevel)"

install_node_dependencies() {
	local package_dir="$1"

	if [ -f "$package_dir/package-lock.json" ]; then
		npm --prefix "$package_dir" ci
		return 0
	fi

	npm --prefix "$package_dir" install
}

log.pushTask "Installing root CI Node dependencies"
install_node_dependencies .
log.popTask

log.pushTask "Installing admin-panel CI Node dependencies"
install_node_dependencies apps/admin-panel
log.popTask
