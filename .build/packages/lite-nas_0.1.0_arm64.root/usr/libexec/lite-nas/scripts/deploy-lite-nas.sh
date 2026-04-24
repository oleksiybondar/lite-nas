#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/deploy/lite-nas.sh"

manage_nats_config=1

while [ "$#" -gt 0 ]; do
	case "$1" in
	--no-nats-config)
		manage_nats_config=0
		shift
		;;
	-h | --help)
		deploy.liteNAS.usage
		exit 0
		;;
	*)
		log.error "Unknown option: $1"
		deploy.liteNAS.usage >&2
		exit 2
		;;
	esac
done

sudo.guard.requireRoot "scripts/deploy-lite-nas.sh"
deploy.liteNAS.requireTools
deploy.liteNAS.bootstrap "$manage_nats_config"
