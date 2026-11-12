"""Installed Security Automation CLI and host-effect test support."""

import json
import os
import re
import shlex
import time
from collections.abc import Generator
from pathlib import PurePosixPath
from typing import Any

import pytest
from constants import (
    SECURITY_AUTOMATION_CLI_BINARY,
    SECURITY_LOGGING_MANAGER_LOGIN,
    SECURITY_LOGGING_MANAGER_PASSWORD,
)
from fixtures.api_client import RedactTestPassword
from hyperiontf import CLIClient

TEST_SOURCE_IP = "198.18.254.250"


def ufw_has_source_block(client: CLIClient, address: str) -> bool:
    """Identify a direct HTTPS DENY on a named interface for one test IP."""
    if not re.fullmatch(r"198\.18\.[0-9]{1,3}\.[0-9]{1,3}", address):
        raise ValueError
    client.execute("sudo -n /usr/sbin/ufw status", timeout=15)
    client.assert_exit_code(0)
    return any(
        "443/tcp on " in line
        and "DENY" in line
        and re.search(rf"(?<![0-9.]){re.escape(address)}(?![0-9.])", line)
        for line in client.output.splitlines()
    )


def wait_for_ufw_source(client: CLIClient, address: str, blocked: bool) -> bool:
    """Poll the real firewall until the requested test-source state is visible."""
    deadline = time.monotonic() + 30
    while time.monotonic() < deadline:
        if ufw_has_source_block(client, address) == blocked:
            return True
        time.sleep(0.5)
    return False


def automation_command(*args: str) -> str:
    """Quote the installed CLI and validated test-owned operands for the shell."""
    binary = PurePosixPath(SECURITY_AUTOMATION_CLI_BINARY)
    if not binary.is_absolute() or not re.fullmatch(r"/[A-Za-z0-9_./-]+", str(binary)):
        raise ValueError
    return shlex.join([str(binary), *args])


def run_automation_cli(
    client: CLIClient,
    login: str,
    password: str,
    *args: str,
    expect_success: bool = True,
) -> dict[str, Any] | None:
    """Authenticate for one command and return its JSON result when successful."""
    redactor = RedactTestPassword(password)
    client.logger.addFilter(redactor)
    try:
        client.exec_interactive(automation_command(*args))
        client.wait("Username:", timeout=10)
        client.send_keys(login)
        client.wait("Password:", timeout=10)
        client.send_keys(password)
        client.wait(timeout=30)
        if not expect_success:
            return None
        client.assert_exit_code(0)
        lines = client.output.splitlines()
        for line in reversed(lines):
            try:
                result = json.loads(line)
            except json.JSONDecodeError:
                continue
            if isinstance(result, dict):
                return result
        pytest.fail("automation CLI returned no JSON object")
    finally:
        client.logger.removeFilter(redactor)


@pytest.fixture
def enrolled_automation() -> None:
    """Require an explicitly enrolled effect-capable dedicated test installation."""
    if os.environ.get("LITENAS_AUTOMATION_ENROLLMENT") != "1":
        pytest.skip("set LITENAS_AUTOMATION_ENROLLMENT=1 on an enrolled dedicated test host")


@pytest.fixture
def automation_security_identity() -> tuple[str, str]:
    """Supply the installed security operator's interactive CLI credentials."""
    return SECURITY_LOGGING_MANAGER_LOGIN, SECURITY_LOGGING_MANAGER_PASSWORD


@pytest.fixture
def source_ip_cleanup(
    request: pytest.FixtureRequest,
    cli_client: CLIClient,
    automation_security_identity: tuple[str, str],
) -> Generator[str, None, None]:
    """Restore only the test source's automation incident after effect checks."""
    request.getfixturevalue("enrolled_automation")
    yield TEST_SOURCE_IP
    login, password = automation_security_identity
    active = run_automation_cli(cli_client, login, password, "active")
    for item in (active or {}).get("items", []):
        if isinstance(item, dict) and item.get("target") == TEST_SOURCE_IP:
            incident_id = item.get("incident_id")
            if isinstance(incident_id, str) and re.fullmatch(r"[A-Za-z0-9]{1,64}", incident_id):
                run_automation_cli(cli_client, login, password, "force-normalize", incident_id)
                if not wait_for_ufw_source(cli_client, TEST_SOURCE_IP, False):
                    pytest.fail("automation test source remained blocked after cleanup")
