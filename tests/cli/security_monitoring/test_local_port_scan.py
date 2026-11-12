"""CLI system test for local blocked TCP port scan detection."""

import re
import time
from ipaddress import IPv4Address

import pytest
from constants import SECURITY_LOGGING_MANAGER_CLI_BINARY
from fixtures.logging_manager_cli import (
    as_event_items,
    execute_logging_manager_json,
    wait_for_logging_manager_event,
)
from hyperiontf import CLIClient, expect

pytestmark = [
    pytest.mark.cli,
    pytest.mark.LogAnalysis,
    pytest.mark.NetworkSecurity,
    pytest.mark.SecurityLoggingManager,
]

last_scan_finished_at: list[float] = []


def must_run_local_port_scan(cli_client: CLIClient, mode: str) -> str:
    """Run a fixed root-owned probe pattern and return its source address."""
    if last_scan_finished_at:
        # UFW low logging limits the entire host's burst, including earlier cases.
        time.sleep(max(0, 240 - (time.monotonic() - last_scan_finished_at[0])))
    cli_client.execute(
        f"sudo -n -- /usr/local/libexec/lite-nas-test-port-scan {mode}",
        timeout=180,
    )
    cli_client.assert_exit_code(0)
    last_scan_finished_at[:] = [time.monotonic()]
    match = re.search(r"\bsource_ip=(198\.18\.[0-9]{1,3}\.[0-9]{1,3})\b", cli_client.output)
    if match is None:
        raise AssertionError
    return str(IPv4Address(match.group(1)))


@pytest.mark.parametrize(
    ("mode", "category", "message"),
    [
        ("--ports", "port_scan", "Blocked TCP port scan from"),
        ("--horizontal", "horizontal_scan", "Blocked TCP horizontal scan from"),
        ("--repeated", "repeated_attempts", "Repeated blocked TCP attempts from"),
    ],
)
def test_local_blocked_probe_pattern_creates_security_alert(
    security_cli_client: CLIClient,
    testsudo_cli_client: CLIClient,
    mode: str,
    category: str,
    message: str,
) -> None:
    """Test case: a blocked TCP probe pattern creates its security alert.

    Preparation:
    - Log Analysis, Network Security, UFW, and Security Logging Manager are
      running on a dedicated test host.
    - The restricted sudo test user can run the root-owned scan helper, which
      creates a temporary local network namespace and veth pair.

    Action:
    - Send the selected blocked-probe pattern from the temporary namespace.

    Expected result:
    - A new active security alert has the matching category and a `secnet` ID.

    Requirement: network-security/FR-001, network-security/FR-002,
    network-security/FR-004.
    """
    previous = as_event_items(
        execute_logging_manager_json(
            security_cli_client,
            SECURITY_LOGGING_MANAGER_CLI_BINARY,
            "--cmd getAlerts --page 1 --pageSize 5",
            quiet=True,
        )
    )
    latest_rec_id = max((int(str(event.get("RecID", 0))) for event in previous), default=0)

    source_ip = must_run_local_port_scan(testsudo_cli_client, mode)

    event = wait_for_logging_manager_event(
        security_cli_client,
        SECURITY_LOGGING_MANAGER_CLI_BINARY,
        category=category,
        source="network-security",
        text=f"{message} {source_ip}",
        status="active",
        after_rec_id=latest_rec_id,
    )
    expect(str((event or {}).get("EventID", ""))).to_match(r"^secnet_[0-9]{8}$")
