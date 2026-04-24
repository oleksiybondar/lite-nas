#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/common.sh"

readonly LITE_NAS_DEB_PACKAGE_BUILD_SCRIPTS=(
	"$LITE_NAS_REPO_ROOT/scripts/package/build-lite-nas-deb.sh"
	"$LITE_NAS_REPO_ROOT/scripts/package/build-system-metrics-deb.sh"
)

readonly LITE_NAS_DEB_PACKAGE_CHANGELOGS=(
	"$LITE_NAS_REPO_ROOT/packaging/debian/lite-nas/usr/share/doc/lite-nas/changelog.Debian"
	"$LITE_NAS_REPO_ROOT/packaging/debian/lite-nas-system-metrics/usr/share/doc/lite-nas-system-metrics/changelog.Debian"
)

readonly LITE_NAS_SHARED_CONSUMER_GO_MODS=(
	"$LITE_NAS_REPO_ROOT/services/system-metrics/go.mod"
	"$LITE_NAS_REPO_ROOT/services/web-gateway/go.mod"
	"$LITE_NAS_REPO_ROOT/apps/system-metrics-cli/go.mod"
)

usage() {
	cat <<'MSG'
Usage: scripts/dev/bump-versions.sh

Interactive developer helper that updates:
- default Debian package versions in package build scripts
- Debian changelog headers for known packages
- lite-nas/shared dependency versions in known service/app go.mod files
MSG
}

replace_in_file() {
	local path="$1"
	local before="$2"
	local after="$3"

	if ! grep -Fq "$before" "$path"; then
		log.error "Expected text not found in $path: $before"
		exit 1
	fi

	python3 - "$path" "$before" "$after" <<'PY'
from pathlib import Path
import sys

path = Path(sys.argv[1])
before = sys.argv[2]
after = sys.argv[3]

content = path.read_text(encoding="utf-8")
updated = content.replace(before, after)

if content == updated:
    raise SystemExit(f"replacement target not updated in {path}")

path.write_text(updated, encoding="utf-8")
PY
}

prompt_new_version() {
	local label="$1"
	local current="$2"
	local response=""

	printf '%s current %s\n' "$label:" "$current" >&2
	printf 'new version: ' >&2
	IFS= read -r response

	if [ -z "$response" ]; then
		printf '%s\n' "$current"
		return
	fi

	printf '%s\n' "$response"
}

current_deb_version() {
	sed -n "s/^package_version=\"\${LITE_NAS_PACKAGE_VERSION:-\\(.*\\)}\"$/\\1/p" \
		"$LITE_NAS_REPO_ROOT/scripts/package/build-lite-nas-deb.sh"
}

current_shared_version() {
	sed -n 's/^require lite-nas\/shared \(v[^ ]*\)$/\1/p' \
		"${LITE_NAS_SHARED_CONSUMER_GO_MODS[0]}"
}

update_deb_package_versions() {
	local old_version="$1"
	local new_version="$2"
	local path=""

	if [ "$old_version" = "$new_version" ]; then
		log.info "Debian package version unchanged: $old_version"
		return
	fi

	log.pushTask "Updating Debian package versions"

	for path in "${LITE_NAS_DEB_PACKAGE_BUILD_SCRIPTS[@]}"; do
		log.info "Updating $(realpath --relative-to="$LITE_NAS_REPO_ROOT" "$path")"
		replace_in_file \
			"$path" \
			"package_version=\"\${LITE_NAS_PACKAGE_VERSION:-$old_version}\"" \
			"package_version=\"\${LITE_NAS_PACKAGE_VERSION:-$new_version}\""
	done

	for path in "${LITE_NAS_DEB_PACKAGE_CHANGELOGS[@]}"; do
		local package_name=""
		package_name="$(basename "$(dirname "$path")")"
		log.info "Updating $(realpath --relative-to="$LITE_NAS_REPO_ROOT" "$path")"
		replace_in_file \
			"$path" \
			"${package_name} ($old_version) unstable; urgency=medium" \
			"${package_name} ($new_version) unstable; urgency=medium"
	done

	log.popTask
}

update_shared_dependency_versions() {
	local old_version="$1"
	local new_version="$2"
	local path=""

	if [ "$old_version" = "$new_version" ]; then
		log.info "lite-nas/shared dependency version unchanged: $old_version"
		return
	fi

	log.pushTask "Updating lite-nas/shared dependency versions"

	for path in "${LITE_NAS_SHARED_CONSUMER_GO_MODS[@]}"; do
		log.info "Updating $(realpath --relative-to="$LITE_NAS_REPO_ROOT" "$path")"
		replace_in_file \
			"$path" \
			"require lite-nas/shared $old_version" \
			"require lite-nas/shared $new_version"
	done

	log.popTask
}

main() {
	if [ "${1:-}" = "--help" ] || [ "${1:-}" = "-h" ]; then
		usage
		exit 0
	fi

	if [ "$#" -ne 0 ]; then
		usage >&2
		exit 64
	fi

	cd "$LITE_NAS_REPO_ROOT"

	local deb_current=""
	deb_current="$(current_deb_version)"
	if [ -z "$deb_current" ]; then
		log.error "Unable to determine current Debian package version."
		exit 1
	fi

	local deb_new=""
	deb_new="$(prompt_new_version "DEB" "$deb_current")"

	local shared_current=""
	shared_current="$(current_shared_version)"
	if [ -z "$shared_current" ]; then
		log.error "Unable to determine current lite-nas/shared dependency version."
		exit 1
	fi

	local shared_new=""
	shared_new="$(prompt_new_version "SHARED (applies to all known service/app go.mod files)" "$shared_current")"

	update_deb_package_versions "$deb_current" "$deb_new"
	update_shared_dependency_versions "$shared_current" "$shared_new"

	log.info "Debian package version: $deb_current -> $deb_new"
	log.info "lite-nas/shared dependency version: $shared_current -> $shared_new"
}

main "$@"
