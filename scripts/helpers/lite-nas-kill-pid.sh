#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 1 ] || [[ ! "$1" =~ ^[1-9][0-9]*$ ]]; then
	printf 'Usage: lite-nas-kill-pid PID\n' >&2
	exit 64
fi

exec /usr/bin/kill -- "$1"
