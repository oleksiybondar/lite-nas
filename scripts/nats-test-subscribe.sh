#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/helpers/logger.sh"

subscription=""
server_override=""
ca_override=""
cert_override=""
key_override=""
profile_override=""

declare -A PROFILE_SERVER
declare -A PROFILE_CA
declare -A PROFILE_CERT
declare -A PROFILE_KEY

declare -A SUBJECT_PROFILE

PROFILE_SERVER["system-metrics-cli"]="tls://127.0.0.1:4222"
PROFILE_CA["system-metrics-cli"]="/etc/lite-nas/certificates/transport/root-ca.crt"
PROFILE_CERT["system-metrics-cli"]="/etc/lite-nas/certificates/transport/lite-nas-system-metrics-cli/client.crt"
PROFILE_KEY["system-metrics-cli"]="/etc/lite-nas/certificates/transport/lite-nas-system-metrics-cli/client.key"

PROFILE_SERVER["auth-service"]="tls://127.0.0.1:4222"
PROFILE_CA["auth-service"]="/etc/lite-nas/certificates/transport/root-ca.crt"
PROFILE_CERT["auth-service"]="/etc/lite-nas/certificates/transport/lite-nas-auth-service/client.crt"
PROFILE_KEY["auth-service"]="/etc/lite-nas/certificates/transport/lite-nas-auth-service/client.key"

PROFILE_SERVER["system-logging-manager-cli"]="tls://127.0.0.1:4222"
PROFILE_CA["system-logging-manager-cli"]="/etc/lite-nas/certificates/transport/root-ca.crt"
PROFILE_CERT["system-logging-manager-cli"]="/etc/lite-nas/certificates/transport/lite-nas-sys-log-mgr-cli/client.crt"
PROFILE_KEY["system-logging-manager-cli"]="/etc/lite-nas/certificates/transport/lite-nas-sys-log-mgr-cli/client.key"

PROFILE_SERVER["security-logging-manager-cli"]="tls://127.0.0.1:4222"
PROFILE_CA["security-logging-manager-cli"]="/etc/lite-nas/certificates/transport/root-ca.crt"
PROFILE_CERT["security-logging-manager-cli"]="/etc/lite-nas/certificates/transport/lite-nas-sec-log-mgr-cli/client.crt"
PROFILE_KEY["security-logging-manager-cli"]="/etc/lite-nas/certificates/transport/lite-nas-sec-log-mgr-cli/client.key"

PROFILE_SERVER["zfs-metrics"]="tls://127.0.0.1:4222"
PROFILE_CA["zfs-metrics"]="/etc/lite-nas/certificates/transport/root-ca.crt"
PROFILE_CERT["zfs-metrics"]="/etc/lite-nas/certificates/transport/lite-nas-zfs-metrics/client.crt"
PROFILE_KEY["zfs-metrics"]="/etc/lite-nas/certificates/transport/lite-nas-zfs-metrics/client.key"

PROFILE_SERVER["zfs-metrics-cli"]="tls://127.0.0.1:4222"
PROFILE_CA["zfs-metrics-cli"]="/etc/lite-nas/certificates/transport/root-ca.crt"
PROFILE_CERT["zfs-metrics-cli"]="/etc/lite-nas/certificates/transport/lite-nas-zfs-metrics-cli/client.crt"
PROFILE_KEY["zfs-metrics-cli"]="/etc/lite-nas/certificates/transport/lite-nas-zfs-metrics-cli/client.key"

supported_subscriptions=(
	"system.metrics.events.stats"
	"system.metrics.rpc.stats.get"
	"system.metrics.rpc.history.get"
	"auth.rpc.login"
	"auth.rpc.refresh"
	"auth.rpc.logout"
	"auth.rpc.token.validate"
	"auth.rpc.lockdown.set"
	"auth.events.lockdown.changed"
	"system-alert"
	"system-alert-occurrence"
	"system-logging-manager.getAlerts"
	"system-logging-manager.getAlert"
	"system-logging-manager.getActiveAlerts"
	"system-logging-manager.getUnacknowledgedActiveAlerts"
	"system-logging-manager.updateAlertState"
	"system-logging-manager.acknowledgeAlert"
	"system-logging-manager.muteAlert"
	"security-alert"
	"security-alert-occurrence"
	"security-logging-manager.getAlerts"
	"security-logging-manager.getAlert"
	"security-logging-manager.getActiveAlerts"
	"security-logging-manager.getUnacknowledgedActiveAlerts"
	"security-logging-manager.updateAlertState"
	"security-logging-manager.acknowledgeAlert"
	"security-logging-manager.muteAlert"
	"zfs.metrics.events.snapshot"
	"zfs.metrics.rpc.snapshot.get"
)

