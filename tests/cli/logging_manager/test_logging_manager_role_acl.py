"""CLI system tests for role-based logging-manager access boundaries."""

import pytest
from constants import SYSTEM_LOGGING_MANAGER_CLI_BINARY
from hyperiontf import CLIClient

pytestmark = [
    pytest.mark.cli,
    pytest.mark.SystemLoggingManager,
    pytest.mark.SecurityLoggingManager,
]


def test_security_role_cannot_query_system_logging_manager(
    security_cli_client: CLIClient,
) -> None:
    """Test case: a security-only role cannot read system-manager alerts.

    Preparation:
    - The security test account has the security role but not operator,
      administrator, or sudo roles.
    - The System Logging Manager CLI and service are available.

    Action:
    - Query the System Logging Manager as the security test account.

    Expected result:
    - The CLI denies the cross-domain query for insufficient role permissions.

    Requirement: logging-managers/SR-002.
    """
    security_cli_client.execute(
        f"{SYSTEM_LOGGING_MANAGER_CLI_BINARY} " "--cmd getAlerts --page 1 --pageSize 1 --json"
    )
    output = security_cli_client.output.lower()
    denied_for_role = any(
        reason in output
        for reason in ("permission", "unauthorized", "forbidden", "insufficient role")
    )
    assert security_cli_client.exit_code != 0
    assert denied_for_role, (
        "Expected the system logging-manager CLI to reject the security-only role; "
        f"exit_code={security_cli_client.exit_code}, output={security_cli_client.output!r}."
    )
