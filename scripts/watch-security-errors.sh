#!/usr/bin/env bash
# Live view of warnings/errors from NATS, the security services, the alert
# pipeline and the mail path (email notifiers + postfix), to trace why alerts
# are not being delivered.
#
# Sources:
#   - journald: NATS server, postfix, and the LiteNAS units (stderr/startup
#     failures, restarts, crashes). Services send stdout to null, so the journal
#     only holds what they print on stderr and systemd lifecycle events.
#   - /var/log/lite-nas/<service>.log: the services' own RFC 5424 logs, which
#     is where runtime errors normally go (output=file in each service config).
#
# Usage: sudo bash watch-security-errors.sh [--all] [--since "10 min ago"]
set -euo pipefail

# Standalone: no repository helpers required, so it can be copied to any host.
log.info() { printf '[INFO] %s\n' "$*" >&2; }
log.warn() { printf '[WARN] %s\n' "$*" >&2; }
log.error() { printf '[ERROR] %s\n' "$*" >&2; }

LOG_DIR="/var/log/lite-nas"

# Security, alerting and email related LiteNAS services (systemd unit suffix ==
# log file base name unless overridden below).
services=(
	security-automation
	security-logging-manager
	security-email-notifier
	system-logging-manager
	system-email-notifier
	network-security
	log-analysis
	antivirus
	service-watchdog
)
# Units whose log file name differs from the unit name.
declare -A log_file_name=(
	[integrity]="lite-nas-integrity"
)
services+=(integrity)

# External units that are part of the alert path.
external_units=(nats-server "postfix*")

show_all=0
since="now"

usage() {
	cat <<'EOF'
Usage: sudo watch-security-errors.sh [--all] [--since TIME]

Follows NATS, security services, alert/email services and postfix in real time.

  --all          Show every level (default: warning and worse only)
  --since TIME   Also show history from TIME (journalctl syntax, e.g. "30 min ago")
  -h, --help     Show this help
EOF
}

while [ "$#" -gt 0 ]; do
	case "$1" in
	--all)
		show_all=1
		shift
		;;
	--since)
		if [ -z "${2:-}" ]; then
			log.error "Missing value for --since"
			exit 2
		fi
		since="$2"
		shift 2
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

if [ "$(id -u)" -ne 0 ]; then
	log.error "Run with sudo or as root (log files in $LOG_DIR are root-only)."
	exit 1
fi
if ! command -v journalctl >/dev/null 2>&1; then
	log.error "journalctl not found; systemd is required."
	exit 1
fi

# Prefix each line with a source label and, unless --all, drop RFC 5424 lines
# less severe than warning (severity = PRI % 8, warning = 4). Lines that are not
# RFC 5424 are kept so unexpected output is never hidden.
filter_log_lines() {
	local label="$1"
	awk -v label="$label" -v all="$show_all" '
		{
			if (!all && match($0, /^<[0-9]+>/)) {
				pri = substr($0, 2, RLENGTH - 2) + 0
				if ((pri % 8) > 4) next
			}
			printf "[%s] %s\n", label, $0
			fflush()
		}'
}

pids=()
cleanup() {
	trap - EXIT INT TERM
	# Signal the whole process group so the `tail` side of each tail|awk
	# pipeline stops too, not only the last pipeline element.
	kill 0 2>/dev/null || true
}
trap cleanup EXIT INT TERM

# Journal: all units in one merged stream; journalctl prints the unit name.
journal_args=(-f --no-pager -o short-iso -p "$([ "$show_all" -eq 1 ] && echo debug || echo warning)")
if [ "$since" = "now" ]; then
	journal_args+=(-n 0)
else
	journal_args+=(--since "$since")
fi
for svc in "${services[@]}"; do
	journal_args+=(-u "lite-nas-${svc}.service")
done
for unit in "${external_units[@]}"; do
	journal_args+=(-u "$unit")
done

log.info "Watching journal: NATS, postfix and ${#services[@]} LiteNAS services ($([ "$show_all" -eq 1 ] && echo 'all levels' || echo 'warning+'))"
journalctl "${journal_args[@]}" &
pids+=("$!")

# Service log files.
tail_n=0
[ "$since" = "now" ] || tail_n=50
for svc in "${services[@]}"; do
	file="$LOG_DIR/${log_file_name[$svc]:-$svc}.log"
	if [ ! -e "$file" ]; then
		log.warn "Log file missing (service never started or logging not to file): $file"
		continue
	fi
	tail -n "$tail_n" -F "$file" 2>/dev/null | filter_log_lines "$svc" &
	pids+=("$!")
done

log.info "Press Ctrl+C to stop."
wait