for subject in "${supported_subscriptions[@]}"; do
	SUBJECT_PROFILE["$subject"]="system-metrics-cli"
done
SUBJECT_PROFILE["auth.rpc.login"]="auth-service"
SUBJECT_PROFILE["auth.rpc.refresh"]="auth-service"
SUBJECT_PROFILE["auth.rpc.logout"]="auth-service"
SUBJECT_PROFILE["auth.rpc.token.validate"]="auth-service"
SUBJECT_PROFILE["auth.rpc.lockdown.set"]="auth-service"
SUBJECT_PROFILE["auth.events.lockdown.changed"]="auth-service"
SUBJECT_PROFILE["system-alert"]="system-logging-manager-cli"
SUBJECT_PROFILE["system-alert-occurrence"]="system-logging-manager-cli"
SUBJECT_PROFILE["system-logging-manager.getAlerts"]="system-logging-manager-cli"
SUBJECT_PROFILE["system-logging-manager.getAlert"]="system-logging-manager-cli"
SUBJECT_PROFILE["system-logging-manager.getActiveAlerts"]="system-logging-manager-cli"
SUBJECT_PROFILE["system-logging-manager.getUnacknowledgedActiveAlerts"]="system-logging-manager-cli"
SUBJECT_PROFILE["system-logging-manager.updateAlertState"]="system-logging-manager-cli"
SUBJECT_PROFILE["system-logging-manager.acknowledgeAlert"]="system-logging-manager-cli"
SUBJECT_PROFILE["system-logging-manager.muteAlert"]="system-logging-manager-cli"
SUBJECT_PROFILE["security-alert"]="security-logging-manager-cli"
SUBJECT_PROFILE["security-alert-occurrence"]="security-logging-manager-cli"
SUBJECT_PROFILE["security-logging-manager.getAlerts"]="security-logging-manager-cli"
SUBJECT_PROFILE["security-logging-manager.getAlert"]="security-logging-manager-cli"
SUBJECT_PROFILE["security-logging-manager.getActiveAlerts"]="security-logging-manager-cli"
SUBJECT_PROFILE["security-logging-manager.getUnacknowledgedActiveAlerts"]="security-logging-manager-cli"
SUBJECT_PROFILE["security-logging-manager.updateAlertState"]="security-logging-manager-cli"
SUBJECT_PROFILE["security-logging-manager.acknowledgeAlert"]="security-logging-manager-cli"
SUBJECT_PROFILE["security-logging-manager.muteAlert"]="security-logging-manager-cli"
SUBJECT_PROFILE["zfs.metrics.events.snapshot"]="zfs-metrics-cli"
SUBJECT_PROFILE["zfs.metrics.rpc.snapshot.get"]="zfs-metrics-cli"

usage() {
	cat <<'MSG'
Usage: scripts/nats-test-subscribe.sh --subscription SUBJECT [options]

Required:
  --subscription SUBJECT   NATS subject to subscribe to.

Options:
  --profile NAME           TLS profile to use (default chosen by subject map).
  --config PATH            Deprecated. Ignored.
  --server URL             Override messaging URL from config.
  --ca PATH                Override CA certificate path from profile map.
  --cert PATH              Override client certificate path from profile map.
  --key PATH               Override client private key path from profile map.
  -h, --help               Show this help.

Examples:
  scripts/nats-test-subscribe.sh --subscription system.metrics.events.stats
  scripts/nats-test-subscribe.sh --subscription auth.events.lockdown.changed --profile auth-service

Profiles:
  - system-metrics-cli
  - auth-service
  - system-logging-manager-cli
  - security-logging-manager-cli
  - zfs-metrics
  - zfs-metrics-cli

Supported subscriptions:
  - system.metrics.events.stats
  - system.metrics.rpc.stats.get
  - system.metrics.rpc.history.get
  - auth.rpc.login
  - auth.rpc.refresh
  - auth.rpc.logout
  - auth.rpc.token.validate
  - auth.rpc.lockdown.set
  - auth.events.lockdown.changed
  - system-alert
  - system-alert-occurrence
  - system-logging-manager.getAlerts
  - system-logging-manager.getAlert
  - system-logging-manager.getActiveAlerts
  - system-logging-manager.getUnacknowledgedActiveAlerts
  - system-logging-manager.updateAlertState
  - system-logging-manager.acknowledgeAlert
  - system-logging-manager.muteAlert
  - security-alert
  - security-alert-occurrence
  - security-logging-manager.getAlerts
  - security-logging-manager.getAlert
  - security-logging-manager.getActiveAlerts
  - security-logging-manager.getUnacknowledgedActiveAlerts
  - security-logging-manager.updateAlertState
  - security-logging-manager.acknowledgeAlert
  - security-logging-manager.muteAlert
  - zfs.metrics.events.snapshot
  - zfs.metrics.rpc.snapshot.get
MSG
}

