#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/common.sh"

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

bump_counter=0
has_version=0
for arg in "${common_args[@]}"; do
	if [[ "$arg" == --version=* ]]; then
		has_version=1
		break
	fi
done
if [ "$has_version" -eq 0 ]; then
	if [ -n "${LITE_NAS_PACKAGE_VERSION:-}" ]; then
		common_args+=("--version=$LITE_NAS_PACKAGE_VERSION")
	else
		common_args+=("--version=$(packageVersion.localAlpha)")
		bump_counter=1
	fi
fi

"$SCRIPT_DIR/build-lite-nas-deb.sh" "${common_args[@]}"

if [ "$bump_counter" -eq 1 ]; then
	packageVersion.bumpLocalBuildCounter
fi
