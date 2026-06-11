#!/bin/sh
set -eu

readonly LITE_NAS_POSTFIX_CAPTURE_FILE="${LITE_NAS_POSTFIX_CAPTURE_FILE:-/var/tmp/lite-nas-postfix-test-mail.log}"
readonly LITE_NAS_POSTFIX_CAPTURE_LOCK_DIR="${LITE_NAS_POSTFIX_CAPTURE_LOCK_DIR:-${LITE_NAS_POSTFIX_CAPTURE_FILE}.lock}"

sender="${1:-}"
recipient="${2:-}"

acquire_lock() {
	while ! mkdir "$LITE_NAS_POSTFIX_CAPTURE_LOCK_DIR" 2>/dev/null; do
		:
	done
}

release_lock() {
	rmdir "$LITE_NAS_POSTFIX_CAPTURE_LOCK_DIR"
}

main() {
	acquire_lock
	trap release_lock EXIT

	{
		printf '=== LITENAS TEST MAIL BEGIN ===\n'
		printf 'Envelope-Sender: %s\n' "$sender"
		printf 'Envelope-Recipient: %s\n' "$recipient"
		printf '\n'
		cat
		printf '\n=== LITENAS TEST MAIL END ===\n\n'
	} >>"$LITE_NAS_POSTFIX_CAPTURE_FILE"
}

main "$@"
