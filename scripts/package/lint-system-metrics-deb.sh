#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/common.sh"

if [ "$#" -ne 1 ]; then
	log.error "Usage: scripts/package/lint-system-metrics-deb.sh <package.deb>"
	exit 64
fi

package_path="$1"

if [ ! -f "$package_path" ]; then
	log.error "Missing package file: $package_path"
	exit 1
fi

log.requireCommand "lintian" "Install lintian and retry."

log.pushTask "Running lintian on $(basename "$package_path")"
lintian --fail-on error --display-experimental --pedantic "$package_path"
log.popTask
