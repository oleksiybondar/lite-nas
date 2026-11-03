#!/usr/bin/env bash

DEPLOY_HELPER_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$DEPLOY_HELPER_DIR/../helpers/common.sh"

deploy.backupRestore.requireTools() {
	local tool
	for tool in install systemctl apparmor_parser touch chown chmod ln; do
		log.requireCommand "$tool" "Install the required system tooling and retry."
	done
}

deploy.backupRestore.deploy() {
	local binary="$1"
	local should_start="${2:-1}"
	local target=/usr/libexec/lite-nas/backup-restore
	if [ ! -f "$binary" ]; then
		log.error "Missing backup-restore binary: $binary"
		return 1
	fi
	install -D -o root -g root -m 0755 "$binary" "$target"
	install -d -o root -g root -m 0700 /var/lib/lite-nas/backup
	install -d -o root -g root -m 0751 /var/log/lite-nas
	if [ ! -f /var/log/lite-nas/backup-restore.log ]; then
		install -o root -g root -m 0600 /dev/null /var/log/lite-nas/backup-restore.log
	fi
	"$LITE_NAS_REPO_ROOT/scripts/deploy-configs.sh" --no-restart
	"$LITE_NAS_REPO_ROOT/scripts/rotate-nats-certificates.sh" --if-missing
	"$LITE_NAS_REPO_ROOT/scripts/rotate-auth-token-certificates.sh" --if-missing
	install -D -o root -g root -m 0644 \
		"$LITE_NAS_REPO_ROOT/configs/etc/systemd/system/lite-nas-backup-restore.service" \
		/etc/systemd/system/lite-nas-backup-restore.service
	apparmor_parser -r /etc/apparmor.d/usr.libexec.lite-nas.backup-restore
	systemctl daemon-reload
	if [ "$should_start" = "1" ]; then
		deploy.enableAndRefreshService lite-nas-backup-restore.service
	else
		log.info "Skipping backup-restore enable/start."
	fi
}
