"""System CLI test suite for notifier delivery through Postfix capture mode."""

import json
import os
import shlex
import time
from dataclasses import dataclass

import pytest
from constants import (
    SECURITY_LOGGING_MANAGER_CLI_BINARY,
    SYSTEM_LOGGING_MANAGER_CLI_BINARY,
)
from hyperiontf import CLIClient, expect
from test_data.logging_manager_cli import create_event_payload

POSTFIX_CAPTURE_FILE = os.environ.get(
    "LITE_NAS_POSTFIX_CAPTURE_FILE",
    "/var/tmp/lite-nas-postfix-test-mail.log",  # noqa: S108
)


@dataclass(frozen=True)
class NotifierCaptureContext:
    """Provide per-notifier CLI fixture and event-id prefix details."""

    fixture_name: str
    cli_binary: str
    event_id_prefix: str


NOTIFIER_CAPTURE_CASES = [
    pytest.param(
        NotifierCaptureContext(
            fixture_name="operator_cli_client",
            cli_binary=SYSTEM_LOGGING_MANAGER_CLI_BINARY,
            event_id_prefix="sysmail",
        ),
        marks=[pytest.mark.SystemLoggingManager, pytest.mark.SystemEmailNotifier],
        id="system-notifier-via-system-logging-manager",
    ),
    pytest.param(
        NotifierCaptureContext(
            fixture_name="security_cli_client",
            cli_binary=SECURITY_LOGGING_MANAGER_CLI_BINARY,
            event_id_prefix="secmail",
        ),
        marks=[pytest.mark.SecurityLoggingManager, pytest.mark.SecurityEmailNotifier],
        id="security-notifier-via-security-logging-manager",
    ),
]


def build_event_id(prefix: str) -> str:
    """Build a logging-manager-valid event ID with suite prefix and timestamp suffix."""
    suffix = int(time.time_ns() % 100_000_000)
    return f"{prefix}_{suffix:08d}"


def create_event_with_cli(cli_client: CLIClient, cli_binary: str, event_id: str) -> None:
    """Submit one alert event through the selected logging-manager CLI."""
    payload = json.dumps(create_event_payload(event_id), separators=(",", ":"))
    cli_client.execute(
        f"{cli_binary} --cmd createEvent --data {shlex.quote(payload)}",
        timeout=30,
    )
    cli_client.assert_exit_code(0)


def capture_file_contains_event_id(
    cli_client: CLIClient,
    event_id: str,
    *,
    timeout_seconds: int = 30,
) -> bool:
    """Poll the Postfix capture file until the expected event ID appears."""
    deadline = time.time() + timeout_seconds
    found = False
    while time.time() < deadline:
        cli_client.execute(
            "if grep -F -- "
            f"{shlex.quote(event_id)} {shlex.quote(POSTFIX_CAPTURE_FILE)} "
            ">/dev/null 2>&1; "
            "then echo '__FOUND__'; else echo '__MISSING__'; fi"
        )
        cli_client.assert_exit_code(0)
        if "__FOUND__" in cli_client.output:
            found = True
            break
        time.sleep(1)
    return found


@pytest.mark.cli
@pytest.mark.parametrize("notifier_capture_context", NOTIFIER_CAPTURE_CASES)
def test_email_notifier_writes_alert_to_postfix_capture_file(
    request: pytest.FixtureRequest,
    notifier_capture_context: NotifierCaptureContext,
) -> None:
    """Test case: notifier delivery writes the created event ID into Postfix capture output.

    Preparation:
    - Postfix test capture mode is enabled and writes mail to
      `/var/tmp/lite-nas-postfix-test-mail.log`.
    - The parametrized logging-manager CLI and role user are available.

    Action:
    - Create one unique alert event via the parametrized CLI.
    - Poll the Postfix capture file for the created event ID.

    Expected result:
    - The capture file contains the event ID, proving notifier mail emission was routed.
    """
    context = notifier_capture_context
    cli_client = request.getfixturevalue(context.fixture_name)
    event_id = build_event_id(context.event_id_prefix)
    create_event_with_cli(cli_client, context.cli_binary, event_id)
    found = capture_file_contains_event_id(cli_client, event_id)
    expect(found).to_be(True)
