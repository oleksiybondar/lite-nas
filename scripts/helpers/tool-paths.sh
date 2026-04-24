#!/usr/bin/env bash

if [ -n "${LITE_NAS_TOOL_PATHS_LOADED:-}" ]; then
	return 0
fi
readonly LITE_NAS_TOOL_PATHS_LOADED=1

TOOL_PATHS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$TOOL_PATHS_DIR/../.." && pwd)"
export GOBIN="${REPO_ROOT}/.bin"
GO_PATH="$(go env GOPATH 2>/dev/null || true)"
export PATH="${GOBIN}:${GO_PATH}/bin:${PATH}"
