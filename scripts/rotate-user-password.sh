#!/usr/bin/env bash
# Set a policy-checked temporary password and require a change at next login.
set +x
set -euo pipefail
umask 077

usage() {
	printf 'Usage: sudo %s USERNAME\n' "$0"
}

if [ "$EUID" -ne 0 ] || [ -z "${SUDO_COMMAND:-}" ] || [ -z "${SUDO_USER:-}" ] || [[ ! "${SUDO_UID:-}" =~ ^[0-9]+$ ]]; then
	printf 'Run this script with sudo.\n' >&2
	exit 1
fi

if [ "$#" -ne 1 ] || [[ ! "$1" =~ ^[a-zA-Z_][a-zA-Z0-9_.-]*\$?$ ]]; then
	usage >&2
	exit 2
fi
username="$1"

for tool in getent pwmake chpasswd passwd; do
	if ! command -v "$tool" >/dev/null 2>&1; then
		printf 'Required command is missing: %s\n' "$tool" >&2
		exit 1
	fi
done

if ! getent passwd "$username" >/dev/null; then
	printf 'User does not exist: %s\n' "$username" >&2
	exit 2
fi

# pwmake uses libpwquality and its /etc/security/pwquality.conf settings.
password="$(pwmake 128)"
if [ -z "$password" ] || [[ "$password" == *$'\n'* || "$password" == *$'\r'* ]]; then
	printf 'Password generator returned an invalid password.\n' >&2
	exit 1
fi

if ! printf '%s:%s\n' "$username" "$password" | chpasswd; then
	printf 'Could not set password for %s.\n' "$username" >&2
	exit 1
fi

if ! passwd --expire -- "$username"; then
	printf 'Password was changed, but expiry failed for %s; handle this account now.\n' "$username" >&2
	printf 'Username: %s\nPassword: %s\n' "$username" "$password"
	exit 1
fi

printf 'Username: %s\nPassword: %s\n' "$username" "$password"
