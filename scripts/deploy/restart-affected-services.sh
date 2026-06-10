#!/usr/bin/env bash

affected_services=(
	apparmor
	nats-server
	postfix
	nginx
)
systemd_reloaded=0

deploy.reloadSystemdOnce() {
	if [ "$systemd_reloaded" -eq 1 ]; then
		return
	fi

	if ! deploy.hasUsableSystemd; then
		log.warn "Systemd is not available in this environment; skipping dependency service reload."
		systemd_reloaded=1
		return
	fi

	log.pushTask "Reloading systemd manager configuration"
	log.popTask
	systemd_reloaded=1
}

deploy.restartService() {
	local service="$1"

	if deploy.hasUsableSystemd; then
		deploy.reloadSystemdOnce
		systemctl restart "$service"
		return
	fi

	if deploy.hasServiceCommand; then
		if service "$service" restart >/dev/null 2>&1; then
			return
		fi

		log.warn "Service command is present but cannot restart $service in this environment; skipping."
		return
	fi

	log.warn "No usable service manager is available for $service; skipping restart."
}

deploy.enableService() {
	local service="$1"

	if deploy.hasUsableSystemd; then
		deploy.reloadSystemdOnce
		systemctl enable "$service"
		return
	fi

	if deploy.hasServiceCommand; then
		return
	fi

	log.warn "No usable service manager is available for $service; skipping enable."
}

deploy.restartAffectedServices() {
	local service

	log.pushTask "Enabling and restarting affected services"
	for service in "${affected_services[@]}"; do
		deploy.enableService "$service"
		deploy.restartService "$service"
	done
	log.popTask
}
