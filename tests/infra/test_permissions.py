"""System infrastructure test suite for deployed LiteNAS path permissions."""

import os

import pytest
from hyperiontf import CLIClient

ETC_PERMISSION_CASES = [
    pytest.param(
        "/etc/lite-nas",
        "root:lite-nas",
        "711",
        id="etc-lite-nas-dir",
    ),
    pytest.param(
        "/etc/lite-nas/auth.conf",
        "root:lite-nas",
        "640",
        id="auth-conf",
    ),
    pytest.param(
        "/etc/lite-nas/web-gateway.conf",
        "root:lite-nas",
        "640",
        id="web-gateway-conf",
    ),
    pytest.param(
        "/etc/lite-nas/system-metrics.conf",
        "root:lite-nas",
        "640",
        id="system-metrics-conf",
    ),
    pytest.param(
        "/etc/lite-nas/system-logging-manager.conf",
        "root:lite-nas",
        "640",
        id="system-logging-manager-conf",
    ),
    pytest.param(
        "/etc/lite-nas/security-logging-manager.conf",
        "root:lite-nas",
        "640",
        id="security-logging-manager-conf",
    ),
    pytest.param(
        "/etc/lite-nas/system-metrics-cli.conf",
        "root:root",
        "644",
        id="system-metrics-cli-conf",
    ),
    pytest.param(
        "/etc/lite-nas/system-logging-manager-cli.conf",
        "root:lite-nas-operator",
        "640",
        id="system-logging-manager-cli-conf",
    ),
    pytest.param(
        "/etc/lite-nas/security-logging-manager-cli.conf",
        "root:lite-nas-security",
        "640",
        id="security-logging-manager-cli-conf",
    ),
]

LOG_PERMISSION_CASES = [
    pytest.param(
        "/var/log/lite-nas",
        "root:lite-nas",
        "751",
        id="log-dir",
    ),
    pytest.param(
        "/var/log/lite-nas/auth-service.log",
        "root:root",
        "640",
        id="auth-log",
    ),
    pytest.param(
        "/var/log/lite-nas/web-gateway.log",
        "lite-nas-web-gateway:lite-nas-web-gateway",
        "640",
        id="web-gateway-log",
    ),
    pytest.param(
        "/var/log/lite-nas/system-metrics.log",
        "lite-nas-system-metrics:lite-nas-system-metrics",
        "640",
        id="system-metrics-log",
    ),
    pytest.param(
        "/var/log/lite-nas/system-logging-manager.log",
        "lite-nas-sys-log-mgr:lite-nas-sys-log-mgr",
        "640",
        id="system-logging-manager-log",
    ),
    pytest.param(
        "/var/log/lite-nas/security-logging-manager.log",
        "lite-nas-sec-log-mgr:lite-nas-sec-log-mgr",
        "640",
        id="security-logging-manager-log",
    ),
    pytest.param(
        "/var/log/lite-nas/system-metrics-cli.log",
        "root:root",
        "666",
        id="system-metrics-cli-log",
    ),
    pytest.param(
        "/var/log/lite-nas/system-logging-manager-cli.log",
        "root:root",
        "666",
        id="system-logging-manager-cli-log",
    ),
    pytest.param(
        "/var/log/lite-nas/security-logging-manager-cli.log",
        "root:root",
        "666",
        id="security-logging-manager-cli-log",
    ),
]

VAR_LIB_PERMISSION_CASES = [
    pytest.param(
        "/var/lib/lite-nas/system-logging-manager",
        "lite-nas-sys-log-mgr:lite-nas-sys-log-mgr",
        "700",
        False,
        id="system-logging-manager-db-dir",
    ),
    pytest.param(
        "/var/lib/lite-nas/system-logging-manager/log.db",
        "lite-nas-sys-log-mgr:lite-nas-sys-log-mgr",
        "600",
        True,
        id="system-logging-manager-db-file",
    ),
    pytest.param(
        "/var/lib/lite-nas/security-logging-manager",
        "lite-nas-sec-log-mgr:lite-nas-sec-log-mgr",
        "700",
        False,
        id="security-logging-manager-db-dir",
    ),
    pytest.param(
        "/var/lib/lite-nas/security-logging-manager/log.db",
        "lite-nas-sec-log-mgr:lite-nas-sec-log-mgr",
        "600",
        True,
        id="security-logging-manager-db-file",
    ),
]


def assert_path_owner_and_mode(
    cli_client: CLIClient,
    path: str,
    expected_owner_group: str,
    expected_mode: str,
) -> None:
    """Verify a deployed path exists and matches expected owner/group and mode."""
    cli_client.execute(
        f"if [ -e '{path}' ]; then "
        f"actual=\"$(stat -c '%U:%G %a' '{path}')\"; "
        f"if [ \"$actual\" = '{expected_owner_group} {expected_mode}' ]; then "
        f"echo '__OK__ {path}'; "
        "else "
        'echo "__BAD__ $actual"; '
        "fi; "
        "else "
        "echo '__MISSING__'; "
        "fi"
    )
    cli_client.assert_output_contains(f"__OK__ {path}")


@pytest.mark.infra
@pytest.mark.parametrize(("path", "owner_group", "mode"), ETC_PERMISSION_CASES)
def test_lite_nas_etc_paths_have_expected_permissions(
    cli_client: CLIClient,
    path: str,
    owner_group: str,
    mode: str,
) -> None:
    """Test case: deployed LiteNAS etc config paths exist with expected permissions.

    Preparation:
    - The LiteNAS Debian package has been deployed on the target host.
    - Deployment scripts installed LiteNAS config files under `/etc/lite-nas`.

    Action:
    - Query owner/group and numeric mode for the parametrized etc path.

    Expected result:
    - The path exists and matches the expected owner/group and mode.
    """
    assert_path_owner_and_mode(cli_client, path, owner_group, mode)


@pytest.mark.infra
@pytest.mark.parametrize(("path", "owner_group", "mode"), LOG_PERMISSION_CASES)
def test_lite_nas_log_paths_have_expected_permissions(
    cli_client: CLIClient,
    path: str,
    owner_group: str,
    mode: str,
) -> None:
    """Test case: deployed LiteNAS log paths exist with expected permissions.

    Preparation:
    - The LiteNAS Debian package has been deployed on the target host.
    - Deployment scripts installed LiteNAS-managed logs under `/var/log/lite-nas`.

    Action:
    - Query owner/group and numeric mode for the parametrized log path.

    Expected result:
    - The path exists and matches the expected owner/group and mode.
    """
    assert_path_owner_and_mode(cli_client, path, owner_group, mode)


@pytest.mark.infra
@pytest.mark.parametrize(
    ("path", "owner_group", "mode", "requires_privileged_metadata"),
    VAR_LIB_PERMISSION_CASES,
)
def test_lite_nas_var_lib_paths_have_expected_permissions(
    cli_client: CLIClient,
    path: str,
    owner_group: str,
    mode: str,
    requires_privileged_metadata: bool,
) -> None:
    """Test case: deployed LiteNAS state paths exist with expected permissions.

    Preparation:
    - The LiteNAS Debian package has been deployed on the target host.
    - Deployment scripts installed LiteNAS-managed state under `/var/lib/lite-nas`.

    Action:
    - Query owner/group and numeric mode for the parametrized state path.

    Expected result:
    - The path exists and matches the expected owner/group and mode.
    """
    if requires_privileged_metadata and os.geteuid() != 0:
        pytest.skip(
            "Metadata verification for strict 0600 state files requires root-level traversal."
        )
    assert_path_owner_and_mode(cli_client, path, owner_group, mode)
