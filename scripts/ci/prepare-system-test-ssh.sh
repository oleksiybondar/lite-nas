#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/logger.sh"

if [ "$EUID" -ne 0 ]; then
	log.error "Run this test-host preparation script with sudo."
	exit 1
fi

log.pushTask "Preparing loopback SSH password authentication for system tests"
apt-get update
DEBIAN_FRONTEND=noninteractive apt-get install -y openssh-server
install -d -m 0755 /etc/ssh/sshd_config.d
# Only loopback connections need password prompts for the SIEM rejection tests.
cat >/etc/ssh/sshd_config.d/00-lite-nas-system-tests.conf <<'EOF'
Match LocalAddress 127.0.0.1,::1
    PasswordAuthentication yes
    KbdInteractiveAuthentication no
Match all
EOF
/usr/sbin/sshd -t
systemctl restart ssh
log.popTask
