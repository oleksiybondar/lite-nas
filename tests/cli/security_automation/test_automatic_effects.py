"""Installed automatic Security Automation effect checks on an enrolled host."""

import re
import time

import pytest
from constants import (
    SECURITY_LOGGING_MANAGER_LOGIN,
    SECURITY_LOGGING_MANAGER_PASSWORD,
    TEST_SUDO_LOGIN,
    TEST_SUDO_PASSWORD,
)
from fixtures.api_client import RedactTestPassword
from fixtures.security_automation import run_automation_cli, wait_for_ufw_source
from hyperiontf import CLIClient, expect

pytestmark = [pytest.mark.cli, pytest.mark.SecurityAutomation]
SSH_TERMINATION_TIMEOUT = 45


def restore_test_source(client: CLIClient, address: str) -> None:
    """Normalize only an active incident attributed to this test source."""
    active = run_automation_cli(
        client,
        SECURITY_LOGGING_MANAGER_LOGIN,
        SECURITY_LOGGING_MANAGER_PASSWORD,
        "active",
    )
    for item in (active or {}).get("items", []):
        if isinstance(item, dict) and item.get("target") == address:
            incident_id = item.get("incident_id")
            if isinstance(incident_id, str) and re.fullmatch(r"[A-Za-z0-9]{1,64}", incident_id):
                run_automation_cli(
                    client,
                    SECURITY_LOGGING_MANAGER_LOGIN,
                    SECURITY_LOGGING_MANAGER_PASSWORD,
                    "force-normalize",
                    incident_id,
                )


@pytest.mark.NetworkSecurity
@pytest.mark.usefixtures("enrolled_automation")
def test_local_scan_automatically_blocks_source_in_ufw(
    cli_client: CLIClient,
    testsudo_cli_client: CLIClient,
) -> None:
    """Test case: a detected local scan applies the source block in UFW.

    Preparation:
    - A dedicated host has enrolled `network_reconnaissance` with a tested
      single-alert threshold and the root-owned isolated scan helper.

    Action:
    - Scan blocked ports from a temporary network namespace.

    Expected result:
    - UFW contains a DENY rule for the helper's actual source IP.

    Requirement: security-automation/FR-001, security-automation/FR-005.
    """
    testsudo_cli_client.execute(
        "sudo -n -- /usr/local/libexec/lite-nas-test-port-scan --ports", timeout=180
    )
    testsudo_cli_client.assert_exit_code(0)
    match = re.search(
        r"\bsource_ip=(198\.18\.[0-9]{1,3}\.[0-9]{1,3})\b", testsudo_cli_client.output
    )
    if match is None:
        pytest.fail("scan helper returned no benchmark-network source IP")
    address = match.group(1)
    try:
        expect(wait_for_ufw_source(cli_client, address, True)).to_be(True)
    finally:
        restore_test_source(cli_client, address)


@pytest.mark.Auth
@pytest.mark.SSH
@pytest.mark.usefixtures("enrolled_automation")
def test_repeated_ssh_sudo_denial_terminates_attacking_session(
    cli_client: CLIClient,
) -> None:
    """Test case: repeated sudo denials terminate the attributed SSH session.

    Preparation:
    - A dedicated host has enrolled `repeated_sudo_denial` with verified
      session attribution and a one-way loginctl termination profile.
    - Local SSH password login is enabled for the restricted test account.

    Action:
    - Open an SSH session and make three unauthorized sudo attempts in it.

    Expected result:
    - The SSH connection is closed by the automation before its idle timeout.

    Requirement: security-automation/FR-005, security-automation/FR-009.
    """
    if not re.fullmatch(r"[A-Za-z0-9_][A-Za-z0-9_.-]*", TEST_SUDO_LOGIN):
        raise ValueError
    redactor = RedactTestPassword(TEST_SUDO_PASSWORD)
    cli_client.logger.addFilter(redactor)
    try:
        cli_client.exec_interactive(
            "/usr/bin/ssh -tt -F /dev/null -o PreferredAuthentications=password "
            "-o PubkeyAuthentication=no -o NumberOfPasswordPrompts=1 "
            "-o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no "
            f"-o ConnectTimeout=5 {TEST_SUDO_LOGIN}@localhost "
            "'for i in 1 2 3; do sudo -n /usr/bin/true; sleep 1; done; sleep 90'"
        )
        cli_client.wait("'s password:", timeout=10)
        cli_client.send_keys(TEST_SUDO_PASSWORD)
        started = time.monotonic()
        cli_client.wait(timeout=SSH_TERMINATION_TIMEOUT)
        expect(
            time.monotonic() - started < SSH_TERMINATION_TIMEOUT and cli_client.exit_code != 0
        ).to_be(True)
    finally:
        cli_client.logger.removeFilter(redactor)
