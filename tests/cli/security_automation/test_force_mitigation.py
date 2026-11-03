"""Installed CLI tests for owned UFW automation effects."""

import pytest
from constants import SECURITY_LOGGING_MANAGER_LOGIN, SECURITY_LOGGING_MANAGER_PASSWORD
from fixtures.security_automation import run_automation_cli, wait_for_ufw_source
from hyperiontf import CLIClient, expect

pytestmark = [
    pytest.mark.cli,
    pytest.mark.SecurityAutomation,
    pytest.mark.NetworkSecurity,
]


def test_force_mitigate_blocks_source_in_ufw(
    cli_client: CLIClient,
    source_ip_cleanup: str,
) -> None:
    """Test case: CLI force mitigation applies the source block in UFW.

    Preparation:
    - A dedicated host has enrolled `network_reconnaissance` and permits
      the security operator to use the automation CLI.
    - The benchmark-network source IP is reserved for this test.

    Action:
    - Force the rule with its named requester IP operand and confirmation.

    Expected result:
    - UFW contains a DENY rule for that exact source IP.

    Requirement: security-automation/FR-005, security-automation/FR-009.
    """
    run_automation_cli(
        cli_client,
        SECURITY_LOGGING_MANAGER_LOGIN,
        SECURITY_LOGGING_MANAGER_PASSWORD,
        "force-mitigate",
        "network_reconnaissance",
        f"requester_ip_address={source_ip_cleanup}",
        "--yes",
    )
    expect(wait_for_ufw_source(cli_client, source_ip_cleanup, True)).to_be(True)


def test_force_normalize_removes_only_owned_source_block(
    cli_client: CLIClient,
    source_ip_cleanup: str,
) -> None:
    """Test case: CLI force normalize removes the rule's verified UFW effect.

    Preparation:
    - A dedicated host has enrolled the source-block rule and the fixture
      owns the benchmark-network address.

    Action:
    - Force mitigation, find its active incident, and force normalize it.

    Expected result:
    - UFW no longer contains a DENY rule for that source address.

    Requirement: security-automation/FR-006, security-automation/FR-009.
    """
    login, password = SECURITY_LOGGING_MANAGER_LOGIN, SECURITY_LOGGING_MANAGER_PASSWORD
    run_automation_cli(
        cli_client,
        login,
        password,
        "force-mitigate",
        "network_reconnaissance",
        f"requester_ip_address={source_ip_cleanup}",
        "--yes",
    )
    active = run_automation_cli(cli_client, login, password, "active")
    matching = [
        item
        for item in (active or {}).get("items", [])
        if isinstance(item, dict) and item.get("target") == source_ip_cleanup
    ]
    if len(matching) != 1:
        pytest.fail(f"expected one owned incident for {source_ip_cleanup}, got {len(matching)}")
    incident_id = str(matching[0]["incident_id"])
    run_automation_cli(cli_client, login, password, "force-normalize", incident_id)
    expect(wait_for_ufw_source(cli_client, source_ip_cleanup, False)).to_be(True)
