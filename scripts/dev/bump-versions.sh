#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/common.sh"

readonly LITE_NAS_PACKAGE_BUILD_SCRIPTS=(
	"$LITE_NAS_REPO_ROOT/scripts/package/build-lite-nas-deb.sh"
	"$LITE_NAS_REPO_ROOT/scripts/package/build-system-metrics-deb.sh"
)

readonly LITE_NAS_PACKAGE_CHANGELOGS=(
	"$LITE_NAS_REPO_ROOT/packaging/debian/lite-nas/usr/share/doc/lite-nas/changelog.Debian"
	"$LITE_NAS_REPO_ROOT/packaging/debian/lite-nas-system-metrics/usr/share/doc/lite-nas-system-metrics/changelog.Debian"
)

usage() {
	cat <<'MSG'
Usage: scripts/dev/bump-versions.sh

Interactive developer helper that updates:
- default Debian package versions in packaging scripts
- Debian changelog headers for known packages
- the lite-nas/shared dependency version in all known service/app Go modules
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

	sed -i "s|$before|$after|g" "$path"
}

prompt_new_version() {
	local label="$1"
	local current="$2"
	local response=""

	printf '%s current %s\n' "$label:" "$current"
	printf 'new version: '
	IFS= read -r response

	if [ -z "$response" ]; then
		log.info "Keeping $label at $current"
		printf '%s\n' "$current"
		return
	fi

	printf '%s\n' "$response"
}

discover_go_modules_requiring_shared() {
	find "$LITE_NAS_REPO_ROOT" -name go.mod -not -path '*/vendor/*' -print | LC_ALL=C sort | while IFS= read -r go_mod; do
		if grep -Eq '^require lite-nas/shared v' "$go_mod"; then
			printf '%s\n' "$go_mod"
		fi
	done
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

	local current_deb_version=""
	current_deb_version="$(sed -n "s/^package_version=\"\${LITE_NAS_PACKAGE_VERSION:-\\(.*\\)}\"$/\\1/p" scripts/package/build-lite-nas-deb.sh)"
	if [ -z "$current_deb_version" ]; then
		log.error "Unable to determine current Debian package version."
		exit 1
	fi

	local new_deb_version=""
	new_deb_version="$(prompt_new_version "DEB" "$current_deb_version")"

	local current_shared_version=""
	current_shared_version="$(sed -n 's/^require lite-nas\/shared \(v[^ ]*\)$/\1/p' services/system-metrics/go.mod)"
	if [ -z "$current_shared_version" ]; then
		log.error "Unable to determine current lite-nas/shared dependency version."
		exit 1
	fi

	local new_shared_version=""
	new_shared_version="$(prompt_new_version "SHARED (applies to all service/app go.mod files)" "$current_shared_version")"

	log.pushTask "Updating Debian package versions"
	local package_script=""
	for package_script in "${LITE_NAS_PACKAGE_BUILD_SCRIPTS[@]}"; do
		replace_in_file \
			"$package_script" \
			"package_version=\"\${LITE_NAS_PACKAGE_VERSION:-$current_deb_version}\"" \
			"package_version=\"\${LITE_NAS_PACKAGE_VERSION:-$new_deb_version}\""
	done

	local changelog=""
	for changelog in "${LITE_NAS_PACKAGE_CHANGELOGS[@]}"; do
		local package_name=""
		package_name="$(basename "$(dirname "$changelog")")"
		replace_in_file \
			"$changelog" \
			"${package_name} ($current_deb_version) unstable; urgency=medium" \
			"${package_name} ($new_deb_version) unstable; urgency=medium"
	done
	log.popTask

	log.pushTask "Updating lite-nas/shared dependency versions"
	local go_mod=""
	while IFS= read -r go_mod; do
		log.info "Updating $(realpath --relative-to="$LITE_NAS_REPO_ROOT" "$go_mod")"
		replace_in_file \
			"$go_mod" \
			"require lite-nas/shared $current_shared_version" \
			"require lite-nas/shared $new_shared_version"
	done < <(discover_go_modules_requiring_shared)
	log.popTask

	log.info "Updated Debian version to $new_deb_version"
	log.info "Updated lite-nas/shared dependency version to $new_shared_version across all known service/app Go modules"
}

main "$@"
