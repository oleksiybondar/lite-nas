#!/usr/bin/env bash

if [ -n "${LITE_NAS_REPO_FILES_LOADED:-}" ]; then
	return 0
fi
readonly LITE_NAS_REPO_FILES_LOADED=1

if [ -z "${LITE_NAS_REPO_ROOT:-}" ]; then
	REPO_FILES_SOURCE="${BASH_SOURCE[0]}"
	if command -v realpath >/dev/null 2>&1; then
		REPO_FILES_SOURCE="$(realpath "$REPO_FILES_SOURCE")"
	fi

	REPO_FILES_DIR="$(cd "$(dirname "$REPO_FILES_SOURCE")" && pwd)"
	LITE_NAS_REPO_ROOT="$(cd "$REPO_FILES_DIR/../.." && pwd)"
	export LITE_NAS_REPO_ROOT
fi

repo.files.collectShellFiles() {
	# shellcheck disable=SC2178,SC2034
	local -n files_ref="$1"

	# shellcheck disable=SC2034
	mapfile -t files_ref < <(find "$LITE_NAS_REPO_ROOT" -type f \( -name '*.sh' -o -name '*.bash' -o -name '*.zsh' \) \
		-not -path "$LITE_NAS_REPO_ROOT/.git/*" \
		-not -path '*/node_modules/*' \
		-not -path "$LITE_NAS_REPO_ROOT/.venv/*" \
		-not -path "$LITE_NAS_REPO_ROOT/dist/*" \
		-not -path "$LITE_NAS_REPO_ROOT/build/*")
}

repo.files.collectMarkdownFiles() {
	# shellcheck disable=SC2178,SC2034
	local -n files_ref="$1"

	# shellcheck disable=SC2034
	mapfile -t files_ref < <(find "$LITE_NAS_REPO_ROOT" -type f -name '*.md' \
		-not -path "$LITE_NAS_REPO_ROOT/.git/*" \
		-not -path '*/node_modules/*' \
		-not -path "$LITE_NAS_REPO_ROOT/.venv/*" \
		-not -path "$LITE_NAS_REPO_ROOT/logs/*" \
		-not -path "$LITE_NAS_REPO_ROOT/tests/logs/*" \
		-not -path "$LITE_NAS_REPO_ROOT/dist/*" \
		-not -path "$LITE_NAS_REPO_ROOT/build/*")
}
