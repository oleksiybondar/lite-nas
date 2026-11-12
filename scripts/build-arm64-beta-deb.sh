#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/common.sh"

package_version=""
run_number="0"
output_dir="$LITE_NAS_REPO_ROOT/.build/beta-arm64"

usage() {
	cat <<'MSG'
Usage: scripts/build-arm64-beta-deb.sh [options]

Builds the complete LiteNAS ARM64 package, including all frontend assets.
The build must run on an ARM64 Linux host because some components use cgo.

Options:
  --version=VERSION       Explicit Debian version. Defaults to
                          <base>~beta.<run-number>.<short-git-sha>.
  --run-number=NUMBER     Beta sequence for the default version (default: 0).
  --output-dir=PATH       Package output directory (default: .build/beta-arm64).
  -h, --help              Show this help.
MSG
}

while [ "$#" -gt 0 ]; do
	case "$1" in
	--version=*) package_version="${1#*=}" ;;
	--run-number=*) run_number="${1#*=}" ;;
	--output-dir=*) output_dir="${1#*=}" ;;
	-h | --help)
		usage
		exit 0
		;;
	*)
		log.error "Unknown option: $1"
		usage >&2
		exit 64
		;;
	esac
	shift
done

if [ "$(go env GOARCH)" != "arm64" ]; then
	log.error "This full ARM64 package build requires an ARM64 Linux host (current Go architecture: $(go env GOARCH))."
	exit 1
fi
if ! [[ "$run_number" =~ ^[0-9]+$ ]]; then
	log.error "Beta run number must be a non-negative integer: $run_number"
	exit 64
fi

cd "$LITE_NAS_REPO_ROOT"
if [ -z "$package_version" ]; then
	short_sha="$(git rev-parse --short=7 HEAD)"
	package_version="$(packageVersion.prBeta "$run_number" "$short_sha")"
fi
if [ "${output_dir#/}" = "$output_dir" ]; then
	output_dir="$LITE_NAS_REPO_ROOT/$output_dir"
fi

artifact_dir="$LITE_NAS_REPO_ROOT/.build/all/linux-arm64"
"$SCRIPT_DIR/build-all.sh" --output-dir="$artifact_dir"
"$SCRIPT_DIR/package/build-lite-nas-deb-from-source.sh" \
	--source-layout=build-all \
	--source-root="$LITE_NAS_REPO_ROOT/.build/all" \
	--package-arch=arm64 \
	--version="$package_version" \
	--output-dir="$output_dir"

log.info "Built single LiteNAS beta package: $output_dir/lite-nas_${package_version}_arm64.deb"
