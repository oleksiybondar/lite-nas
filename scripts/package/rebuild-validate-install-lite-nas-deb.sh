#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/common.sh"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../config/version.conf"

package_arch=""
package_version=""
install_recommends=1
run_as_user=()
should_bump_local_build_counter=0

usage() {
	cat <<'MSG'
Usage: scripts/package/rebuild-validate-install-lite-nas-deb.sh [options]

Options:
  --package-arch ARCH       Package architecture to build/install (amd64 or arm64). Defaults to current Go architecture.
  --version VERSION         Package version to build. Defaults to the current local alpha version from scripts/config/package-build-counter.txt.
  --no-install-recommends   Install only hard package dependencies.
  -h, --help                Show this help.
MSG
}

while [ "$#" -gt 0 ]; do
	case "$1" in
	--package-arch)
		if [ -z "${2:-}" ]; then
			log.error "Missing value for --package-arch"
			usage >&2
			exit 2
		fi
		package_arch="$2"
		shift 2
		;;
	--version)
		if [ -z "${2:-}" ]; then
			log.error "Missing value for --version"
			usage >&2
			exit 2
		fi
		package_version="$2"
		shift 2
		;;
	--no-install-recommends)
		install_recommends=0
		shift
		;;
	-h | --help)
		usage
		exit 0
		;;
	*)
		log.error "Unknown option: $1"
		usage >&2
		exit 2
		;;
	esac
done

if [ -z "$package_arch" ]; then
	package_arch="$(build.resolveTargetArch)"
fi

case "$package_arch" in
amd64 | arm64) ;;
*)
	log.error "Unsupported package architecture: $package_arch"
	exit 64
	;;
esac

if [ -z "$package_version" ]; then
	package_version="$(packageVersion.localAlpha)"
	should_bump_local_build_counter=1
fi

log.requireCommand "apt-get" "Install apt-get and retry."
log.requireCommand "dpkg" "Install dpkg and retry."

if [ "$(id -u)" -eq 0 ] && [ -n "${SUDO_USER:-}" ] && [ "$SUDO_USER" != "root" ]; then
	run_as_user=(sudo -u "$SUDO_USER" env "HOME=$(getent passwd "$SUDO_USER" | cut -d: -f6)")
fi

stop_background_package_managers() {
	log.pushTask "Stopping background package manager services"
	sudo systemctl stop packagekit unattended-upgrades >/dev/null 2>&1 || true
	sudo pkill -f unattended-upgrade-shutdown >/dev/null 2>&1 || true
	sudo pkill -f packagekitd >/dev/null 2>&1 || true
	log.popTask
}

purge_existing_lite_nas_package() {
	local status=""

	status="$(dpkg-query -W -f='${Status}' lite-nas 2>/dev/null || true)"
	if [ -z "$status" ]; then
		log.info "LiteNAS is not currently installed."
		return 0
	fi

	log.pushTask "Purging existing LiteNAS package state"
	sudo apt-get remove -y lite-nas >/dev/null 2>&1 || true
	sudo apt-get purge -y lite-nas >/dev/null 2>&1 || true
	sudo dpkg --remove --force-remove-reinstreq lite-nas >/dev/null 2>&1 || true
	sudo dpkg --purge --force-remove-reinstreq lite-nas >/dev/null 2>&1 || true
	sudo DEBIAN_FRONTEND=noninteractive dpkg --configure -a >/dev/null 2>&1 || true
	log.popTask
}

repair_local_artifact_ownership() {
	if [ "${#run_as_user[@]}" -eq 0 ]; then
		return 0
	fi

	log.pushTask "Repairing local artifact ownership for ${SUDO_USER}"
	sudo chown -R "$SUDO_USER:$SUDO_USER" \
		"$LITE_NAS_REPO_ROOT/.build" \
		"$LITE_NAS_REPO_ROOT/.cache" \
		>/dev/null 2>&1 || true
	log.popTask
}

main() {
	local package_path=""

	stop_background_package_managers
	purge_existing_lite_nas_package
	repair_local_artifact_ownership

	log.pushTask "Building local LiteNAS package ${package_version} (${package_arch})"
	"${run_as_user[@]}" "$LITE_NAS_REPO_ROOT/scripts/build-lite-nas-package.sh" \
		--version="$package_version" \
		--output-dir="$LITE_NAS_REPO_ROOT/.build/packages"
	log.popTask

	package_path="$LITE_NAS_REPO_ROOT/.build/packages/lite-nas_${package_version}_${package_arch}.deb"
	if [ ! -f "$package_path" ]; then
		log.error "Expected built package not found: $package_path"
		exit 1
	fi

	"${run_as_user[@]}" "$LITE_NAS_REPO_ROOT/scripts/ci/validate-lite-nas-deb-contents.sh" "$package_path" "$package_arch"

	log.pushTask "Installing rebuilt LiteNAS package"
	if [ "$install_recommends" -eq 0 ]; then
		sudo env DEBIAN_FRONTEND=noninteractive \
			"$LITE_NAS_REPO_ROOT/scripts/package/install-lite-nas-deb.sh" \
			--package "$package_path" \
			--no-install-recommends
	else
		sudo env DEBIAN_FRONTEND=noninteractive \
			"$LITE_NAS_REPO_ROOT/scripts/package/install-lite-nas-deb.sh" \
			--package "$package_path"
	fi
	log.popTask

	if [ "$should_bump_local_build_counter" -eq 1 ]; then
		packageVersion.bumpLocalBuildCounter
	fi

	log.info "Rebuilt, validated, and installed package: $package_path"
}

main "$@"
