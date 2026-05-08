#!/usr/bin/env bash

affected_services=(
	nats-server
	nginx
)
systemd_reloaded=0

deploy.reloadSystemdOnce() {
	if [ "$systemd_reloaded" -eq 1 ]; then
		return
	fi

	log.pushTask "Reloading systemd manager configuration"
	systemctl daemon-reload
	log.popTask
	systemd_reloaded=1
}

deploy.restartService() {
	local service="$1"

	if command -v systemctl >/dev/null 2>&1; then
		deploy.reloadSystemdOnce
		systemctl restart "$service"
		return
	fi

	if command -v service >/dev/null 2>&1; then
		service "$service" restart
		return
	fi

	log.error "Missing service manager; cannot restart $service."
	exit 1
}

deploy.enableService() {
	local service="$1"

	if command -v systemctl >/dev/null 2>&1; then
		deploy.reloadSystemdOnce
		systemctl enable "$service"
		return
	fi

	if command -v service >/dev/null 2>&1; then
		return
	fi

	log.error "Missing service manager; cannot enable $service."
	exit 1
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
