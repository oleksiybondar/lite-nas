#!/usr/bin/env bash
set -euo pipefail

# Run after package installation to verify the privileged workers are actually
# executing under the profiles and Unix accounts shipped by the package.
for manager in services-manager session-manager usb-manager; do
	user="lite-nas-$manager"
	unit="$user.service"
	profile="/usr/libexec/lite-nas/$manager"
	main_pid="$(systemctl show --property=MainPID --value "$unit")"
	if [[ ! "$main_pid" =~ ^[1-9][0-9]*$ ]]; then
		echo "Invalid MainPID for $unit: $main_pid" >&2
		exit 1
	fi
	if [ "$(systemctl show --property=User --value "$unit")" != "$user" ]; then
		echo "Wrong service user for $unit" >&2
		exit 1
	fi
	if [ "$(cat "/proc/$main_pid/attr/current")" != "$profile (enforce)" ]; then
		echo "AppArmor is not enforcing $profile" >&2
		exit 1
	fi
	visudo -c -f "/etc/sudoers.d/$user" >/dev/null
done
