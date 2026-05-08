#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

common_args=()

for arg in "$@"; do
	case "$arg" in
	--version=* | --output-dir=*) ;;
	*)
		printf 'Unknown option: %s\n' "$arg" >&2
		exit 64
		;;
	esac
done

common_args=("$@")

"$SCRIPT_DIR/build-lite-nas-deb.sh" "${common_args[@]}"
