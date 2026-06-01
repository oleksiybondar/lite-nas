#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PACKAGE_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/helpers/common.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/apparmor.sh"
# shellcheck disable=SC1091
source "$PACKAGE_ROOT/scripts/deploy/postfix.sh"

readonly LITE_NAS_POSTFIX_TEST_CAPTURE_CONFIG_TARGET_DIR="${LITE_NAS_POSTFIX_CONFIG_TARGET_DIR:-/etc/postfix}"
readonly LITE_NAS_POSTFIX_TEST_CAPTURE_MAIN_CF_TARGET="${LITE_NAS_POSTFIX_MAIN_CF_TARGET:-$LITE_NAS_POSTFIX_TEST_CAPTURE_CONFIG_TARGET_DIR/main.cf}"
readonly LITE_NAS_POSTFIX_TEST_CAPTURE_MASTER_CF_TARGET="${LITE_NAS_POSTFIX_MASTER_CF_TARGET:-$LITE_NAS_POSTFIX_TEST_CAPTURE_CONFIG_TARGET_DIR/master.cf}"
readonly LITE_NAS_POSTFIX_TEST_CAPTURE_NORMAL_CONFIG_DIR="${LITE_NAS_POSTFIX_NORMAL_CONFIG_DIR:-$PACKAGE_ROOT/configs/etc/postfix}"
readonly LITE_NAS_POSTFIX_TEST_CAPTURE_TEST_CONFIG_DIR="${LITE_NAS_POSTFIX_TEST_CONFIG_DIR:-$PACKAGE_ROOT/configs/etc/postfix/testing}"
readonly LITE_NAS_POSTFIX_TEST_CAPTURE_HELPER_SOURCE="${LITE_NAS_POSTFIX_CAPTURE_HELPER_SOURCE:-$PACKAGE_ROOT/scripts/postfix/postfix-test-pipe.sh}"
readonly LITE_NAS_POSTFIX_TEST_CAPTURE_HELPER_TARGET="${LITE_NAS_POSTFIX_CAPTURE_HELPER_TARGET:-/usr/libexec/lite-nas/postfix-test-pipe}"
readonly LITE_NAS_POSTFIX_TEST_CAPTURE_FILE="${LITE_NAS_POSTFIX_CAPTURE_FILE:-/var/tmp/lite-nas-postfix-test-mail.log}"
readonly LITE_NAS_POSTFIX_TEST_CAPTURE_LOCK_DIR="${LITE_NAS_POSTFIX_CAPTURE_LOCK_DIR:-${LITE_NAS_POSTFIX_TEST_CAPTURE_FILE}.lock}"
readonly LITE_NAS_POSTFIX_TEST_CAPTURE_SERVICE_NAME="${LITE_NAS_POSTFIX_SERVICE_NAME:-postfix}"

mode="capture"
reset_capture=1

usage() {
	cat <<'MSG'
Usage: scripts/postfix/configure-test-capture.sh [options]

Options:
  --mode MODE         MODE is "capture" (default) or "normal".
  --keep-capture      Preserve existing capture file contents in capture mode.
  -h, --help          Show this help.
MSG
}

require_tools() {
	local tool
	local tools=(cat chmod chown install mkdir postconf systemctl)
	for tool in "${tools[@]}"; do
		log.requireCommand "$tool" "Install the required tooling and retry."
	done
}

install_postfix_config() {
	local source_dir="$1"

	if [ ! -f "$source_dir/main.cf" ] || [ ! -f "$source_dir/master.cf" ]; then
		log.error "Missing Postfix config files in $source_dir"
		exit 1
	fi

	install -d -m 0755 "$LITE_NAS_POSTFIX_TEST_CAPTURE_CONFIG_TARGET_DIR"
	install -m 0644 "$source_dir/main.cf" "$LITE_NAS_POSTFIX_TEST_CAPTURE_MAIN_CF_TARGET"
	install -m 0644 "$source_dir/master.cf" "$LITE_NAS_POSTFIX_TEST_CAPTURE_MASTER_CF_TARGET"
}

install_capture_helper() {
	local target_dir

	if [ ! -f "$LITE_NAS_POSTFIX_TEST_CAPTURE_HELPER_SOURCE" ]; then
		log.error "Missing capture helper source: $LITE_NAS_POSTFIX_TEST_CAPTURE_HELPER_SOURCE"
		exit 1
	fi

	target_dir="$(dirname "$LITE_NAS_POSTFIX_TEST_CAPTURE_HELPER_TARGET")"
	install -d -m 0755 "$target_dir"
	install -m 0755 "$LITE_NAS_POSTFIX_TEST_CAPTURE_HELPER_SOURCE" "$LITE_NAS_POSTFIX_TEST_CAPTURE_HELPER_TARGET"
}

prepare_capture_file() {
	local capture_dir

	capture_dir="$(dirname "$LITE_NAS_POSTFIX_TEST_CAPTURE_FILE")"
	install -d -m 0755 "$capture_dir"

	if [ "$reset_capture" = "1" ]; then
		: >"$LITE_NAS_POSTFIX_TEST_CAPTURE_FILE"
	fi

	chown root:root "$LITE_NAS_POSTFIX_TEST_CAPTURE_FILE"
	chmod 0644 "$LITE_NAS_POSTFIX_TEST_CAPTURE_FILE"

	rm -rf "$LITE_NAS_POSTFIX_TEST_CAPTURE_LOCK_DIR"
}

validate_postfix_config() {
	postconf -n >/dev/null
	postconf default_transport >/dev/null
}

restart_postfix() {
	systemctl daemon-reload
	systemctl restart "$LITE_NAS_POSTFIX_TEST_CAPTURE_SERVICE_NAME.service"
}

while [ "$#" -gt 0 ]; do
	case "$1" in
	--mode)
		if [ -z "${2:-}" ]; then
			log.error "Missing value for --mode"
			usage >&2
			exit 2
		fi
		mode="$2"
		shift 2
		;;
	--keep-capture)
		reset_capture=0
		shift
		;;
	-h | --help)
		usage
		exit 0
		;;
	*)
		log.error "Unknown option: $1"
		usage >&2
		exit 2
		;;
	esac
done

case "$mode" in
capture | normal) ;;
*)
	log.error "Unsupported mode: $mode"
	usage >&2
	exit 2
	;;
esac

sudo.guard.requireRoot "scripts/postfix/configure-test-capture.sh"
require_tools
deploy.apparmor.requireTools

log.pushTask "Deploying managed AppArmor profiles"
deploy.apparmor.deploy 1
log.popTask

if [ "$mode" = "capture" ]; then
	log.pushTask "Configuring Postfix capture mode"
	install_postfix_config "$LITE_NAS_POSTFIX_TEST_CAPTURE_TEST_CONFIG_DIR"
	install_capture_helper
	prepare_capture_file
	validate_postfix_config
	restart_postfix
	log.popTask
	log.info "Postfix capture mode enabled. Mail is appended to $LITE_NAS_POSTFIX_TEST_CAPTURE_FILE"
	exit 0
fi

log.pushTask "Restoring normal Postfix mode"
install_postfix_config "$LITE_NAS_POSTFIX_TEST_CAPTURE_NORMAL_CONFIG_DIR"
deploy.postfix.applyAuthenticationConfig
validate_postfix_config
restart_postfix
log.popTask
log.info "Normal Postfix mode restored."
