#!/usr/bin/env bash

if [ -n "${LITE_NAS_DEPLOY_STREAM_SERVICE_LOADED:-}" ]; then
	return 0
fi
readonly LITE_NAS_DEPLOY_STREAM_SERVICE_LOADED=1

DEPLOY_HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/../helpers/common.sh"

readonly LITE_NAS_STREAM_SERVICE_USER="${LITE_NAS_STREAM_SERVICE_USER:-lite-nas-stream-service}"
readonly LITE_NAS_STREAM_SERVICE_READ_GROUP="${LITE_NAS_STREAM_SERVICE_READ_GROUP:-lite-nas-fs-rd}"
readonly LITE_NAS_STREAM_SERVICE_BINARY_TARGET="${LITE_NAS_STREAM_SERVICE_BINARY_TARGET:-/usr/libexec/lite-nas/stream-service}"
readonly LITE_NAS_STREAM_SERVICE_CONFIG_SOURCE="${LITE_NAS_STREAM_SERVICE_CONFIG_SOURCE:-$LITE_NAS_REPO_ROOT/configs/etc/lite-nas/stream-service.conf}"
readonly LITE_NAS_STREAM_SERVICE_UNIT_SOURCE="${LITE_NAS_STREAM_SERVICE_UNIT_SOURCE:-$LITE_NAS_REPO_ROOT/configs/etc/systemd/system/lite-nas-stream-service.service}"
readonly LITE_NAS_STREAM_SERVICE_PROFILE_SOURCE="${LITE_NAS_STREAM_SERVICE_PROFILE_SOURCE:-$LITE_NAS_REPO_ROOT/configs/etc/apparmor.d/usr.libexec.lite-nas.stream-service}"
readonly LITE_NAS_STREAM_SERVICE_SUDOERS_SOURCE="${LITE_NAS_STREAM_SERVICE_SUDOERS_SOURCE:-$LITE_NAS_REPO_ROOT/configs/etc/sudoers.d/lite-nas-stream-service-read}"

deploy.streamService.requireTools() {
	local tool
	for tool in getent groupadd id useradd usermod install chown chmod apparmor_parser find findmnt ffmpeg heif-dec visudo; do
		log.requireCommand "$tool" "Install the required system tooling and retry."
	done
	local encoders
	encoders="$(ffmpeg -hide_banner -encoders 2>/dev/null)"
	if [[ "$encoders" != *libx264* ]]; then
		log.error "The installed ffmpeg lacks the libx264 encoder."
		return 1
	fi
	if [[ "$(heif-dec --list-decoders 2>/dev/null)" != *libde265* ]]; then
		log.error "The installed libheif lacks the HEVC decoder plugin."
		return 1
	fi
}

