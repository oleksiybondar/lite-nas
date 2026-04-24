#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/tool-paths.sh"

if [ "$#" -lt 1 ]; then
	exit 0
fi

formatter="$1"
shift

existing_files=()
for file in "$@"; do
	if [ -f "$file" ]; then
		existing_files+=("$file")
	fi
done

if [ "${#existing_files[@]}" -eq 0 ]; then
	exit 0
fi

case "$formatter" in
gofumpt)
	gofumpt -w "${existing_files[@]}"
	;;
goimports)
	goimports -local lite-nas -w "${existing_files[@]}"
	;;
*)
	printf 'Unsupported formatter: %s\n' "$formatter" >&2
	exit 1
	;;
esac
