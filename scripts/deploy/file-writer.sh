#!/usr/bin/env bash

if [ -n "${LITE_NAS_DEPLOY_FILE_WRITER_LOADED:-}" ]; then
	return 0
fi
readonly LITE_NAS_DEPLOY_FILE_WRITER_LOADED=1

DEPLOY_HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/../helpers/common.sh"

readonly LITE_NAS_FILE_WRITER_BINARY_TARGET="${LITE_NAS_FILE_WRITER_BINARY_TARGET:-/usr/libexec/lite-nas/lite-nas-file-writer}"
readonly LITE_NAS_FILE_WRITER_SYMLINK_TARGET="${LITE_NAS_FILE_WRITER_SYMLINK_TARGET:-/usr/bin/lite-nas-file-writer}"
readonly LITE_NAS_FILE_WRITER_PROFILE_SOURCE="${LITE_NAS_FILE_WRITER_PROFILE_SOURCE:-$LITE_NAS_REPO_ROOT/configs/etc/apparmor.d/usr.libexec.lite-nas.lite-nas-file-writer}"
readonly LITE_NAS_FILE_WRITER_PROFILE_TARGET="${LITE_NAS_FILE_WRITER_PROFILE_TARGET:-/etc/apparmor.d/usr.libexec.lite-nas.lite-nas-file-writer}"
readonly LITE_NAS_FILE_WRITER_SUDOERS_SOURCE="${LITE_NAS_FILE_WRITER_SUDOERS_SOURCE:-$LITE_NAS_REPO_ROOT/configs/etc/sudoers.d/lite-nas-file-writer}"
readonly LITE_NAS_FILE_WRITER_SUDOERS_TARGET="${LITE_NAS_FILE_WRITER_SUDOERS_TARGET:-/etc/sudoers.d/lite-nas-file-writer}"

deploy.fileWriter.requireTools() {
	log.requireCommand "install" "Install coreutils and retry."
	log.requireCommand "realpath" "Install coreutils and retry."
	log.requireCommand "chown" "Install coreutils and retry."
	log.requireCommand "chmod" "Install coreutils and retry."
	log.requireCommand "ln" "Install coreutils and retry."
	log.requireCommand "apparmor_parser" "Install AppArmor and retry."
	log.requireCommand "visudo" "Install sudo and retry."
}

deploy.fileWriter.deploy() {
	local source_binary="$1"
	local load_profile="${2:-1}"
	if [ ! -f "$source_binary" ]; then
		log.error "Missing file-writer binary: $source_binary"
		exit 1
	fi
	if [ ! -f "$LITE_NAS_FILE_WRITER_PROFILE_SOURCE" ]; then
		log.error "Missing file-writer AppArmor profile: $LITE_NAS_FILE_WRITER_PROFILE_SOURCE"
		exit 1
	fi
	if [ ! -f "$LITE_NAS_FILE_WRITER_SUDOERS_SOURCE" ]; then
		log.error "Missing file-writer sudoers rule: $LITE_NAS_FILE_WRITER_SUDOERS_SOURCE"
		exit 1
	fi
	deploy.validateSudoersDropIn "$LITE_NAS_FILE_WRITER_SUDOERS_SOURCE"
	install -D -m 0440 -o root -g root \
		"$LITE_NAS_FILE_WRITER_SUDOERS_SOURCE" "$LITE_NAS_FILE_WRITER_SUDOERS_TARGET"

	install -D -m 0644 -o root -g root \
		"$LITE_NAS_FILE_WRITER_PROFILE_SOURCE" "$LITE_NAS_FILE_WRITER_PROFILE_TARGET"
	if [ "$load_profile" = "1" ]; then
		apparmor_parser -r \
			-I "$(dirname "$LITE_NAS_FILE_WRITER_PROFILE_TARGET")" \
			-I /etc/apparmor.d \
			"$LITE_NAS_FILE_WRITER_PROFILE_TARGET"
	fi
	if [ "$(realpath "$source_binary")" = "$(realpath -m "$LITE_NAS_FILE_WRITER_BINARY_TARGET")" ]; then
		chown root:root "$LITE_NAS_FILE_WRITER_BINARY_TARGET"
		chmod 0755 "$LITE_NAS_FILE_WRITER_BINARY_TARGET"
	else
		install -D -m 0755 -o root -g root "$source_binary" "$LITE_NAS_FILE_WRITER_BINARY_TARGET"
	fi
	install -d -m 0755 "$(dirname "$LITE_NAS_FILE_WRITER_SYMLINK_TARGET")"
	ln -sfn "$LITE_NAS_FILE_WRITER_BINARY_TARGET" "$LITE_NAS_FILE_WRITER_SYMLINK_TARGET"
}
