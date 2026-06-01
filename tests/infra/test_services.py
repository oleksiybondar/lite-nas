"""System infrastructure test suite for installed LiteNAS systemd services."""

import pytest
from constants import SYSTEMD_SERVICES
from hyperiontf import CLIClient

SERVICE_DOMAIN_MARKS = {
    "lite-nas-auth": pytest.mark.Auth,
    "lite-nas-web-gateway": pytest.mark.WebGateway,
    "lite-nas-system-metrics": pytest.mark.SystemMetrics,
    "lite-nas-system-logging-manager": pytest.mark.SystemLoggingManager,
    "lite-nas-security-logging-manager": pytest.mark.SecurityLoggingManager,
    "lite-nas-system-email-notifier": pytest.mark.SystemEmailNotifier,
    "lite-nas-security-email-notifier": pytest.mark.SecurityEmailNotifier,
    "postfix": pytest.mark.Postfix,
    "apparmor": pytest.mark.AppArmor,
    "nginx": pytest.mark.Nginx,
    "nats-server": pytest.mark.NATS,
}

SERVICE_CASES = [
    pytest.param(service, marks=SERVICE_DOMAIN_MARKS[service], id=service)
    for service in SYSTEMD_SERVICES
]


@pytest.mark.infra
@pytest.mark.parametrize("service", SERVICE_CASES)
def test_service_is_active(cli_client: CLIClient, service: str) -> None:
    """Test case: installed LiteNAS service is active.

    Preparation:
    - The LiteNAS package has installed and started the parametrized systemd service.

    Action:
    - Query systemd for the service active state.

    Expected result:
    - The service reports `active`.
    """
    cli_client.execute(f"systemctl is-active {service}")
    cli_client.assert_output_contains("active")


@pytest.mark.infra
@pytest.mark.parametrize("service", SERVICE_CASES)
def test_service_is_enabled(cli_client: CLIClient, service: str) -> None:
    """Test case: installed LiteNAS service is enabled for startup.

    Preparation:
    - The LiteNAS package has installed the parametrized systemd service unit.

    Action:
    - Query systemd for the service enablement state.

    Expected result:
    - The service reports `enabled`.
    """
    cli_client.execute(f"systemctl is-enabled {service}")
    cli_client.assert_output_contains("enabled")


@pytest.mark.infra
@pytest.mark.parametrize("service", SERVICE_CASES)
def test_service_has_no_failed_state(cli_client: CLIClient, service: str) -> None:
    """Test case: installed LiteNAS service is not in a failed systemd state.

    Preparation:
    - The LiteNAS package has installed and started the parametrized systemd service.

    Action:
    - Query systemd for the service failed-state classification.

    Expected result:
    - The service is classified as active rather than failed.
    """
    cli_client.execute(f"systemctl is-failed {service} || true")
    cli_client.assert_output_contains("active")
