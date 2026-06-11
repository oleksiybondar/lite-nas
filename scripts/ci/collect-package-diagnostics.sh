#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1091
source "$SCRIPT_DIR/../helpers/common.sh"

output_dir=""

usage() {
	cat <<'MSG'
Usage: scripts/ci/collect-package-diagnostics.sh --output-dir PATH

Options:
  --output-dir PATH   Directory where package diagnostics will be written.
  -h, --help          Show this help.
MSG
}

args.parse "$@"
if ! args.assertKnown output-dir help h; then
	log.error "Unknown option: --$(args.unknownKeys output-dir help h | head -n 1)"
	usage >&2
	exit 64
fi
if args.has h || args.has help; then
	usage
	exit 0
fi
if args.has output-dir && ! output_dir="$(args.require_arg output-dir)"; then
	log.error "Missing value for --output-dir"
	usage >&2
	exit 64
fi
if [ -z "$output_dir" ]; then
	log.error "Missing required option: --output-dir"
	usage >&2
	exit 64
fi

mkdir -p "$output_dir"
output_dir="$(realpath "$output_dir")"

write_output() {
	local name="$1"
	shift

	{
		printf '$'
		printf ' %q' "$@"
		printf '\n\n'
		"$@"
	} >"$output_dir/$name" 2>&1 || true
}

write_root_output() {
	local name="$1"
	shift

	{
		printf '$ sudo'
		printf ' %q' "$@"
		printf '\n\n'
		sudo "$@"
	} >"$output_dir/$name" 2>&1 || true
}

copy_if_present() {
	local source_path="$1"
	local destination_name="$2"

	if [ -e "$source_path" ]; then
		sudo cp -a "$source_path" "$output_dir/$destination_name" >/dev/null 2>&1 || true
	fi
}

log.pushTask "Collecting package diagnostics into $output_dir"

write_output environment.txt env
write_output uname.txt uname -a
write_output disk-usage.txt df -h
write_output process-list.txt ps aux
write_output lite-nas-package-query.txt dpkg-query -W -f="\${Package}\t\${Status}\t\${Version}\n" lite-nas postfix nats-server nginx zfsutils-linux sudo aide
write_root_output lite-nas-dpkg-status.txt dpkg -s lite-nas
write_root_output lite-nas-package-files.txt dpkg -L lite-nas
write_root_output dpkg-audit.txt dpkg --audit
write_root_output apt-policy.txt apt-cache policy lite-nas postfix nats-server nginx zfsutils-linux sudo aide
write_root_output nats-config-listing.txt ls -lR /etc/nats-server
write_root_output lite-nas-etc-listing.txt ls -lR /etc/lite-nas
write_root_output lite-nas-libexec-listing.txt ls -lR /usr/libexec/lite-nas
write_root_output lite-nas-systemd-listing.txt ls -lR /etc/systemd/system

copy_if_present /var/log/dpkg.log dpkg.log
copy_if_present /var/log/apt/history.log apt-history.log
copy_if_present /var/log/apt/term.log apt-term.log
copy_if_present /var/lib/dpkg/info/lite-nas.postinst lite-nas.postinst
copy_if_present /var/lib/dpkg/info/lite-nas.config lite-nas.config

for unit in \
	nats-server \
	nginx \
	postfix \
	lite-nas-auth \
	lite-nas-rbac \
	lite-nas-system-logging-manager \
	lite-nas-security-logging-manager \
	lite-nas-system-email-notifier \
	lite-nas-security-email-notifier \
	lite-nas-system-metrics \
	lite-nas-zfs-metrics \
	lite-nas-web-gateway \
	lite-nas-resources-monitor; do
	write_root_output "systemctl-status-${unit}.txt" systemctl status "$unit"
	write_root_output "journalctl-${unit}.txt" journalctl -u "$unit" --no-pager -n 400
done

log.popTask
log.info "Collected package diagnostics: $output_dir"