while [ "$#" -gt 0 ]; do
	case "$1" in
	--subscription)
		if [ -z "${2:-}" ]; then
			log.error "Missing value for --subscription"
			usage >&2
			exit 2
		fi
		subscription="$2"
		shift 2
		;;
	--profile)
		if [ -z "${2:-}" ]; then
			log.error "Missing value for --profile"
			usage >&2
			exit 2
		fi
		profile_override="$2"
		shift 2
		;;
	--config)
		shift 2
		;;
	--server)
		if [ -z "${2:-}" ]; then
			log.error "Missing value for --server"
			usage >&2
			exit 2
		fi
		server_override="$2"
		shift 2
		;;
	--ca)
		if [ -z "${2:-}" ]; then
			log.error "Missing value for --ca"
			usage >&2
			exit 2
		fi
		ca_override="$2"
		shift 2
		;;
	--cert)
		if [ -z "${2:-}" ]; then
			log.error "Missing value for --cert"
			usage >&2
			exit 2
		fi
		cert_override="$2"
		shift 2
		;;
	--key)
		if [ -z "${2:-}" ]; then
			log.error "Missing value for --key"
			usage >&2
			exit 2
		fi
		key_override="$2"
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

if [ -z "$subscription" ]; then
	log.error "Missing required --subscription option."
	usage >&2
	exit 2
fi

log.requireCommand "nats" "Install NATS CLI and retry."

subject_is_supported=0
for current_subject in "${supported_subscriptions[@]}"; do
	if [ "$current_subject" = "$subscription" ]; then
		subject_is_supported=1
		break
	fi
done
if [ "$subject_is_supported" -ne 1 ]; then
	log.error "Unsupported subscription: $subscription"
	log.info "Run with --help to see supported subscriptions."
	exit 2
fi

profile_name="$profile_override"
if [ -z "$profile_name" ]; then
	profile_name="${SUBJECT_PROFILE[$subscription]}"
fi
if [ -z "$profile_name" ]; then
	log.error "No profile is mapped for subscription: $subscription"
	exit 2
fi
if [ -z "${PROFILE_SERVER[$profile_name]:-}" ] || [ -z "${PROFILE_CA[$profile_name]:-}" ] || [ -z "${PROFILE_CERT[$profile_name]:-}" ] || [ -z "${PROFILE_KEY[$profile_name]:-}" ]; then
	log.error "Unknown or incomplete profile: $profile_name"
	exit 2
fi

server_url="${server_override:-${PROFILE_SERVER[$profile_name]}}"
ca_file="${ca_override:-${PROFILE_CA[$profile_name]}}"
cert_file="${cert_override:-${PROFILE_CERT[$profile_name]}}"
key_file="${key_override:-${PROFILE_KEY[$profile_name]}}"

if [ -z "$server_url" ]; then
	log.error "Missing messaging.url in config and no --server override was provided."
	exit 1
fi

for path_value in "$ca_file" "$cert_file" "$key_file"; do
	if [ -z "$path_value" ]; then
		log.error "Missing one or more TLS paths (ca/cert/key)."
		exit 1
	fi
	if [ ! -f "$path_value" ]; then
		log.error "TLS file not found: $path_value"
		exit 1
	fi
done

log.info "Subscribing to subject: $subscription"
log.info "Profile: $profile_name"
log.info "Server: $server_url"

nats \
	--server "$server_url" \
	--tlsca "$ca_file" \
	--tlscert "$cert_file" \
	--tlskey "$key_file" \
	sub "$subscription"
