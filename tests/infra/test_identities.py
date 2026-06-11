"""System infrastructure test suite for LiteNAS users and groups."""

import pytest
from hyperiontf import CLIClient

NON_INTERACTIVE_USERS = [
    "lite-nas-web-gateway",
    "lite-nas-system-metrics",
    "lite-nas-sys-log-mgr",
    "lite-nas-sec-log-mgr",
]

USER_DOMAIN_MARKS = {
    "lite-nas-web-gateway": pytest.mark.WebGateway,
    "lite-nas-system-metrics": pytest.mark.SystemMetrics,
    "lite-nas-sys-log-mgr": pytest.mark.SystemLoggingManager,
    "lite-nas-sec-log-mgr": pytest.mark.SecurityLoggingManager,
}

REQUIRED_GROUPS = [
    "lite-nas",
    "lite-nas-security",
    "lite-nas-operator",
]

GROUP_DOMAIN_MARKS = {
    "lite-nas": pytest.mark.WebGateway,
    "lite-nas-security": pytest.mark.SecurityLoggingManager,
    "lite-nas-operator": pytest.mark.SystemLoggingManager,
}

USER_CASES = [
    pytest.param(
        user,
        marks=USER_DOMAIN_MARKS[user],
        id=user,
    )
    for user in NON_INTERACTIVE_USERS
]

GROUP_CASES = [
    pytest.param(
        group,
        marks=GROUP_DOMAIN_MARKS[group],
        id=group,
    )
    for group in REQUIRED_GROUPS
]


def assert_system_user_exists_and_is_non_interactive(
    cli_client: CLIClient,
    user: str,
) -> None:
    """Verify the runtime user exists and uses a non-interactive shell."""
    cli_client.execute(f"getent passwd {user}")
    cli_client.assert_output_contains(f"{user}:")
    cli_client.assert_output_contains("/usr/sbin/nologin")


def assert_system_group_exists(cli_client: CLIClient, group: str) -> None:
    """Verify the required LiteNAS group exists in the local group database."""
    cli_client.execute(f"getent group {group}")
    cli_client.assert_output_contains(f"{group}:")


@pytest.mark.infra
@pytest.mark.parametrize("user", USER_CASES)
def test_non_interactive_runtime_user_exists(
    cli_client: CLIClient,
    user: str,
) -> None:
    """Test case: LiteNAS runtime user exists with non-interactive login shell.

    Preparation:
    - The LiteNAS Debian package has been deployed on the target host.
    - Service deployment scripts created the parametrized runtime user.

    Action:
    - Query passwd entries for the parametrized user.

    Expected result:
    - The user entry exists and uses `/usr/sbin/nologin`.
    """
    assert_system_user_exists_and_is_non_interactive(cli_client, user)


@pytest.mark.infra
@pytest.mark.parametrize("group", GROUP_CASES)
def test_required_lite_nas_group_exists(
    cli_client: CLIClient,
    group: str,
) -> None:
    """Test case: LiteNAS bootstrap and role groups exist.

    Preparation:
    - The LiteNAS Debian package has been deployed on the target host.
    - Bootstrap deployment created required shared and role groups.

    Action:
    - Query the local group database for the parametrized group.

    Expected result:
    - The group entry exists.
    """
    assert_system_group_exists(cli_client, group)
