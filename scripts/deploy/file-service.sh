#!/usr/bin/env bash

if [ -n "${LITE_NAS_DEPLOY_FILE_SERVICE_LOADED:-}" ]; then
	return 0
fi
readonly LITE_NAS_DEPLOY_FILE_SERVICE_LOADED=1

DEPLOY_HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/../helpers/common.sh"

readonly LITE_NAS_FILE_SERVICE_USER="${LITE_NAS_FILE_SERVICE_USER:-lite-nas-file-service}"
readonly LITE_NAS_FILE_SERVICE_READ_GROUP="${LITE_NAS_FILE_SERVICE_READ_GROUP:-lite-nas-fs-rd}"
readonly LITE_NAS_FILE_SERVICE_BINARY_TARGET="${LITE_NAS_FILE_SERVICE_BINARY_TARGET:-/usr/libexec/lite-nas/file-service}"
readonly LITE_NAS_FILE_SERVICE_CONFIG_SOURCE="${LITE_NAS_FILE_SERVICE_CONFIG_SOURCE:-$LITE_NAS_REPO_ROOT/configs/etc/lite-nas/file-service.conf}"
readonly LITE_NAS_FILE_SERVICE_UNIT_SOURCE="${LITE_NAS_FILE_SERVICE_UNIT_SOURCE:-$LITE_NAS_REPO_ROOT/configs/etc/systemd/system/lite-nas-file-service.service}"
readonly LITE_NAS_FILE_SERVICE_PROFILE_SOURCE="${LITE_NAS_FILE_SERVICE_PROFILE_SOURCE:-$LITE_NAS_REPO_ROOT/configs/etc/apparmor.d/usr.libexec.lite-nas.file-service}"
readonly LITE_NAS_FILE_SERVICE_SUDOERS_SOURCE="${LITE_NAS_FILE_SERVICE_SUDOERS_SOURCE:-$LITE_NAS_REPO_ROOT/configs/etc/sudoers.d/lite-nas-file-service-commands}"

deploy.fileService.requireTools() {
	local tool
	for tool in getent groupadd id useradd usermod install chown chmod visudo apparmor_parser find setfacl; do
		log.requireCommand "$tool" "Install the required system tooling and retry."
	done
}

# Existing home content needs an access ACL; the group alone cannot traverse it.
# Default ACLs keep newly created home content readable until its owner changes
# the effective ACL mask (for example with chmod 0600); the sudo reader covers
# that case when the service can still traverse the parent directories.
deploy.fileService.provisionReadACL() {
	local marker=/var/lib/lite-nas/file-service/read-acl-provisioned
	if [ -f "$marker" ]; then
		return 0
	fi
	if [ ! -d /home ]; then
		return 0
	fi
	find /home -type d -exec setfacl -m \
		"g:$LITE_NAS_FILE_SERVICE_READ_GROUP:rx,d:g:$LITE_NAS_FILE_SERVICE_READ_GROUP:rx" {} +
	find /home -type f -exec setfacl -m "g:$LITE_NAS_FILE_SERVICE_READ_GROUP:r" {} +
	install -m 0600 -o root -g root /dev/null "$marker"
}

deploy.fileService.ensureUser() {
	deploy.ensureSystemGroup lite-nas "LiteNAS configuration group"
	deploy.ensureSystemGroup "$LITE_NAS_FILE_SERVICE_USER" "file service group"
	deploy.ensureSystemGroup "$LITE_NAS_FILE_SERVICE_READ_GROUP" "storage read group"
	if ! id "$LITE_NAS_FILE_SERVICE_USER" >/dev/null 2>&1; then
		useradd --system --no-create-home --home-dir /nonexistent \
			--shell /usr/sbin/nologin --gid "$LITE_NAS_FILE_SERVICE_USER" \
			--groups "lite-nas,$LITE_NAS_FILE_SERVICE_READ_GROUP" \
			"$LITE_NAS_FILE_SERVICE_USER"
	else
		usermod --append --groups "lite-nas,$LITE_NAS_FILE_SERVICE_READ_GROUP" \
			"$LITE_NAS_FILE_SERVICE_USER"
	fi
}

deploy.fileService.enableAndStart() {
	deploy.enableAndRefreshService lite-nas-file-service.service
}

deploy.fileService.deploy() {
	local source_binary="$1"
	local should_start="${2:-1}"
	if [ ! -f "$source_binary" ]; then
		log.error "Missing file-service binary: $source_binary"
		exit 1
	fi
	deploy.fileService.ensureUser
	install -d -m 0700 -o "$LITE_NAS_FILE_SERVICE_USER" -g "$LITE_NAS_FILE_SERVICE_USER" \
		/var/lib/lite-nas/file-service/uploads
	deploy.fileService.provisionReadACL
	"$LITE_NAS_REPO_ROOT/scripts/rotate-nats-certificates.sh" --if-missing --user "$LITE_NAS_FILE_SERVICE_USER"
	if [ "$(realpath -- "$source_binary")" = "$(realpath -m -- "$LITE_NAS_FILE_SERVICE_BINARY_TARGET")" ]; then
		chown root:root "$LITE_NAS_FILE_SERVICE_BINARY_TARGET"
		chmod 0755 "$LITE_NAS_FILE_SERVICE_BINARY_TARGET"
	else
		install -D -m 0755 -o root -g root "$source_binary" "$LITE_NAS_FILE_SERVICE_BINARY_TARGET"
	fi
	deploy.installMergedConfig "$LITE_NAS_FILE_SERVICE_CONFIG_SOURCE" /etc/lite-nas/file-service.conf \
		"$LITE_NAS_FILE_SERVICE_USER" lite-nas
	install -D -m 0644 -o root -g root "$LITE_NAS_FILE_SERVICE_UNIT_SOURCE" \
		/etc/systemd/system/lite-nas-file-service.service
	install -d -m 0751 -o root -g lite-nas /var/log/lite-nas
	if [ ! -e /var/log/lite-nas/file-service.log ]; then
		install -m 0640 -o "$LITE_NAS_FILE_SERVICE_USER" -g "$LITE_NAS_FILE_SERVICE_USER" \
			/dev/null /var/log/lite-nas/file-service.log
	fi
	chown "$LITE_NAS_FILE_SERVICE_USER:$LITE_NAS_FILE_SERVICE_USER" /var/log/lite-nas/file-service.log
	chmod 0640 /var/log/lite-nas/file-service.log
	deploy.validateSudoersDropIn "$LITE_NAS_FILE_SERVICE_SUDOERS_SOURCE"
	install -D -m 0440 -o root -g root "$LITE_NAS_FILE_SERVICE_SUDOERS_SOURCE" \
		/etc/sudoers.d/lite-nas-file-service-commands
	install -D -m 0644 -o root -g root "$LITE_NAS_FILE_SERVICE_PROFILE_SOURCE" \
		/etc/apparmor.d/usr.libexec.lite-nas.file-service
	apparmor_parser -r -I /etc/apparmor.d /etc/apparmor.d/usr.libexec.lite-nas.file-service
	if [ "$should_start" = "1" ]; then
		deploy.fileService.enableAndStart
	fi
}
