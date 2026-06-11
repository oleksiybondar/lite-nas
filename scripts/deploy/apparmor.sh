#!/usr/bin/env bash

if [ -n "${LITE_NAS_DEPLOY_APPARMOR_LOADED:-}" ]; then
	return 0
fi
readonly LITE_NAS_DEPLOY_APPARMOR_LOADED=1

DEPLOY_HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/../helpers/common.sh"

readonly LITE_NAS_APPARMOR_SOURCE_DIR="${LITE_NAS_APPARMOR_SOURCE_DIR:-$LITE_NAS_REPO_ROOT/configs/etc/apparmor.d}"
readonly LITE_NAS_APPARMOR_TARGET_DIR="${LITE_NAS_APPARMOR_TARGET_DIR:-/etc/apparmor.d}"

deploy.apparmor.requireTools() {
	local tool
	local tools=(apparmor_parser cp find install mktemp rm systemctl)
	for tool in "${tools[@]}"; do
		log.requireCommand "$tool" "Install the required AppArmor tooling and retry."
	done
}

deploy.apparmor.installConfig() {
	if [ ! -d "$LITE_NAS_APPARMOR_SOURCE_DIR" ]; then
		log.error "Missing AppArmor config source directory: $LITE_NAS_APPARMOR_SOURCE_DIR"
		exit 1
	fi

	install -d -m 0755 "$LITE_NAS_APPARMOR_TARGET_DIR"
	cp -a "$LITE_NAS_APPARMOR_SOURCE_DIR/." "$LITE_NAS_APPARMOR_TARGET_DIR/"
}

deploy.apparmor.eachManagedProfile() {
	find "$LITE_NAS_APPARMOR_TARGET_DIR" -maxdepth 1 -type f \
		\( -name 'usr.libexec.lite-nas.*' -o -name 'usr.lib.postfix.sbin' -o -name 'usr.sbin.nginx' \) \
		-print0
}

deploy.apparmor.validateConfig() {
	local profile_path
	local parser_cache_dir

	parser_cache_dir="$(mktemp -d)"

	while IFS= read -r -d '' profile_path; do
		apparmor_parser -Q -W -L "$parser_cache_dir" \
			-I "$LITE_NAS_APPARMOR_TARGET_DIR" \
			-I /etc/apparmor.d \
			"$profile_path"
	done < <(deploy.apparmor.eachManagedProfile)

	rm -rf "$parser_cache_dir"
}

deploy.apparmor.reloadConfig() {
	local profile_path

	while IFS= read -r -d '' profile_path; do
		apparmor_parser -r \
			-I "$LITE_NAS_APPARMOR_TARGET_DIR" \
			-I /etc/apparmor.d \
			"$profile_path"
	done < <(deploy.apparmor.eachManagedProfile)

	if command -v systemctl >/dev/null 2>&1 && systemctl list-unit-files apparmor.service >/dev/null 2>&1; then
		systemctl reload apparmor.service || systemctl restart apparmor.service || true
	fi
}

deploy.apparmor.deploy() {
	local should_reload="${1:-1}"

	deploy.apparmor.installConfig
	deploy.apparmor.validateConfig

	if [ "$should_reload" = "1" ]; then
		deploy.apparmor.reloadConfig
		return 0
	fi

	log.info "Skipping AppArmor reload."
}
