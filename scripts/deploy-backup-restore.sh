#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/common.sh"

sudo.guard.requireRoot "scripts/deploy-backup-restore.sh"
for tool in go install groupadd getent id useradd usermod systemctl apparmor_parser touch chown chmod mktemp rm ln; do
	log.requireCommand "$tool" "Install the required system tooling and retry."
done

if ! getent group lite-nas-backup-restore-cli >/dev/null 2>&1; then
	groupadd --system lite-nas-backup-restore-cli
fi
if ! id lite-nas-backup-restore-cli >/dev/null 2>&1; then
	useradd --system --no-create-home --home-dir /nonexistent --shell /usr/sbin/nologin \
		--gid lite-nas-backup-restore-cli lite-nas-backup-restore-cli
fi

build_dir="$(mktemp -d)"
trap 'rm -rf "$build_dir"' EXIT
"$SCRIPT_DIR/build-backup-restore-binary.sh" "--output=$build_dir/backup-restore"
"$SCRIPT_DIR/build-backup-restore-cli-binary.sh" "--output=$build_dir/backup-restore-cli"

install -D -o root -g root -m 0755 "$build_dir/backup-restore" /usr/libexec/lite-nas/backup-restore
install -D -o root -g root -m 0700 "$build_dir/backup-restore-cli" /usr/libexec/lite-nas/backup-restore-cli
ln -sfn /usr/libexec/lite-nas/backup-restore-cli /usr/bin/backup-restore-cli
install -d -o root -g root -m 0700 /var/lib/lite-nas/backup
install -d -o root -g root -m 0751 /var/log/lite-nas
touch /var/log/lite-nas/backup-restore.log /var/log/lite-nas/backup-restore-cli.log
chown root:root /var/log/lite-nas/backup-restore.log /var/log/lite-nas/backup-restore-cli.log
chmod 0600 /var/log/lite-nas/backup-restore.log /var/log/lite-nas/backup-restore-cli.log

"$SCRIPT_DIR/deploy-configs.sh" --no-restart
"$SCRIPT_DIR/rotate-nats-certificates.sh" --if-missing
"$SCRIPT_DIR/rotate-auth-token-certificates.sh" --if-missing

install -D -o root -g root -m 0644 \
	"$SCRIPT_DIR/../configs/etc/systemd/system/lite-nas-backup-restore.service" \
	/etc/systemd/system/lite-nas-backup-restore.service
apparmor_parser -r /etc/apparmor.d/usr.libexec.lite-nas.backup-restore
apparmor_parser -r /etc/apparmor.d/usr.libexec.lite-nas.backup-restore-cli
systemctl daemon-reload
systemctl restart nats-server.service
systemctl enable --now lite-nas-backup-restore.service
