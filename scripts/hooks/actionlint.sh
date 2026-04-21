#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/tool-paths.sh"

if ! command -v actionlint >/dev/null 2>&1; then
	printf 'Missing required command: actionlint\n' >&2
	printf 'Run ./scripts/install-dev-dependencies.sh to install developer tooling.\n' >&2
	exit 127
fi

actionlint "$@"
