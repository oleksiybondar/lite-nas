#!/usr/bin/env bash

if [ -n "${LITE_NAS_COMMON_LOADED:-}" ]; then
	return 0
fi
readonly LITE_NAS_COMMON_LOADED=1

COMMON_SOURCE="${BASH_SOURCE[0]}"
if command -v realpath >/dev/null 2>&1; then
	COMMON_SOURCE="$(realpath "$COMMON_SOURCE")"
fi

COMMON_DIR="$(cd "$(dirname "$COMMON_SOURCE")" && pwd)"
export LITE_NAS_REPO_ROOT="${LITE_NAS_REPO_ROOT:-$(cd "$COMMON_DIR/../.." && pwd)}"

# shellcheck disable=SC1091
source "$COMMON_DIR/logger.sh"
# shellcheck disable=SC1091
source "$COMMON_DIR/sudo-guard.sh"
# shellcheck disable=SC1091
source "$COMMON_DIR/tool-paths.sh"
# shellcheck disable=SC1091
source "$COMMON_DIR/go-modules.sh"
