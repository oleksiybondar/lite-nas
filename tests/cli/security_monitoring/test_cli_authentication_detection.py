"""CLI system tests for authentication detections from PAM audit logs."""

import re
import time
import uuid

import pytest
from constants import (
    SECURITY_LOGGING_MANAGER_CLI_BINARY,
    TEST_SUDO_LOGIN,
)
from fixtures.logging_manager_cli import as_event_items, execute_logging_manager_json
from hyperiontf import CLIClient

pytestmark = [
    pytest.mark.cli,
    pytest.mark.Auth,
    pytest.mark.LogAnalysis,
    pytest.mark.SecurityLoggingManager,
]

EVENT_LIST_COMMAND = "--cmd getAlerts --page 1 --pageSize 25"
SSH_LOGIN_PATTERN = re.compile(r"sec-ssh-[0-9a-f]{12}")
SYSTEM_LOGIN_PATTERN = re.compile(r"[a-z_][a-z0-9_-]{0,31}")
ALERT_WAIT_SECONDS = 25.0


def current_security_alerts(security_cli_client: CLIClient) -> list[dict[str, object]]:
    """Read recent security manager events through the authenticated CLI."""
    payload = execute_logging_manager_json(
        security_cli_client,
        SECURITY_LOGGING_MANAGER_CLI_BINARY,
        EVENT_LIST_COMMAND,
    )
    return as_event_items(payload)


def submit_failed_su_password(cli_client: CLIClient, login: str) -> None:
    """Attempt one rejected `su -` login for a validated system account."""
    if not SYSTEM_LOGIN_PATTERN.fullmatch(login):
        raise ValueError
    cli_client.exec_interactive(f"su - {login}")
    cli_client.wait("Password:", timeout=5)
    cli_client.send_keys("invalid-system-test-password")


def wait_for_cli_security_alert(
    security_cli_client: CLIClient,
    *,
    after_rec_id: int,
    category: str,
    channel: str,
    status: str | None = None,
) -> dict[str, object] | None:
    """Poll for a new security alert matching its category, origin, and state."""
    deadline = time.monotonic() + ALERT_WAIT_SECONDS
    while time.monotonic() < deadline:
        for event in current_security_alerts(security_cli_client):
            if (
                event.get("Source") == "log-analysis"
                and event.get("Category") == category
                and int(event.get("RecID", 0)) > after_rec_id
                and f"via {channel}" in str(event.get("Message", ""))
                and (status is None or event.get("Status") == status)
            ):
                return event
        time.sleep(0.5)
    return None


def test_failed_su_password_is_reported_as_local_cli_authentication(
    security_cli_client: CLIClient,
) -> None:
    """Test case: a failed local account switch creates a channel-labeled alert.

    Preparation:
    - Log Analysis and Security Logging Manager are running, and the security
      test account can read security events through the manager CLI.

    Action:
    - Attempt one `su -` login to the configured sudo test account with an
      invalid password.

    Expected result:
    - A new `auth_failure` alert identifies its origin as `local CLI`.

    Requirement: task.md/SVC-001, logging-managers/FR-001.
    """
    previous_alerts = current_security_alerts(security_cli_client)
    latest_rec_id = max(
        (int(event.get("RecID", 0)) for event in previous_alerts),
        default=0,
    )

    submit_failed_su_password(security_cli_client, TEST_SUDO_LOGIN)

    event = wait_for_cli_security_alert(
        security_cli_client,
        after_rec_id=latest_rec_id,
        category="auth_failure",
        channel="local CLI",
        status="normal",
    )
    assert (
        event is not None
    ), "Expected a new local-CLI auth_failure alert after the rejected su login."


def test_failed_ssh_password_is_reported_as_ssh_cli_authentication(
    security_cli_client: CLIClient,
) -> None:
    """Test case: a rejected SSH password creates an SSH/CLI-labeled alert.

    Preparation:
    - The local SSH server accepts password authentication, and the security
      test account can read events through the security manager CLI.

    Action:
    - Attempt one SSH password login for a unique nonexistent account.

    Expected result:
    - A new `auth_failure` alert identifies its origin as `SSH/CLI`.

    Requirement: task.md/SVC-001, logging-managers/FR-001.
    """
    login = f"sec-ssh-{uuid.uuid4().hex[:12]}"
    if not SSH_LOGIN_PATTERN.fullmatch(login):
        raise ValueError

    previous_alerts = current_security_alerts(security_cli_client)
    latest_rec_id = max(
        (int(event.get("RecID", 0)) for event in previous_alerts),
        default=0,
    )

    security_cli_client.exec_interactive(
        "/usr/bin/ssh -F /dev/null -o PreferredAuthentications=password "
        "-o PubkeyAuthentication=no -o NumberOfPasswordPrompts=1 "
        "-o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no "
        f"-o ConnectTimeout=5 {login}@localhost"
    )
    security_cli_client.wait("'s password:", timeout=10)
    security_cli_client.send_keys("invalid-system-test-password")

    event = wait_for_cli_security_alert(
        security_cli_client,
        after_rec_id=latest_rec_id,
        category="auth_failure",
        channel="SSH/CLI",
        status="normal",
    )
    assert (
        event is not None
    ), "Expected a new SSH/CLI auth_failure alert after the rejected SSH login."


def test_repeated_failed_su_passwords_raise_local_cli_password_guessing_alert(
    security_cli_client: CLIClient,
) -> None:
    """Test case: repeated failed local account switches trigger password guessing.

    Preparation:
    - The account keyed PAM password-guessing rule is enabled with a threshold
      of three failures in two minutes.
    - Log Analysis and Security Logging Manager are running.

    Action:
    - Attempt three `su -` logins to the configured sudo test account, entering
      an invalid password each time.

    Expected result:
    - A new active `password_guessing` alert identifies its origin as `local CLI`.

    Requirement: task.md/SVC-001, logging-managers/FR-004.
    """
    previous_alerts = current_security_alerts(security_cli_client)
    latest_rec_id = max(
        (int(event.get("RecID", 0)) for event in previous_alerts),
        default=0,
    )

    for _ in range(3):
        submit_failed_su_password(security_cli_client, TEST_SUDO_LOGIN)

    event = wait_for_cli_security_alert(
        security_cli_client,
        after_rec_id=latest_rec_id,
        category="password_guessing",
        channel="local CLI",
        status="active",
    )
    assert (
        event is not None
    ), "Expected an active local-CLI password_guessing alert after three failed su logins."