# This group only gives the process read/traverse ability; RBAC still decides
# each caller's read access. Existing datasets receive ACLs at deployment.
deploy.streamService.provisionReadACL() {
	local root
	local mount_options
	local roots=(/home)
	if ! command -v setfacl >/dev/null 2>&1; then
		log.warn "setfacl is unavailable; continuing without storage read ACL provisioning."
		return 0
	fi
	while IFS= read -r root; do
		case "$root" in
		/bin | /bin/* | /boot | /boot/* | /dev | /dev/* | /etc | /etc/* | /lib | /lib/* | /lib64 | /lib64/* | /proc | /proc/* | /root | /root/* | /run | /run/* | /sbin | /sbin/* | /snap | /snap/* | /sys | /sys/* | /usr | /usr/* | /var | /var/*)
			continue
			;;
		/*) roots+=("$root") ;;
		*)
			log.error "Invalid ZFS mountpoint: $root"
			return 1
			;;
		esac
	done < <(findmnt -rn -t zfs -o TARGET)
	for root in "${roots[@]}"; do
		if [ ! -d "$root" ] || [ "$root" = / ]; then
			continue
		fi
		if ! mount_options="$(findmnt -rn -T "$root" -o OPTIONS)"; then
			log.warn "Could not inspect mount options for $root; continuing without its read ACL."
			continue
		fi
		if [[ ",$mount_options," == *,noacl,* ]]; then
			log.warn "Skipping read ACL provisioning on $root: its filesystem has ACLs disabled."
			continue
		fi
		if ! find "$root" -xdev -type d -exec setfacl -m \
			"g:$LITE_NAS_STREAM_SERVICE_READ_GROUP:rx,d:g:$LITE_NAS_STREAM_SERVICE_READ_GROUP:rx" {} +; then
			log.warn "Could not provision every directory read ACL under $root; continuing deployment."
		fi
		if ! find "$root" -xdev -type f -exec setfacl -m "g:$LITE_NAS_STREAM_SERVICE_READ_GROUP:r" {} +; then
			log.warn "Could not provision every file read ACL under $root; continuing deployment."
		fi
	done
}

deploy.streamService.ensureUser() {
	deploy.ensureSystemGroup lite-nas "LiteNAS configuration group"
	deploy.ensureSystemGroup "$LITE_NAS_STREAM_SERVICE_USER" "stream service group"
	deploy.ensureSystemGroup "$LITE_NAS_STREAM_SERVICE_READ_GROUP" "storage read group"
	if ! id "$LITE_NAS_STREAM_SERVICE_USER" >/dev/null 2>&1; then
		useradd --system --no-create-home --home-dir /nonexistent \
			--shell /usr/sbin/nologin --gid "$LITE_NAS_STREAM_SERVICE_USER" \
			--groups "lite-nas,$LITE_NAS_STREAM_SERVICE_READ_GROUP" \
			"$LITE_NAS_STREAM_SERVICE_USER"
	else
		usermod --append --groups "lite-nas,$LITE_NAS_STREAM_SERVICE_READ_GROUP" \
			"$LITE_NAS_STREAM_SERVICE_USER"
	fi
}

deploy.streamService.enableAndStart() {
	deploy.enableAndRefreshService lite-nas-stream-service.service
}

deploy.streamService.deploy() {
	local source_binary="$1"
	local should_start="${2:-1}"
	if [ ! -f "$source_binary" ]; then
		log.error "Missing stream-service binary: $source_binary"
		return 1
	fi
	deploy.streamService.requireTools
	deploy.streamService.ensureUser
	deploy.streamService.provisionReadACL
	"$LITE_NAS_REPO_ROOT/scripts/rotate-nats-certificates.sh" --if-missing --user "$LITE_NAS_STREAM_SERVICE_USER"
	install -D -m 0755 -o root -g root -- "$source_binary" "$LITE_NAS_STREAM_SERVICE_BINARY_TARGET"
	install -D -m 0640 -o "$LITE_NAS_STREAM_SERVICE_USER" -g lite-nas \
		"$LITE_NAS_STREAM_SERVICE_CONFIG_SOURCE" /etc/lite-nas/stream-service.conf
	install -D -m 0644 -o root -g root "$LITE_NAS_STREAM_SERVICE_UNIT_SOURCE" \
		/etc/systemd/system/lite-nas-stream-service.service
	install -d -m 0751 -o root -g lite-nas /var/log/lite-nas
	if [ ! -e /var/log/lite-nas/stream-service.log ]; then
		install -m 0640 -o "$LITE_NAS_STREAM_SERVICE_USER" -g "$LITE_NAS_STREAM_SERVICE_USER" \
			/dev/null /var/log/lite-nas/stream-service.log
	fi
	chown "$LITE_NAS_STREAM_SERVICE_USER:$LITE_NAS_STREAM_SERVICE_USER" /var/log/lite-nas/stream-service.log
	chmod 0640 /var/log/lite-nas/stream-service.log
	deploy.validateSudoersDropIn "$LITE_NAS_STREAM_SERVICE_SUDOERS_SOURCE"
	install -D -m 0440 -o root -g root "$LITE_NAS_STREAM_SERVICE_SUDOERS_SOURCE" \
		/etc/sudoers.d/lite-nas-stream-service-read
	install -D -m 0644 -o root -g root "$LITE_NAS_STREAM_SERVICE_PROFILE_SOURCE" \
		/etc/apparmor.d/usr.libexec.lite-nas.stream-service
	apparmor_parser -r -I /etc/apparmor.d /etc/apparmor.d/usr.libexec.lite-nas.stream-service
	if [ "$should_start" = "1" ]; then
		deploy.streamService.enableAndStart
	fi
}
