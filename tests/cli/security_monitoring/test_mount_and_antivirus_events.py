"""System CLI tests for mount lifecycle and antivirus security events."""

import os
import time

import pytest
from constants import (
    SECURITY_LOGGING_MANAGER_CLI_BINARY,
    SYSTEM_LOGGING_MANAGER_CLI_BINARY,
)
from fixtures.logging_manager_cli import as_event_items, execute_logging_manager_json
from fixtures.loopback_mount import LoopbackMount, RamfsMount
from hyperiontf import CLIClient

pytestmark = [pytest.mark.cli]
EVENT_LIST_COMMAND = "--cmd getAlerts --page 1 --pageSize 500"
EVENT_WAIT_SECONDS = 90.0
EVENT_QUERY_BY_SOURCE = {
    "mount-monitor": (SYSTEM_LOGGING_MANAGER_CLI_BINARY, "Mount lifecycle", "active"),
    "antivirus": (SECURITY_LOGGING_MANAGER_CLI_BINARY, "Malware detection", "normal"),
}


def wait_for_event(
    cli_client: CLIClient,
    *,
    source: str,
    text: str,
) -> dict[str, object] | None:
    """Poll one manager CLI until a matching event reaches the requested state."""
    cli_binary, category, status = EVENT_QUERY_BY_SOURCE[source]
    deadline = time.monotonic() + EVENT_WAIT_SECONDS
    while time.monotonic() < deadline:
        payload = execute_logging_manager_json(cli_client, cli_binary, EVENT_LIST_COMMAND)
        for event in as_event_items(payload):
            if (
                event.get("Category") == category
                and event.get("Source") == source
                and text in str(event.get("Message", ""))
                and event.get("Status") == status
            ):
                return event
        time.sleep(1.0)
    return None


@pytest.mark.MountMonitor
@pytest.mark.SystemLoggingManager
@pytest.mark.skipif(os.geteuid() != 0, reason="mount system tests require root privileges")
def test_mount_event_is_emitted_for_a_new_mount(
    loopback_mount: LoopbackMount,
    operator_cli_client: CLIClient,
) -> None:
    """Test case: a new filesystem mount creates an active mount event.

    Preparation:
    - Mount Monitor and System Logging Manager are running.
    - A unique ext4 loopback image is mounted below `/mnt`.

    Action:
    - Query the System Logging Manager CLI for the event for that mount.

    Expected result:
    - An active `Mount lifecycle` event identifies the mounted filesystem.

    Cleanup:
    - The mount fixture unmounts the loopback image during teardown.

    Requirement: task.md/SVC-002, logging-managers/FR-003.
    """
    event = wait_for_event(
        operator_cli_client,
        source="mount-monitor",
        text=str(loopback_mount.mount_point),
    )
    assert event is not None, f"No active mount event for {loopback_mount.mount_point}."


@pytest.mark.MountMonitor
@pytest.mark.SystemLoggingManager
@pytest.mark.skipif(os.geteuid() != 0, reason="mount system tests require root privileges")
def test_ramfs_mount_emits_informational_event_with_filesystem_type(
    ramfs_mount: RamfsMount,
    operator_cli_client: CLIClient,
) -> None:
    """Test case: mounting ramfs produces an informational event naming its filesystem.

    Preparation:
    - Mount Monitor and System Logging Manager are running.
    - A unique `ramfs` filesystem is mounted under `/mnt`.

    Action:
    - Query the System Logging Manager CLI for the new mount event.

    Expected result:
    - An informational `Mount lifecycle` event identifies `ramfs` and the
      unique mount point.

    Requirement: task.md/SVC-002, logging-managers/FR-001.
    """
    event = wait_for_event(
        operator_cli_client,
        source="mount-monitor",
        text=f"(ramfs): {ramfs_mount.mount_point}",
    )
    assert event is not None, f"Expected an informational ramfs mount event, got {event!r}."
    assert event.get("Severity") == "info"


@pytest.mark.Antivirus
@pytest.mark.SecurityLoggingManager
@pytest.mark.skipif(os.geteuid() != 0, reason="mount system tests require root privileges")
def test_eicar_mount_creates_one_resolved_antivirus_event(
    infected_loopback_mount: LoopbackMount,
    security_cli_client: CLIClient,
) -> None:
    """Test case: an EICAR signature on a unique mount creates one resolved malware event.

    Preparation:
    - Mount Monitor, Antivirus, and Security Logging Manager are running.
    - A unique ext4 loopback image under `/mnt` contains the EICAR test signature
      in two files before it is mounted, so the scanner sees both on its initial scan.

    Action:
    - Allow Antivirus to scan the newly observed mount and query events through
      the Security Logging Manager CLI.

    Expected result:
    - One `Malware detection` event from Antivirus is normal after registration
      and has no occurrence row for this one scan.

    Requirement: task.md/SVC-003, logging-managers/FR-001.
    """
    event = wait_for_event(
        security_cli_client,
        source="antivirus",
        text=f"{infected_loopback_mount.mount_point} (2 findings)",
    )
    assert (
        event is not None
    ), f"Expected one resolved malware event without occurrences, got {event!r}."
    assert event.get("LastRecID") is None
